// medias.go
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
	Streams []struct {
		Type   string `json:"codec_type,omitempty"`
		Codec  string `json:"codec_name,omitempty"`
		Width  int    `json:"width,omitempty"`
		Height int    `json:"height,omitempty"`
	} `json:"streams"`
}

type MediaData struct {
	Width      int         `json:"width,omitempty"`
	Height     int         `json:"height,omitempty"`
	DurationMs int         `json:"durationMs,omitempty"`
	FFProbe    interface{} `json:"ffprobe,omitempty"`
}

func getMimeType(logger *slog.Logger, file *filesystem.File) string {
	reader, err := file.Reader.Open()
	if err != nil {
		logger.Error("❌ Erreur ouverture fichier", "err", err)
		return "unknown"
	}
	defer reader.Close()

	// Lire les 512 premiers octets
	header := make([]byte, 512)
	_, err = reader.Read(header)
	if err != nil && err != io.EOF {
		logger.Error("❌ Erreur lecture fichier", "err", err)
		return "unknown"
	}

	// Détection du mime type
	mimeType := http.DetectContentType(header)
	logger.Info("🔍 MIME type détecté", "mimetype", mimeType)

	return mimeType
}

func getVideoInfo(logger *slog.Logger, file *filesystem.File) (*MediaData, error) {
	// Ouvrir le fichier (ReadSeekCloser)
	reader, err := file.Reader.Open()
	if err != nil {
		logger.Error("❌ Erreur ouverture fichier", "err", err)
		return nil, err
	}
	defer reader.Close()

	// Préparer ffprobe avec stdin
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-i", "pipe:0",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Stdin = reader

	if err := cmd.Run(); err != nil {
		logger.Error("❌ Erreur exécution ffprobe", "err", err, "out", out.String())
		return nil, err
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		logger.Error("❌ Erreur parsing JSON ffprobe", "err", err)
		return nil, err
	}

	// Créer la structure MediaData
	mediaData := &MediaData{}

	// Chercher le stream vidéo
	for _, stream := range probe.Streams {
		if stream.Type == "video" {
			mediaData.Width = stream.Width
			mediaData.Height = stream.Height
			break
		}
	}

	// Duration en millisecondes
	durationFloat, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err == nil {
		mediaData.DurationMs = int(durationFloat * 1000)
	}

	// Stocker le ffprobe complet
	var ffprobeData interface{}
	json.Unmarshal(out.Bytes(), &ffprobeData)
	mediaData.FFProbe = ffprobeData

	return mediaData, nil
}

func getImageInfo(logger *slog.Logger, file *filesystem.File) (*MediaData, error) {
	// Ouvrir le fichier (ReadSeekCloser)
	reader, err := file.Reader.Open()
	if err != nil {
		logger.Error("❌ Erreur ouverture fichier", "err", err)
		return nil, err
	}
	defer reader.Close()

	// image.DecodeConfig lit les dimensions sans décoder toute l'image
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		logger.Error("❌ Erreur lecture dimensions image", "err", err)
		return nil, err
	}

	mediaData := &MediaData{
		Width:  cfg.Width,
		Height: cfg.Height,
	}

	return mediaData, nil
}

func processMediaFile(e *core.RecordRequestEvent) error {
	app := e.App
	media := e.Record
	logger := app.Logger()

	// Vérifier si le fichier existe ou a été modifié
	file, ok := media.GetRaw("file").(*filesystem.File)
	if !ok || file == nil {
		logger.Info("📦 Pas de nouveau fichier")
		return e.Next()
	}

	logger.Info("📦 Processing media file", "filename", file.OriginalName, "size", file.Size)

	// Toujours mettre à jour size et type
	media.Set("size", file.Size)

	// Si le nom est vide, utiliser le nom du fichier
	if media.GetString("name") == "" {
		media.Set("name", file.Name)
	}

	// Détecter le mime type
	mimeType := getMimeType(logger, file)
	media.Set("type", mimeType)

	// Traiter selon le type de média
	var mediaData *MediaData
	var err error

	if strings.HasPrefix(mimeType, "video") {
		mediaData, err = getVideoInfo(logger, file)
		if err != nil {
			logger.Error("❌ Erreur traitement vidéo", "err", err)
		}
	} else if strings.HasPrefix(mimeType, "image") {
		mediaData, err = getImageInfo(logger, file)
		if err != nil {
			logger.Error("❌ Erreur traitement image", "err", err)
		}
	}

	// Si on a des données, les sauvegarder dans le champ data
	if mediaData != nil {
		media.Set("data", mediaData)
	}

	return e.Next()
}

func bindMedias(app *pocketbase.PocketBase) {
	// Hook pour la création
	app.OnRecordCreateRequest("medias").BindFunc(processMediaFile)

	// Hook pour la modification
	app.OnRecordUpdateRequest("medias").BindFunc(processMediaFile)
}
