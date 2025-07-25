// transcode.go - Version corrigée
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// Structure pour parser la sortie JSON de ffprobe (transcodage)
type TranscodeFFProbeOutput struct {
	Streams []TranscodeFFProbeStream `json:"streams"`
	Format  TranscodeFFProbeFormat   `json:"format"`
}

type TranscodeFFProbeStream struct {
	Index      int    `json:"index"`
	CodecType  string `json:"codec_type"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	RFrameRate string `json:"r_frame_rate,omitempty"`
	Duration   string `json:"duration,omitempty"`
	NBFrames   string `json:"nb_frames,omitempty"`
	DurationTS int64  `json:"duration_ts,omitempty"`
	TimeBase   string `json:"time_base,omitempty"`
}

type TranscodeFFProbeFormat struct {
	Duration string `json:"duration"`
}

// Profils de transcodage prédéfinis
var transcodeProfiles = map[string]TranscodeProfile{
	"SD": {
		Name:      "SD",
		Width:     640,
		Height:    360,
		Bitrate:   "500k",
		AudioRate: "64k",
		Preset:    "ultrafast",
		CRF:       25,
	},
	"HD": {
		Name:      "HD",
		Width:     1280,
		Height:    720,
		Bitrate:   "2500k",
		AudioRate: "128k",
		Preset:    "ultrafast",
		CRF:       25,
	},
	"FHD": {
		Name:      "FHD",
		Width:     1920,
		Height:    1080,
		Bitrate:   "5000k",
		AudioRate: "192k",
		Preset:    "ultrafast",
		CRF:       25,
	},
	"UHD": {
		Name:      "UHD",
		Width:     3840,
		Height:    2160,
		Bitrate:   "15000k",
		AudioRate: "256k",
		Preset:    "fast", // Pour UHD, on garde "fast" pour un meilleur équilibre qualité/vitesse
		CRF:       23,
	},
}

// Formats supportés
var supportedFormats = map[string]FormatConfig{
	"H264": {
		Name:      "H264",
		Codec:     "libx264",
		Extension: ".mp4",
		MimeType:  "video/mp4",
	},
	"H265": {
		Name:      "H265",
		Codec:     "libx265",
		Extension: ".mp4",
		MimeType:  "video/mp4",
	},
	"VP8": {
		Name:      "VP8",
		Codec:     "libvpx",
		Extension: ".webm",
		MimeType:  "video/webm",
	},
	"VP9": {
		Name:      "VP9",
		Codec:     "libvpx-vp9",
		Extension: ".webm",
		MimeType:  "video/webm",
	},
	"JPEG": {
		Name:      "JPEG",
		Codec:     "mjpeg",
		Extension: ".jpg",
		MimeType:  "image/jpeg",
	},
}

type TranscodeProfile struct {
	Name      string
	Width     int
	Height    int
	Bitrate   string
	AudioRate string
	Preset    string
	CRF       int
}

type FormatConfig struct {
	Name      string
	Codec     string
	Extension string
	MimeType  string
}

// Handler unifié pour le transcodage et la récupération
// Route: /api/medias/{id}/transcode/{profile}/{format}/{fake_name}
func transcodeHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		mediaId := e.Request.PathValue("id")
		profileName := e.Request.PathValue("profile")
		formatName := e.Request.PathValue("format")
		fakeFilename := e.Request.PathValue("fake_name") // Nom souhaité par le client pour le téléchargement

		logger := app.Logger()

		// Vérifier l'authentification
		authRecord := e.Auth
		if authRecord == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authentication required",
			})
		}

		// Récupérer le record media original
		originalRecord, err := app.FindRecordById("medias", mediaId)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{
				"error": "Media not found",
			})
		}

		// Vérifier que c'est une vidéo (sauf pour JPEG qui peut être une image)
		mimeType := originalRecord.GetString("type")
		if formatName != "JPEG" && !strings.HasPrefix(mimeType, "video") {
			return e.JSON(http.StatusBadRequest, map[string]string{
				"error": "Media is not a video",
			})
		}

		// Vérifier le profil
		profile, profileExists := transcodeProfiles[profileName]
		if !profileExists {
			availableProfiles := make([]string, 0, len(transcodeProfiles))
			for k := range transcodeProfiles {
				availableProfiles = append(availableProfiles, k)
			}
			return e.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Unknown profile: %s. Available: %s", profileName, strings.Join(availableProfiles, ", ")),
			})
		}

		// Vérifier le format
		format, formatExists := supportedFormats[formatName]
		if !formatExists {
			availableFormats := make([]string, 0, len(supportedFormats))
			for k := range supportedFormats {
				availableFormats = append(availableFormats, k)
			}
			return e.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Unknown format: %s. Available: %s", formatName, strings.Join(availableFormats, ", ")),
			})
		}

		// Vérifier les paramètres de query
		queryParams := e.Request.URL.Query()
		isStatus := queryParams.Has("status")
		isDownload := queryParams.Has("download")

		// Chercher si un transcodage existe déjà
		transcodeRecord, err := findTranscodeRecord(app, mediaId, profileName, formatName)

		// Si le transcodage existe
		if err == nil && transcodeRecord != nil {
			status := transcodeRecord.GetString("status")

			switch status {
			case "finished":
				if isDownload {
					// Servir le fichier pour téléchargement avec le nom souhaité par le client
					finalFilename := fakeFilename + format.Extension
					return serveTranscodeFile(app, transcodeRecord, finalFilename, true, format.MimeType, e)
				} else {
					// Retourner les informations
					fileUrl := fmt.Sprintf("/api/files/transcodes/%s/%s", transcodeRecord.Id, transcodeRecord.GetString("output"))
					downloadUrl := fmt.Sprintf("/api/medias/%s/transcode/%s/%s/%s?download", mediaId, profileName, formatName, fakeFilename)

					return e.JSON(http.StatusOK, map[string]interface{}{
						"status":       "ready",
						"file_url":     fileUrl,
						"download_url": downloadUrl,
						"profile":      profileName,
						"format":       formatName,
						"progress":     100,
						"created":      transcodeRecord.GetDateTime("created"),
					})
				}
			case "processing":
				progress := transcodeRecord.GetInt("progress")
				return e.JSON(http.StatusAccepted, map[string]interface{}{
					"status":   "processing",
					"progress": progress,
					"message":  "Transcoding in progress",
				})
			case "failed":
				errorMsg := transcodeRecord.GetString("error")
				return e.JSON(http.StatusInternalServerError, map[string]interface{}{
					"status": "error",
					"error":  errorMsg,
					"logs":   transcodeRecord.GetString("logs"),
				})
			}
		}

		// Si demande de statut uniquement et pas de transcodage
		if isStatus {
			return e.JSON(http.StatusNotFound, map[string]string{
				"status":  "not_found",
				"message": "Transcode not found",
			})
		}

		// Créer un nouveau record de transcodage
		transcodeRecord, err = createTranscodeRecord(app, mediaId, profileName, formatName)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create transcode record",
			})
		}

		// Lancer le transcodage en arrière-plan
		go func() {
			if err := performTranscode(app, originalRecord, transcodeRecord, profile, format); err != nil {
				logger.Error("❌ Erreur transcodage", "err", err, "mediaId", mediaId, "profile", profileName, "format", formatName)
			} else {
				logger.Info("✅ Transcodage terminé", "mediaId", mediaId, "profile", profileName, "format", formatName)
			}
		}()

		return e.JSON(http.StatusAccepted, map[string]interface{}{
			"status":       "processing",
			"progress":     0,
			"transcode_id": transcodeRecord.Id,
			"message":      "Transcoding started",
		})
	}
}

// Trouve un enregistrement de transcodage existant
func findTranscodeRecord(app *pocketbase.PocketBase, mediaId, profile, format string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"transcodes",
		"media = {:mediaId} && profile = {:profile} && format = {:format}",
		"-created",
		1,
		0,
		map[string]any{
			"mediaId": mediaId,
			"profile": profile,
			"format":  format,
		},
	)

	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("no transcode record found")
	}

	return records[0], nil
}

// Crée un nouveau record de transcodage
func createTranscodeRecord(app *pocketbase.PocketBase, mediaId, profile, format string) (*core.Record, error) {
	transcodeCollection, err := app.FindCollectionByNameOrId("transcodes")
	if err != nil {
		return nil, fmt.Errorf("transcodes collection not found: %w", err)
	}

	// Supprimer l'ancien record s'il existe
	if existingRecord, err := findTranscodeRecord(app, mediaId, profile, format); err == nil {
		app.Delete(existingRecord)
	}

	// Récupérer le media pour obtenir le group
	mediaRecord, err := app.FindRecordById("medias", mediaId)
	if err != nil {
		return nil, fmt.Errorf("media record not found: %w", err)
	}

	// Créer un nouveau record
	newRecord := core.NewRecord(transcodeCollection)
	newRecord.Set("media", mediaId)
	newRecord.Set("profile", profile)
	newRecord.Set("format", format)
	newRecord.Set("status", "processing")
	newRecord.Set("progress", 0)
	newRecord.Set("group", mediaRecord.GetString("group")) // Copier le group du media

	if err := app.Save(newRecord); err != nil {
		return nil, fmt.Errorf("failed to save transcode record: %w", err)
	}

	return newRecord, nil
}

// Sert un fichier depuis la collection transcodes
func serveTranscodeFile(app *pocketbase.PocketBase, record *core.Record, customFilename string, forDownload bool, contentType string, e *core.RequestEvent) error {
	fileName := record.GetString("output")
	if fileName == "" {
		return e.JSON(http.StatusNotFound, map[string]string{
			"error": "File not found",
		})
	}

	// Construire le chemin du fichier
	filePath := filepath.Join(app.DataDir(), "storage", record.Collection().Id, record.Id, fileName)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return e.JSON(http.StatusNotFound, map[string]string{
			"error": "File not found on disk",
		})
	}

	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer file.Close()

	// Obtenir les informations du fichier
	fileInfo, err := file.Stat()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get file info",
		})
	}

	// Définir les headers appropriés
	w := e.Response
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Headers de disposition
	if forDownload {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, customFilename))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, customFilename))
	}

	// Copier le contenu du fichier vers la réponse
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}

	return nil
}

// Effectue le transcodage complet avec mise à jour des logs et progression
func performTranscode(app *pocketbase.PocketBase, originalRecord *core.Record, transcodeRecord *core.Record, profile TranscodeProfile, format FormatConfig) error {
	logger := app.Logger()

	// Log de démarrage
	logger.Info("🚀 Démarrage du transcodage", "recordId", transcodeRecord.Id, "profile", profile.Name, "format", format.Name)
	updateTranscodeProgress(app, transcodeRecord, 1, "=== TRANSCODING STARTED ===")

	// Récupérer le fichier original
	originalFile := originalRecord.GetString("file")
	if originalFile == "" {
		logger.Error("❌ Aucun fichier trouvé dans le record original")
		updateTranscodeError(app, transcodeRecord, "No file found in original record")
		return fmt.Errorf("no file found")
	}

	logger.Info("📁 Fichier source trouvé", "file", originalFile)
	updateTranscodeProgress(app, transcodeRecord, 2, fmt.Sprintf("Source file: %s", originalFile))

	// Construire le chemin complet du fichier source
	sourcePath := filepath.Join(app.DataDir(), "storage", originalRecord.Collection().Id, originalRecord.Id, originalFile)
	logger.Info("📂 Chemin source construit", "path", sourcePath)
	updateTranscodeProgress(app, transcodeRecord, 3, fmt.Sprintf("Source path: %s", sourcePath))

	// Vérifier que le fichier source existe
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		logger.Error("❌ Fichier source inexistant", "path", sourcePath)
		updateTranscodeError(app, transcodeRecord, fmt.Sprintf("Source file not found: %s", sourcePath))
		return fmt.Errorf("source file not found")
	}
	logger.Info("✅ Fichier source vérifié")
	updateTranscodeProgress(app, transcodeRecord, 4, "Source file verified")

	// Créer un fichier temporaire pour la sortie
	tempDir := os.TempDir()
	outputFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s_%s%s", transcodeRecord.Id, profile.Name, format.Name, format.Extension))
	logger.Info("📝 Fichier temporaire créé", "path", outputFile)
	updateTranscodeProgress(app, transcodeRecord, 5, fmt.Sprintf("Temp output file: %s", outputFile))

	defer func() {
		logger.Info("🧹 Nettoyage du fichier temporaire", "path", outputFile)
		os.Remove(outputFile)
	}()

	// 1. Analyser la vidéo avec ffprobe (sauf pour JPEG)
	var totalFrames int
	var videoDuration float64
	if format.Name != "JPEG" {
		logger.Info("🔍 === PHASE 1: ANALYSE FFPROBE ===")
		updateTranscodeProgress(app, transcodeRecord, 6, "=== PHASE 1: FFPROBE ANALYSIS ===")

		probeInfo, frames, duration, err := analyzeWithFFProbe(sourcePath)
		if err != nil {
			logger.Error("❌ Erreur FFProbe", "err", err)
			updateTranscodeError(app, transcodeRecord, fmt.Sprintf("FFProbe error: %v", err))
			return err
		}
		totalFrames = frames
		videoDuration = duration
		logger.Info("📊 Analyse FFProbe terminée", "totalFrames", totalFrames, "duration", videoDuration, "probeInfoLength", len(probeInfo))
		updateTranscodeProgress(app, transcodeRecord, 10, fmt.Sprintf("FFProbe analysis complete - %d frames detected (%.2fs)", totalFrames, videoDuration))

		// Sauvegarder les infos complètes dans les logs
		currentLogs := transcodeRecord.GetString("logs")
		transcodeRecord.Set("logs", currentLogs+"\n=== FFPROBE OUTPUT ===\n"+probeInfo+"\n=== END FFPROBE ===")
		app.Save(transcodeRecord)
	}

	// 2. Transcoder le fichier
	if format.Name == "JPEG" {
		// Extraction d'image
		logger.Info("🖼️ === PHASE 2: EXTRACTION IMAGE ===")
		updateTranscodeProgress(app, transcodeRecord, 15, "=== PHASE 2: IMAGE EXTRACTION ===")

		if err := extractImage(sourcePath, outputFile, transcodeRecord, app); err != nil {
			logger.Error("❌ Erreur extraction image", "err", err)
			updateTranscodeError(app, transcodeRecord, fmt.Sprintf("Image extraction error: %v", err))
			return err
		}
		logger.Info("✅ Extraction image terminée")
		updateTranscodeProgress(app, transcodeRecord, 90, "Image extraction completed")
	} else {
		// Transcodage vidéo
		logger.Info("🎬 === PHASE 2: TRANSCODAGE VIDEO ===")
		updateTranscodeProgress(app, transcodeRecord, 15, "=== PHASE 2: VIDEO TRANSCODING ===")

		if err := transcodeVideo(sourcePath, outputFile, profile, format, transcodeRecord, totalFrames, videoDuration, app); err != nil {
			logger.Error("❌ Erreur transcodage vidéo", "err", err)
			updateTranscodeError(app, transcodeRecord, fmt.Sprintf("Video transcoding error: %v", err))
			return err
		}
		logger.Info("✅ Transcodage vidéo terminé")
		updateTranscodeProgress(app, transcodeRecord, 90, "Video transcoding completed")
	}

	// Vérifier que le fichier de sortie existe
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		logger.Error("❌ Fichier de sortie non créé", "path", outputFile)
		updateTranscodeError(app, transcodeRecord, fmt.Sprintf("Output file not created: %s", outputFile))
		return fmt.Errorf("output file not created")
	}

	// Obtenir la taille du fichier de sortie
	if fileInfo, err := os.Stat(outputFile); err == nil {
		logger.Info("📏 Fichier de sortie créé", "path", outputFile, "size", fileInfo.Size())
		updateTranscodeProgress(app, transcodeRecord, 92, fmt.Sprintf("Output file created - size: %d bytes", fileInfo.Size()))
	}

	// 3. Sauvegarder le fichier dans le record
	logger.Info("💾 === PHASE 3: SAUVEGARDE ===")
	updateTranscodeProgress(app, transcodeRecord, 95, "=== PHASE 3: SAVING FILE ===")

	if err := saveTranscodedFile(app, transcodeRecord, outputFile); err != nil {
		logger.Error("❌ Erreur sauvegarde", "err", err)
		updateTranscodeError(app, transcodeRecord, fmt.Sprintf("File saving error: %v", err))
		return err
	}

	logger.Info("💾 Sauvegarde terminée")
	updateTranscodeProgress(app, transcodeRecord, 98, "File saved successfully")

	// 4. Finaliser
	logger.Info("🏁 === FINALISATION ===")
	transcodeRecord.Set("status", "finished")
	transcodeRecord.Set("progress", 100)

	// Log final dans les logs
	currentLogs := transcodeRecord.GetString("logs")
	transcodeRecord.Set("logs", currentLogs+"\n=== TRANSCODING COMPLETED SUCCESSFULLY ===")

	app.Save(transcodeRecord)

	logger.Info("✅ Transcodage terminé avec succès", "recordId", transcodeRecord.Id)
	return nil
}

// Analyse la vidéo avec ffprobe et extrait le nombre total de frames et la durée
func analyzeWithFFProbe(inputPath string) (string, int, float64, error) {
	// Log de démarrage
	fmt.Printf("🔍 Analyse FFProbe démarrée pour: %s\n", inputPath)

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	fmt.Printf("📋 Commande FFProbe: ffprobe %s\n", strings.Join(args, " "))

	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Erreur FFProbe: %v\n", err)
		return "", 0, 0, err
	}

	fmt.Printf("✅ FFProbe terminé, sortie de %d caractères\n", len(output))

	// Parser le JSON
	var probeOutput TranscodeFFProbeOutput
	if err := json.Unmarshal(output, &probeOutput); err != nil {
		fmt.Printf("❌ Erreur parsing JSON FFProbe: %v\n", err)
		return string(output), 0, 0, nil // Retourner quand même la sortie brute
	}

	// Chercher le stream vidéo et extraire les informations
	var totalFrames int
	var duration float64

	// D'abord essayer d'obtenir la durée depuis le format
	if probeOutput.Format.Duration != "" {
		if d, err := strconv.ParseFloat(probeOutput.Format.Duration, 64); err == nil {
			duration = d
			fmt.Printf("📊 Durée trouvée dans format: %.2fs\n", duration)
		}
	}

	for _, stream := range probeOutput.Streams {
		if stream.CodecType == "video" {
			fmt.Printf("🎬 Stream vidéo trouvé: index=%d, résolution=%dx%d\n", stream.Index, stream.Width, stream.Height)

			// Essayer d'obtenir nb_frames directement
			if stream.NBFrames != "" {
				if frames, err := strconv.Atoi(stream.NBFrames); err == nil {
					totalFrames = frames
					fmt.Printf("🎬 Frames trouvées via nb_frames: %d\n", totalFrames)
					break
				}
			}

			// Si pas de nb_frames, essayer de calculer avec duration et framerate
			if totalFrames == 0 {
				// Essayer d'obtenir la durée du stream si pas trouvée dans format
				if duration == 0 && stream.Duration != "" {
					if d, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
						duration = d
						fmt.Printf("📊 Durée trouvée dans stream: %.2fs\n", duration)
					}
				}

				// Calculer avec framerate si on a la durée
				if duration > 0 && stream.RFrameRate != "" {
					// Parse du framerate (format: "30000/1001" ou "30")
					frameRateParts := strings.Split(stream.RFrameRate, "/")
					if len(frameRateParts) == 2 {
						if num, err1 := strconv.ParseFloat(frameRateParts[0], 64); err1 == nil {
							if den, err2 := strconv.ParseFloat(frameRateParts[1], 64); err2 == nil && den != 0 {
								frameRate := num / den
								totalFrames = int(duration * frameRate)
								fmt.Printf("🎬 Frames calculées: %.2fs * %.2ffps = %d\n", duration, frameRate, totalFrames)
								break
							}
						}
					} else if len(frameRateParts) == 1 {
						if frameRate, err := strconv.ParseFloat(frameRateParts[0], 64); err == nil {
							totalFrames = int(duration * frameRate)
							fmt.Printf("🎬 Frames calculées: %.2fs * %.2ffps = %d\n", duration, frameRate, totalFrames)
							break
						}
					}
				}
			}
			break
		}
	}

	// Si on n'a toujours pas de frames, essayer une méthode alternative avec ffprobe
	if totalFrames == 0 {
		fmt.Printf("⚠️ Tentative alternative pour compter les frames...\n")

		// Méthode alternative : compter les frames avec une commande spécifique
		frameArgs := []string{
			"-v", "error",
			"-select_streams", "v:0",
			"-count_packets",
			"-show_entries", "stream=nb_read_packets",
			"-csv=p=0",
			inputPath,
		}

		fmt.Printf("📊 Comptage alternatif: ffprobe %s\n", strings.Join(frameArgs, " "))

		frameCmd := exec.Command("ffprobe", frameArgs...)
		frameOutput, frameErr := frameCmd.Output()
		if frameErr == nil {
			if frames, parseErr := strconv.Atoi(strings.TrimSpace(string(frameOutput))); parseErr == nil {
				totalFrames = frames
				fmt.Printf("🎬 Frames trouvées via comptage alternatif: %d\n", totalFrames)
			} else {
				fmt.Printf("⚠️ Erreur parsing comptage alternatif: %v\n", parseErr)
			}
		} else {
			fmt.Printf("⚠️ Erreur comptage alternatif: %v\n", frameErr)
		}
	}

	fmt.Printf("✅ Analyse terminée: %d frames, %.2fs de durée\n", totalFrames, duration)
	return string(output), totalFrames, duration, nil
}

// Extrait une image
func extractImage(inputPath, outputPath string, transcodeRecord *core.Record, app *pocketbase.PocketBase) error {
	fmt.Printf("🖼️ === EXTRACTION IMAGE ===\n")
	fmt.Printf("📁 Input: %s\n", inputPath)
	fmt.Printf("📁 Output: %s\n", outputPath)

	args := []string{
		"-i", inputPath,
		"-ss", "0",
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	}

	commandLine := "ffmpeg " + strings.Join(args, " ")
	fmt.Printf("📋 Commande: %s\n", commandLine)

	// Sauvegarder la commande dans les logs
	currentLogs := transcodeRecord.GetString("logs")
	transcodeRecord.Set("logs", currentLogs+"\n=== IMAGE EXTRACTION COMMAND ===\n"+commandLine+"\n=== EXTRACTION STDERR ===")
	app.Save(transcodeRecord)

	cmd := exec.Command("ffmpeg", args...)

	// Capturer stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("❌ Erreur création pipe stderr: %v\n", err)
		return err
	}

	// Capturer stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ Erreur création pipe stdout: %v\n", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Erreur démarrage FFmpeg: %v\n", err)
		return err
	}

	fmt.Printf("✅ FFmpeg extraction démarré avec PID: %d\n", cmd.Process.Pid)

	// Capturer toute la sortie
	var allStderr strings.Builder
	var allStdout strings.Builder

	// Goroutine pour stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			allStderr.WriteString(line + "\n")
			fmt.Printf("📊 Extraction stderr: %s\n", line)
		}
	}()

	// Goroutine pour stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			allStdout.WriteString(line + "\n")
			fmt.Printf("📺 Extraction stdout: %s\n", line)
		}
	}()

	// Attendre la fin
	err = cmd.Wait()

	fmt.Printf("🏁 Extraction terminée avec code: %v\n", err)

	// Sauvegarder tous les logs
	stderrLogs := allStderr.String()
	stdoutLogs := allStdout.String()
	currentLogs = transcodeRecord.GetString("logs")

	newLogs := currentLogs + "\n" + stderrLogs + "\n=== EXTRACTION STDOUT ===\n" + stdoutLogs

	if err != nil {
		newLogs += fmt.Sprintf("\n=== EXTRACTION ERROR ===\nExit code: %v", err)
		transcodeRecord.Set("logs", newLogs)
		app.Save(transcodeRecord)

		fmt.Printf("❌ Extraction échouée: %v\n", err)
		return err
	}

	newLogs += "\n=== EXTRACTION COMPLETED ==="
	transcodeRecord.Set("logs", newLogs)
	app.Save(transcodeRecord)

	fmt.Printf("✅ Extraction image réussie\n")
	return nil
}

// Transcode la vidéo avec suivi de progression basé sur les frames et la durée
func transcodeVideo(inputPath, outputPath string, profile TranscodeProfile, format FormatConfig, transcodeRecord *core.Record, totalFrames int, videoDuration float64, app *pocketbase.PocketBase) error {
	logger := app.Logger()

	args := []string{
		"-i", inputPath,
		"-c:v", format.Codec,
		"-preset", profile.Preset,
	}

	// Ajouter CRF seulement pour les codecs qui le supportent
	if format.Codec == "libx264" || format.Codec == "libx265" {
		args = append(args, "-crf", fmt.Sprintf("%d", profile.CRF))
	}

	// Ajouter la résolution
	args = append(args,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
			profile.Width, profile.Height, profile.Width, profile.Height),
	)

	// Ajouter l'audio seulement si ce n'est pas un format image
	if format.Name != "JPEG" {
		args = append(args, "-c:a", "aac", "-b:a", profile.AudioRate)
	}

	// Optimisations spécifiques au format
	if format.Extension == ".mp4" {
		args = append(args, "-movflags", "+faststart")
	}

	// Forcer l'overwrite et configurer la sortie pour le parsing
	args = append(args, "-y", "-progress", "pipe:2", outputPath)

	commandLine := "ffmpeg " + strings.Join(args, " ")
	logger.Info("🚀 Commande FFmpeg", "command", commandLine)
	fmt.Printf("🎬 === COMMANDE FFMPEG ===\n%s\n", commandLine)

	// Sauvegarder la commande dans les logs
	currentLogs := transcodeRecord.GetString("logs")
	transcodeRecord.Set("logs", currentLogs+"\n=== FFMPEG COMMAND ===\n"+commandLine+"\n=== FFMPEG STDERR OUTPUT ===")
	app.Save(transcodeRecord)

	cmd := exec.Command("ffmpeg", args...)

	// Capturer stderr pour la progression
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Capturer stdout pour les logs généraux
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	fmt.Printf("✅ FFmpeg démarré avec PID: %d\n", cmd.Process.Pid)

	// Channel pour synchroniser la fin des goroutines
	done := make(chan bool, 2)
	var allStderr strings.Builder
	var allStdout strings.Builder

	// Goroutine pour capturer stdout (logs généraux)
	go func() {
		defer func() { done <- true }()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			allStdout.WriteString(line + "\n")
			fmt.Printf("📺 FFmpeg stdout: %s\n", line)
			logger.Info("📺 FFmpeg stdout", "line", line)
		}
		fmt.Printf("📺 Fin de capture stdout\n")
	}()

	// Goroutine pour parser la progression depuis stderr
	go func() {
		defer func() { done <- true }()
		scanner := bufio.NewScanner(stderr)
		var currentFrame int
		var currentTime float64

		for scanner.Scan() {
			line := scanner.Text()
			allStderr.WriteString(line + "\n")

			fmt.Printf("📊 FFmpeg stderr: %s\n", line)
			logger.Info("📊 FFmpeg stderr", "line", line)

			// Parser les lignes de progression de ffmpeg
			if strings.HasPrefix(line, "frame=") {
				// Format standard: frame=  123 fps=25.0 q=28.0 size=    1024kB time=00:00:04.92 bitrate=1702.1kbits/s speed=1.01x
				parseFFmpegProgressLine(line, &currentFrame, &currentTime)
			} else if strings.Contains(line, "frame=") {
				// Format alternatif dans la même ligne
				parseFFmpegProgressLine(line, &currentFrame, &currentTime)
			} else if strings.Contains(line, "out_time_ms=") {
				// Format de progression plus précis
				parseFFmpegProgressFormat(line, &currentTime)
			}

			// Calculer et mettre à jour la progression
			progress := calculateProgress(currentFrame, totalFrames, currentTime, videoDuration)
			if progress > 15 && progress <= 90 {
				var message string
				if totalFrames > 0 && currentFrame > 0 {
					percentage := float64(currentFrame) / float64(totalFrames) * 100
					message = fmt.Sprintf("Transcoding... frame %d/%d (%.1f%%)", currentFrame, totalFrames, percentage)
				} else if videoDuration > 0 && currentTime > 0 {
					percentage := currentTime / videoDuration * 100
					message = fmt.Sprintf("Transcoding... %.1fs/%.1fs (%.1f%%)", currentTime, videoDuration, percentage)
				} else {
					message = "Transcoding in progress..."
				}

				updateTranscodeProgress(app, transcodeRecord, progress, message)

				fmt.Printf("⏳ Progression: %s (progress=%d)\n", message, progress)
				logger.Info("⏳ Progression", "currentFrame", currentFrame, "totalFrames", totalFrames, "currentTime", currentTime, "videoDuration", videoDuration, "progress", progress)
			}
		}
		fmt.Printf("📊 Fin de capture stderr\n")
	}()

	// Attendre la fin du processus
	err = cmd.Wait()

	// Attendre que les goroutines finissent de lire les pipes
	<-done
	<-done

	fmt.Printf("🏁 FFmpeg terminé avec code: %v\n", err)

	// Sauvegarder TOUT stderr et stdout dans les logs
	finalLogs := allStdout.String()
	stderrLogs := allStderr.String()
	currentLogs = transcodeRecord.GetString("logs")

	newLogs := currentLogs + "\n" + stderrLogs + "\n=== FFMPEG STDOUT ===\n" + finalLogs

	if err != nil {
		newLogs += fmt.Sprintf("\n=== FFMPEG ERROR ===\nExit code: %v", err)
		transcodeRecord.Set("logs", newLogs)
		app.Save(transcodeRecord)

		fmt.Printf("❌ FFmpeg a échoué: %v\n", err)
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	newLogs += "\n=== FFMPEG COMPLETED SUCCESSFULLY ==="
	transcodeRecord.Set("logs", newLogs)
	app.Save(transcodeRecord)

	fmt.Printf("✅ Transcodage vidéo terminé avec succès\n")
	logger.Info("✅ Transcodage vidéo terminé avec succès")
	return nil
}

// Parse une ligne de progression ffmpeg standard
func parseFFmpegProgressLine(line string, currentFrame *int, currentTime *float64) {
	// Parser frame=
	if frameMatch := regexp.MustCompile(`frame=\s*(\d+)`).FindStringSubmatch(line); len(frameMatch) > 1 {
		if frame, err := strconv.Atoi(frameMatch[1]); err == nil {
			*currentFrame = frame
		}
	}

	// Parser time=HH:MM:SS.ms
	if timeMatch := regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`).FindStringSubmatch(line); len(timeMatch) > 4 {
		hours, _ := strconv.Atoi(timeMatch[1])
		minutes, _ := strconv.Atoi(timeMatch[2])
		seconds, _ := strconv.Atoi(timeMatch[3])
		milliseconds, _ := strconv.Atoi(timeMatch[4])

		*currentTime = float64(hours*3600+minutes*60+seconds) + float64(milliseconds)/100.0
	}
}

