package http

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"riders.agency/pulse/api/internal/modules/worktracker/repository"
	"riders.agency/pulse/api/internal/modules/worktracker/service"
	"riders.agency/pulse/api/internal/platform/problem"
)

const maxTaskAttachmentBytes int64 = 25 * 1024 * 1024

var unsafeTaskAttachmentFileNameRE = regexp.MustCompile(`[^a-zA-Z0-9._ -]+`)
var errTaskAttachmentTooLarge = errors.New("file is too large")

func taskAttachments(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		rows, err := wt.Attachments(r.Context(), taskID, userID)
		write(w, rows, err)
	}
}

func uploadTaskAttachment(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, ok := parseUUID(w, chi.URLParam(r, "task_id"), "task_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		file, err := receiveTaskAttachmentFile(w, r)
		if err != nil {
			problem.WriteModule(w, http.StatusBadRequest, "work", "attachment_upload_failed", "Файл не загружен", err.Error(), "Проверьте файл и повторите загрузку.")
			return
		}
		row, err := wt.CreateAttachment(r.Context(), taskID, userID, repository.TaskAttachmentInput{
			FileName:    file.FileName,
			ContentType: file.ContentType,
			SizeBytes:   file.SizeBytes,
			StoragePath: file.StoragePath,
		})
		if err != nil {
			_ = os.Remove(file.StoragePath)
			writeError(w, err)
			return
		}
		problem.JSON(w, http.StatusCreated, row)
	}
}

func downloadTaskAttachment(wt *service.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attachmentID, ok := parseUUID(w, chi.URLParam(r, "task_attachment_id"), "task_attachment_id")
		if !ok {
			return
		}
		userID, ok := userID(w, r)
		if !ok {
			return
		}
		row, storagePath, err := wt.Attachment(r.Context(), attachmentID, userID)
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", row.ContentType)
		w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(row.FileName, `"`, "")+`"`)
		http.ServeFile(w, r, storagePath)
	}
}

type receivedTaskAttachment struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	StoragePath string
}

func receiveTaskAttachmentFile(w http.ResponseWriter, r *http.Request) (receivedTaskAttachment, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxTaskAttachmentBytes+1024)
	if err := r.ParseMultipartForm(maxTaskAttachmentBytes + 1024); err != nil {
		return receivedTaskAttachment{}, err
	}
	uploaded, header, err := r.FormFile("file")
	if err != nil {
		return receivedTaskAttachment{}, err
	}
	defer uploaded.Close()
	if header.Size > maxTaskAttachmentBytes {
		return receivedTaskAttachment{}, errTaskAttachmentTooLarge
	}
	name := safeTaskAttachmentFileName(header.Filename)
	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	root, err := taskAttachmentUploadRoot()
	if err != nil {
		return receivedTaskAttachment{}, err
	}
	day := time.Now().Format("2006-01-02")
	dir := filepath.Join(root, day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return receivedTaskAttachment{}, err
	}
	path := filepath.Join(dir, uuid.NewString()+strings.ToLower(filepath.Ext(name)))
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return receivedTaskAttachment{}, err
	}
	written, copyErr := io.Copy(dst, io.LimitReader(uploaded, maxTaskAttachmentBytes+1))
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return receivedTaskAttachment{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return receivedTaskAttachment{}, closeErr
	}
	if written > maxTaskAttachmentBytes {
		_ = os.Remove(path)
		return receivedTaskAttachment{}, errTaskAttachmentTooLarge
	}
	return receivedTaskAttachment{FileName: name, ContentType: contentType, SizeBytes: written, StoragePath: path}, nil
}

func taskAttachmentUploadRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("RIDERS_PULSE_UPLOAD_DIR")); root != "" {
		return filepath.Clean(filepath.Join(root, "task-attachments")), nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(wd) == "api" && filepath.Base(filepath.Dir(wd)) == "apps" {
		return filepath.Clean(filepath.Join(wd, "..", "..", "uploads", "task-attachments")), nil
	}
	return filepath.Clean(filepath.Join(wd, "uploads", "task-attachments")), nil
}

func safeTaskAttachmentFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		base = "attachment"
	}
	base = unsafeTaskAttachmentFileNameRE.ReplaceAllString(base, "_")
	base = strings.Trim(base, " .")
	if base == "" {
		base = "attachment"
	}
	if len(base) > 180 {
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		limit := 180 - len(ext)
		if limit < 24 {
			limit = 24
		}
		if len(stem) > limit {
			stem = stem[:limit]
		}
		base = stem + ext
	}
	return base
}
