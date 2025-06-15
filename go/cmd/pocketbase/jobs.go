// medias.go
package main

import (
	"bufio"
	"fmt"
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

const maxConcurrentJobs = 2 // 👈 À adapter selon tes besoins
var jobSlots = make(chan struct{}, maxConcurrentJobs)

func onJobCreate(app core.App, job *core.Record) {
	logger := app.Logger()
	id := job.Id
	action := job.GetString("action")
	input := job.GetString("input")

	// 🔒 Limite le nombre de jobs parallèles
	logger.Info("⏳ waiting for available job slot", "id", id)
	jobSlots <- struct{}{}        // 👈 bloque si tous les slots sont pris
	defer func() { <-jobSlots }() // 👈 libère le slot à la fin
	logger.Info("🚀 acquired job slot", "id", id)

	var mu sync.Mutex

	save := func() {
		mu.Lock()
		defer mu.Unlock()

		job.Set("updated", time.Now().Format(time.RFC3339))

		if err := app.SaveNoValidate(job); err != nil {
			logger.Error("❌ job save failed", "id", id, "err", err)
		}
	}

	var logs string
	addLog := func(level, line string) {
		timestamp := time.Now().Format("[2006-01-02 15:04:05]")

		logger.Info("▶️ job log", "id", id, "level", level, "line", line)

		mu.Lock()
		logs = logs + fmt.Sprintf("%s %s %s\n", timestamp, level, line)
		job.Set("logs", logs)
		mu.Unlock()

		save()
	}

	logger.Info("▶️ job started", "id", id, "action", action, "input", input)
	job.Set("status", "processing")
	job.Set("progress", 1)
	save()

	// Timer déclenché si le progrès ne change pas pendant 10s
	progressUpdated := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	timeoutDuration := 10 * time.Second

	cmd := exec.Command("deno", "run", "--allow-all", "jobs/"+action+".ts", job.Id, input)

	go func() {
		timer := time.NewTimer(timeoutDuration)
		defer timer.Stop()
		for {
			select {
			case <-done:
				return // ✅ arrêt explicite de la goroutine si job fini ou échoué
			case <-progressUpdated:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeoutDuration)
			case <-timer.C:
				addLog("ERROR", "timeout: no progress update")
				job.Set("status", "failed")
				job.Set("error", "no progress update for 10s")
				save()
				cmd.Process.Kill()
				return
			}
		}
	}()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("❌ stdout pipe error", "err", err)
		job.Set("status", "failed")
		job.Set("error", "stdout pipe error")
		save()
		close(done)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("❌ stderr pipe error", "err", err)
		job.Set("status", "failed")
		job.Set("error", "stderr pipe error")
		save()
		close(done)
		return
	}

	var wg sync.WaitGroup

	wg.Add(2)

	// stdout goroutine
	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(stdoutPipe)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "progress ") {
				progressString := strings.TrimPrefix(line, "progress ")
				progressFloat, err := strconv.ParseFloat(progressString, 64)
				if err == nil {
					job.Set("progress", int(progressFloat))
					save()

					select {
					case progressUpdated <- struct{}{}:
					default:
					}
					continue
				}
			}

			if strings.HasPrefix(line, "result ") {
				resultString := strings.TrimPrefix(line, "result ")
				job.Set("result", resultString)
				save()
				continue
			}

			if strings.HasPrefix(line, "error ") {
				errorString := strings.TrimPrefix(line, "error ")
				job.Set("error", errorString)
				save()
				continue
			}

			addLog("INFO", line)
		}

		if err := scanner.Err(); err != nil {
			addLog("ERROR", "stdout scanner error: "+err.Error())
		}
	}()

	// stderr goroutine
	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(stderrPipe)

		for scanner.Scan() {
			line := scanner.Text()
			addLog("ERROR", line)
		}

		if err := scanner.Err(); err != nil {
			addLog("ERROR", "stderr scanner error: "+err.Error())
		}
	}()

	if err := cmd.Start(); err != nil {
		logger.Error("❌ job start error", "id", id, "err", err)
		job.Set("status", "failed")
		job.Set("error", err.Error())
		save()
		close(done)
		return
	}

	if err := cmd.Wait(); err != nil {
		logger.Error("❌ job wait failed", "id", id, "error", err)
		job.Set("status", "failed")
		job.Set("error", err.Error())
		save()

		wg.Wait()
		close(done)
		return
	}

	// Attendre que stdout/stderr se terminent
	wg.Wait()
	close(done)

	time.Sleep(100 * time.Millisecond)

	job.Set("status", "finished")
	job.Set("progress", 100)
	save()

	logger.Info("✅ job finished", "id", id)
}

func bindJobs(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateSuccess("jobs").BindFunc(func(e *core.RecordEvent) error {
		job := e.Record.Clone()
		app := e.App

		go func() {
			onJobCreate(app, job)
		}()

		return nil
	})
}
