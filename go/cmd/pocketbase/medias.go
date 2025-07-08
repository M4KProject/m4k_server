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

func getVideoWidthHeightDuration(logger *slog.Logger, file *filesystem.File) (int, int, int) {
	var width, height, duration int = 0, 0, 0

	// Ouvrir le fichier (ReadSeekCloser)
	reader, err := file.Reader.Open()
	if err != nil {
		logger.Error("❌ Erreur ouverture fichier", "err", err)
		return 0, 0, 0
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
		return 0, 0, 0
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		logger.Error("❌ Erreur parsing JSON ffprobe", "err", err)
		return 0, 0, 0
	}

	// Duration et dimensions
	if len(probe.Streams) > 0 {
		width = probe.Streams[0].Width
		height = probe.Streams[0].Height
	}

	// duration est une string dans ffprobe
	durationFloat, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err == nil {
		duration = int(durationFloat * 1000)
	}

	return width, height, duration
}

func getImageWidthHeight(logger *slog.Logger, file *filesystem.File) (int, int) {
	// Ouvrir le fichier (ReadSeekCloser)
	reader, err := file.Reader.Open()
	if err != nil {
		logger.Error("❌ Erreur ouverture fichier", "err", err)
		return 0, 0
	}
	defer reader.Close()

	// image.DecodeConfig lit les dimensions sans décoder toute l'image
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		logger.Error("❌ Erreur lecture dimensions image", "err", err)
		return 0, 0
	}

	return cfg.Width, cfg.Height
}

func onBeforeCreateMedia(e *core.RecordRequestEvent) error {
	app := e.App
	media := e.Record
	logger := app.Logger()

	file, ok := media.GetRaw("file").(*filesystem.File)
	if !ok {
		logger.Error("❌ Impossible d'accéder au champ 'file'")
		return e.Next()
	}

	logger.Info("📦 onMediaCreate", "filename", file.OriginalName, "size", file.Size)

	media.Set("size", file.Size)

	if media.GetString("name") == "" {
		media.Set("name", file.Name)
	}

	mimeType := getMimeType(logger, file)
	media.Set("type", mimeType)

	if strings.HasPrefix(mimeType, "video") {
		width, height, duration := getVideoWidthHeightDuration(logger, file)
		media.Set("width", width)
		media.Set("height", height)
		media.Set("duration", duration)
	}

	if strings.HasPrefix(mimeType, "image") {
		width, height := getImageWidthHeight(logger, file)
		media.Set("width", width)
		media.Set("height", height)
	}

	return e.Next()
}

func bindMedias(app *pocketbase.PocketBase) {
	app.OnRecordCreateRequest("medias").BindFunc(onBeforeCreateMedia)
	// app.OnRecordAfterCreateSuccess("medias").BindFunc(onAfterCreateMedia)
}
