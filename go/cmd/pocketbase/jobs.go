package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const maxParallelJobs = 3 // Max de jobs exécutés simultanément
const timeoutSecond = 10 * time.Second

var jobSemaphore = make(chan struct{}, maxParallelJobs) // Sémaphore de contrôle de concurrence

type Interval struct {
	stop chan struct{}
}

// setInterval lance une fonction toutes les `ms` millisecondes
func setInterval(fn func(), ms int) *Interval {
	interval := &Interval{stop: make(chan struct{})}

	ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				fn()
			case <-interval.stop:
				ticker.Stop()
				return
			}
		}
	}()

	return interval
}

// clearInterval arrête l'intervalle
func clearInterval(i *Interval) {
	close(i.stop)
}

// handleJobExecution exécute un job en backend (appelé à la création du record)
func handleJobExecution(app *pocketbase.PocketBase, job *core.Record) {
	logger := app.Logger()

	// 🔐 Mutex pour les accès concurrents
	var (
		mu          sync.Mutex
		isUpdate    bool = false
		isLogUpdate bool = false
		logBuilder  strings.Builder
		lastUpdated time.Time
	)

	now := func() string {
		return time.Now().Format(time.RFC3339)
	}

	set := func(key string, value any) {
		mu.Lock()
		defer mu.Unlock()

		job.Set(key, value)
		isUpdate = true
		lastUpdated = time.Now()
	}

	log := func(level string, args ...any) {
		mu.Lock()
		defer mu.Unlock()

		logBuilder.WriteString(level)

		for _, arg := range args {
			logBuilder.WriteString(" ")
			logBuilder.WriteString(fmt.Sprint(arg))
		}

		logBuilder.WriteString("\n")

		isLogUpdate = true
		isUpdate = true
	}

	sync := func() {
		if !isUpdate {
			return
		}

		log("D", "sync", now())

		mu.Lock()
		defer mu.Unlock()

		if isLogUpdate {
			logs := logBuilder.String()
			logger.Debug("📄 job logs", "id", job.Id, "logs", logs)
			job.Set("logs", logBuilder.String())

			isLogUpdate = false
		}

		// Sauvegarde sécurisée du record avec timestamp
		if err := app.Save(job); err != nil {
			logger.Error("❌ job save failed", "id", job.Id, "err", err)
		}

		isUpdate = false
	}

	// Enregistre à la fin du job
	defer sync()

	// Initialisation du job
	logger.Info("▶️ job started", "id", job.Id)
	set("status", "processing")
	set("progress", 1)
	sync()

	jobBytes, err := job.MarshalJSON()
	if err != nil {
		log("E", "failed to marshal job", err.Error())
		set("status", "failed")
		set("error", "Invalid input JSON")
		return
	}

	jobJson := string(jobBytes)

	action := job.GetString("action")
	cmd := exec.Command("deno", "run", "--allow-all", "jobs/"+action+".ts", jobJson)

	// Mise en place du watchdog de sync et timeout
	timer := setInterval(func() {
		if time.Now() > lastUpdated+timeoutSecond {
			set("status", "failed")
			set("error", "No update within timeout")
			cmd.Process.Kill()
		}

		sync()
	}, 2000)
	defer clearInterval(timer)

	// Préparation des pipes stdout / stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("❌ failed to get stdout", "id", job.Id, "err", err)
		set("status", "failed")
		set("error", "stdout pipe error")
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("❌ failed to get stderr", "id", job.Id, "err", err)
		set("status", "failed")
		set("error", "stderr pipe error")
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Traitement de la sortie standard (stdout)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			parts := strings.SplitN(line, " ", 2)
			prefix, rest := parts[0], parts[1]

			switch prefix {
			case "progress":
				set("progress", int(parseFloat(rest)))
			case "result":
				var result any
				if err := json.Unmarshal([]byte(rest), &result); err != nil {
					log("E", "error unmarshal result", line)
					continue
				}
				logger.Info("job result", "id", job.Id, "result", result)
				set("result", result)
			case "error":
				log("E", rest)
			case "warn":
				log("W", rest)
			case "info":
				log("I", rest)
			case "debug":
				log("D", rest)
			default:
				log("D", line)
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Error("❌ stdout scanner error", err)
			log("E", "stdout scanner error", err.Error())
		}
	}()

	// Traitement de la sortie d'erreur (stderr)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			line := scanner.Text()
			log("E", line)
		}

		if err := scanner.Err(); err != nil {
			logger.Error("❌ stderr scanner error", err)
			log("E", "stderr scanner error", err.Error())
		}
	}()

	// Lancement du processus
	if err := cmd.Start(); err != nil {
		logger.Error("❌ failed to start job", "id", job.Id, "err", err)
		set("status", "failed")
		set("error", err.Error())
		return
	}

	// Attente de la fin du processus
	if err := cmd.Wait(); err != nil {
		logger.Error("❌ job process failed", "id", job.Id, "error", err)
		set("status", "failed")
		set("error", err.Error())
		wg.Wait()
		return
	}

	// Attente de fin des goroutines de lecture
	wg.Wait()

	// Petite pause pour capter les logs restants
	time.Sleep(100 * time.Millisecond)

	// Finalisation
	set("status", "finished")
	set("progress", 100)

	logger.Info("✅ job finished", "id", job.Id)
}

// bindJobs attache le handler sur création de job
func bindJobs(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateSuccess("jobs").BindFunc(func(e *core.RecordEvent) error {
		job := e.Record

		go func() {
			// 🔒 Contrôle de concurrence : limite le nombre de jobs actifs
			app.Logger().Debug("⏳ waiting for available job slot", "id", job.Id)
			jobSemaphore <- struct{}{}
			defer func() { <-jobSemaphore }()
			app.Logger().Debug("🚀 acquired job slot", "id", job.Id)

			handleJobExecution(app, job)
		}()

		return nil
	})
}
