package main

import (
	"bufio"
	"encoding/json"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const maxParallelJobs = 3                               // Max de jobs exécutés simultanément
var jobSemaphore = make(chan struct{}, maxParallelJobs) // Sémaphore de contrôle de concurrence

// handleJobExecution exécute un job en backend (appelé à la création du record)
func handleJobExecution(app *pocketbase.PocketBase, job *core.Record) {
	logger := app.Logger()
	jobID := job.Id
	action := job.GetString("action")

	// 🔒 Contrôle de concurrence : limite le nombre de jobs actifs
	logger.Info("⏳ waiting for available job slot", "id", jobID)
	jobSemaphore <- struct{}{}
	defer func() { <-jobSemaphore }()
	logger.Info("🚀 acquired job slot", "id", jobID)

	// 🔐 Mutex pour les accès concurrents à l'objet job
	var mu sync.Mutex

	// Sauvegarde sécurisée du record avec timestamp
	saveJob := func() {
		mu.Lock()
		defer mu.Unlock()

		if err := app.Save(job); err != nil {
			logger.Error("❌ job save failed", "id", jobID, "err", err)
		}
	}

	// Gestion des logs JSON (slice de chaînes)
	var logs [][]any
	logJob := func(level string, args ...any) {
		timestamp := time.Now().Format(time.RFC3339)

		// Construire la ligne de log comme un slice [time, level, ...args]
		line := make([]any, 0, 2+len(args))
		line = append(line, timestamp, level)
		line = append(line, args...)

		logger.Info("📄 job log", "id", jobID, "level", level, "line", line)

		mu.Lock()
		logs = append(logs, line)
		job.Set("logs", logs)
		mu.Unlock()

		saveJob()
	}

	// Initialisation du job
	logger.Info("▶️ job started", "id", jobID, "action", action)
	job.Set("status", "processing")
	job.Set("progress", 1)
	saveJob()

	// Mise en place du watchdog (timeout 10s si aucun progrès)
	progressSignal := make(chan struct{}, 1)
	jobDone := make(chan struct{}, 1)
	timeout := 10 * time.Second

	jobBytes, err := job.MarshalJSON()
	if err != nil {
		logJob("error", "failed to marshal input", err.Error())
		job.Set("status", "failed")
		job.Set("error", "Invalid input JSON")
		saveJob()
		return
	}

	jobJson := string(jobBytes)

	cmd := exec.Command("deno", "run", "--allow-all", "jobs/"+action+".ts", jobJson)

	// Goroutine de surveillance du progrès
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case <-jobDone:
				return
			case <-progressSignal:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				logJob("ERROR", "timeout")
				job.Set("status", "failed")
				job.Set("error", "No progress update within timeout")
				saveJob()
				_ = cmd.Process.Kill()
				return
			}
		}
	}()

	// Préparation des pipes stdout / stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("❌ failed to get stdout", "id", jobID, "err", err)
		job.Set("status", "failed")
		job.Set("error", "stdout pipe error")
		saveJob()
		close(jobDone)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("❌ failed to get stderr", "id", jobID, "err", err)
		job.Set("status", "failed")
		job.Set("error", "stderr pipe error")
		saveJob()
		close(jobDone)
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

			switch {
			case strings.HasPrefix(line, "progress "):
				progressStr := strings.TrimPrefix(line, "progress ")
				if progress, err := strconv.ParseFloat(progressStr, 64); err == nil {
					job.Set("progress", int(progress))
					saveJob()
					select {
					case progressSignal <- struct{}{}:
					default:
					}
				}

			case strings.HasPrefix(line, "result "):
				resultString := strings.TrimPrefix(line, "result ")
				var result any

				// Tente de parser en JSON
				if err := json.Unmarshal([]byte(resultString), &result); err != nil {
					// Si c'est pas du JSON, on enregistre la chaîne brute
					job.Set("result", resultString)
				} else {
					// Si c'est du JSON valide, on enregistre l'objet décodé
					job.Set("result", result)
				}

				saveJob()

			case strings.HasPrefix(line, "error "):
				errMsg := strings.TrimPrefix(line, "error ")
				job.Set("error", errMsg)
				saveJob()

			default:
				logJob("info", line)
			}
		}
		if err := scanner.Err(); err != nil {
			logJob("error", "stdout scanner error", err.Error())
		}
	}()

	// Traitement de la sortie d'erreur (stderr)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logJob("error", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logJob("error", "stderr scanner error: ", err.Error())
		}
	}()

	// Lancement du processus
	if err := cmd.Start(); err != nil {
		logger.Error("❌ failed to start job", "id", jobID, "err", err)
		job.Set("status", "failed")
		job.Set("error", err.Error())
		saveJob()
		close(jobDone)
		return
	}

	// Attente de la fin du processus
	if err := cmd.Wait(); err != nil {
		logger.Error("❌ job process failed", "id", jobID, "error", err)
		job.Set("status", "failed")
		job.Set("error", err.Error())
		saveJob()
		wg.Wait()
		close(jobDone)
		return
	}

	// Attente de fin des goroutines de lecture
	wg.Wait()
	close(jobDone)

	// Petite pause pour capter les logs restants
	time.Sleep(100 * time.Millisecond)

	// Finalisation
	job.Set("status", "finished")
	job.Set("progress", 100)
	saveJob()

	logger.Info("✅ job finished", "id", jobID)
}

// bindJobs attache le handler sur création de job
func bindJobs(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateSuccess("jobs").BindFunc(func(e *core.RecordEvent) error {
		job := e.Record

		go func() {
			handleJobExecution(app, job)
		}()

		return nil
	})
}
