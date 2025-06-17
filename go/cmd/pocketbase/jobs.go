package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	maxParallelJobs = 3 // Max de jobs exécutés simultanément
	timeoutSecond   = 10 * time.Second
)

var jobSemaphore = make(chan struct{}, maxParallelJobs) // Sémaphore de contrôle de concurrence

type Interval struct {
	stop chan struct{}
}

// startInterval démarre une exécution répétée de fn toutes les ms millisecondes
func startInterval(fn func(), ms int) *Interval {
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

// stopInterval arrête l'exécution répétée
func stopInterval(i *Interval) {
	close(i.stop)
}

// startJob exécute un job en backend (appelé à la création du record)
func startJob(app *pocketbase.PocketBase, job *core.Record) {
	logger := app.Logger()

	// 🔐 Mutex pour les accès concurrents
	var (
		mu          sync.Mutex
		isUpdate    atomic.Bool
		isLogUpdate atomic.Bool
		lastUpdated atomic.Int64
		logs        strings.Builder
	)
	isUpdate.Store(false)
	isLogUpdate.Store(false)

	// Utilitaires thread-safe pour modifier l'état du job
	set := func(key string, val any) {
		mu.Lock()

		job.Set(key, val)

		mu.Unlock()

		isUpdate.Store(true)
		lastUpdated.Store(time.Now().UnixNano())
	}

	log := func(level string, args ...any) {
		mu.Lock()

		logs.WriteString(level)

		for _, arg := range args {
			logs.WriteString(" ")
			logs.WriteString(fmt.Sprint(arg))
		}

		logs.WriteString("\n")

		mu.Unlock()

		isUpdate.Store(true)
		lastUpdated.Store(time.Now().UnixNano())
		isLogUpdate.Store(true)
	}

	// Synchronise les modifications du job avec la base
	flushState := func() {
		if !isUpdate.Load() {
			return
		}
		isUpdate.Store(false)

		if isLogUpdate.Load() {
			set("logs", logs.String())
			isLogUpdate.Store(false)
		}

		log("T", time.Now().Format(time.RFC3339))

		mu.Lock()

		// Sauvegarde sécurisée du record avec timestamp
		if err := app.Save(job); err != nil {
			logger.Error("❌ job flushState failed", "id", job.Id, "err", err)
		}

		mu.Unlock()
	}

	defer flushState()

	// Initialisation du job
	logger.Info("▶️ job started", "id", job.Id)
	set("status", "processing")
	set("progress", 1)
	flushState()

	jobJSON, err := job.MarshalJSON()
	if err != nil {
		log("E", "failed to marshal job", err.Error())
		set("status", "failed")
		set("error", "Invalid input JSON")
		return
	}

	// sécurise le nom du script
	action := filepath.Base(job.GetString("action"))
	script := filepath.Join("jobs", action+".ts")

	// process avec contexte
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "deno", "run", "--allow-all", script, string(jobJSON))

	// Démarre le watchdog qui surveille les updates
	timer := startInterval(func() {
		if time.Now().UnixNano()-lastUpdated.Load() > int64(timeoutSecond) {
			set("status", "failed")
			set("error", "No update within timeout")
			cmd.Process.Kill()
		}

		flushState()
	}, 2000)
	defer stopInterval(timer)

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

			if len(parts) < 2 {
				log("D", line)
				continue
			}

			prefix, rest := parts[0], parts[1]

			switch prefix {
			case "progress":
				if val, err := strconv.ParseFloat(rest, 64); err == nil {
					set("progress", int(math.Round(val)))
				} else {
					log("E", "invalid progress value", rest)
				}
				continue
			case "result":
				var result any
				if err := json.Unmarshal([]byte(rest), &result); err != nil {
					result = rest
				}
				logger.Info("job result", "id", job.Id, "result", result)
				set("result", result)
				continue
			case "error":
				log("E", rest)
				continue
			case "warn":
				log("W", rest)
				continue
			case "info":
				log("I", rest)
				continue
			case "debug":
				log("D", rest)
				continue
			default:
				log("D", line)
				continue
			}
		}

		if err := scanner.Err(); err != nil {
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

func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic: %v\n", r)
			}
		}()
		fn()
	}()
}

func handleJob(app *pocketbase.PocketBase, job *core.Record) {
	app.Logger().Debug("handleJob", "id", job.Id, "status", job.GetString("status"))

	if job.GetString("status") != "" { // déjà traité
		return
	}

	safeGo(func() {
		// 🔒 Contrôle de concurrence : limite le nombre de jobs actifs
		app.Logger().Info("⏳ wait slot", "id", job.Id)
		jobSemaphore <- struct{}{}
		defer func() { <-jobSemaphore }()
		app.Logger().Debug("🚀 slot ok", "id", job.Id)

		startJob(app, job)
	})

	// Attendre un peu pour les jobs rapide (obtenir directement le resultat)
	time.Sleep(2000)
}

// bindJobs attache le handler sur création de job
func bindJobs(app *pocketbase.PocketBase) {
	app.OnRecordCreateExecute("jobs").BindFunc(func(e *core.RecordEvent) error {
		handleJob(app, e.Record)
		return e.Next()
	})

	app.OnRecordUpdateExecute("jobs").BindFunc(func(e *core.RecordEvent) error {
		handleJob(app, e.Record)
		return e.Next()
	})
}