// Parse le format de progression détaillé de ffmpeg
func parseFFmpegProgressFormat(line string, currentTime *float64) {
	// Format: out_time_ms=12345678 (en microsecondes)
	if timeMatch := regexp.MustCompile(`out_time_ms=(\d+)`).FindStringSubmatch(line); len(timeMatch) > 1 {
		if microseconds, err := strconv.ParseInt(timeMatch[1], 10, 64); err == nil {
			*currentTime = float64(microseconds) / 1000000.0 // Convertir en secondes
		}
	}
}

// Calcule la progression en pourcentage basée sur les frames ou la durée
func calculateProgress(currentFrame, totalFrames int, currentTime, videoDuration float64) int {
	var percentage float64

	// Priorité aux frames si disponibles
	if totalFrames > 0 && currentFrame > 0 {
		percentage = float64(currentFrame) / float64(totalFrames) * 100
	} else if videoDuration > 0 && currentTime > 0 {
		percentage = currentTime / videoDuration * 100
	} else {
		return 15 // Progression minimale si pas d'info
	}

	// Mapper de 0-100% vers 15-90% pour laisser de la place aux autres phases
	progress := int(percentage*0.75) + 15
	if progress > 90 {
		progress = 90
	}
	if progress < 15 {
		progress = 15
	}

	return progress
}

