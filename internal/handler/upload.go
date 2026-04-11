package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/pkg/response"
)

type UploadHandler struct {
	uploadDir      string
	maxUploadBytes int64
	log            *slog.Logger
}

func NewUploadHandler(uploadDir string, maxUploadSizeMB int64, log *slog.Logger) *UploadHandler {
	return &UploadHandler{
		uploadDir:      uploadDir,
		maxUploadBytes: maxUploadSizeMB * 1024 * 1024,
		log:            log,
	}
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		response.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		response.Error(w, http.StatusBadRequest,
			fmt.Sprintf("file type %s not allowed, use: jpg, jpeg, png, gif, webp, svg", ext))
		return
	}

	filename := uuid.New().String() + ext
	destPath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		h.log.Error("creating upload file",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.log.Error("writing upload file",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	fileURL := "/uploads/" + filename

	response.JSON(w, http.StatusCreated, map[string]string{
		"url":      fileURL,
		"filename": filename,
	})
}