// Sauvegarde le fichier transcodé
func saveTranscodedFile(app *pocketbase.PocketBase, transcodeRecord *core.Record, filePath string) error {
	fmt.Printf("💾 === SAUVEGARDE FICHIER ===\n")
	fmt.Printf("📁 Fichier à sauvegarder: %s\n", filePath)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("❌ Fichier inexistant: %s\n", filePath)
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Obtenir les infos du fichier
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("❌ Erreur stat fichier: %v\n", err)
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fmt.Printf("📏 Taille du fichier: %d bytes\n", fileInfo.Size())

	// Créer l'objet filesystem pour le fichier
	fileFs, err := filesystem.NewFileFromPath(filePath)
	if err != nil {
		fmt.Printf("❌ Erreur création filesystem: %v\n", err)
		return fmt.Errorf("failed to create filesystem: %w", err)
	}

	fmt.Printf("✅ Objet filesystem créé: %s\n", fileFs.Name)

	// Associer le fichier au record
	transcodeRecord.Set("output", []*filesystem.File{fileFs})

	fmt.Printf("🔗 Fichier associé au record\n")

	// Sauvegarder le record
	if err := app.Save(transcodeRecord); err != nil {
		fmt.Printf("❌ Erreur sauvegarde record: %v\n", err)
		return fmt.Errorf("failed to save transcoded record: %w", err)
	}

	fmt.Printf("✅ Record sauvegardé avec succès\n")

	// Log dans le champ logs
	currentLogs := transcodeRecord.GetString("logs")
	transcodeRecord.Set("logs", currentLogs+fmt.Sprintf("\n=== FILE SAVED ===\nFile: %s\nSize: %d bytes\nFilesystem name: %s", filePath, fileInfo.Size(), fileFs.Name))
	app.Save(transcodeRecord)

	return nil
}

// Met à jour la progression du transcodage
func updateTranscodeProgress(app *pocketbase.PocketBase, record *core.Record, progress int, message string) {
	fmt.Printf("📊 Progression: %d%% - %s\n", progress, message)

	record.Set("progress", progress)
	if message != "" {
		currentLogs := record.GetString("logs")
		timestamp := fmt.Sprintf("[%s]", time.Now().Format("15:04:05"))
		record.Set("logs", currentLogs+"\n"+timestamp+" "+message)
	}

	if err := app.Save(record); err != nil {
		fmt.Printf("❌ Erreur sauvegarde progression: %v\n", err)
	} else {
		fmt.Printf("✅ Progression sauvegardée\n")
	}
}

// Met à jour avec une erreur
func updateTranscodeError(app *pocketbase.PocketBase, record *core.Record, errorMsg string) {
	fmt.Printf("❌ ERREUR TRANSCODAGE: %s\n", errorMsg)

	record.Set("status", "failed")
	record.Set("error", errorMsg)

	// Ajouter l'erreur aux logs avec timestamp
	currentLogs := record.GetString("logs")
	timestamp := fmt.Sprintf("[%s]", time.Now().Format("15:04:05"))
	record.Set("logs", currentLogs+"\n"+timestamp+" ERROR: "+errorMsg)

	if err := app.Save(record); err != nil {
		fmt.Printf("❌ Erreur sauvegarde erreur: %v\n", err)
	} else {
		fmt.Printf("✅ Erreur sauvegardée\n")
	}
}

func bindTranscode(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Route REST pour transcodage des medias
		se.Router.GET("/api/medias/{id}/transcode/{profile}/{format}/{fake_name}",
			transcodeHandler(app),
		)

		return se.Next()
	})
}
