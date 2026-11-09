package http

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	sitechatsvc "riders.agency/pulse/api/internal/modules/sitechat/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	sitechat "riders.agency/pulse/api/internal/modules/sitechat/repository"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

const maxSiteChatAttachmentBytes int64 = 25 * 1024 * 1024

var unsafeFileNameRE = regexp.MustCompile(`[^a-zA-Z0-9._ -]+`)

func PublicAttachmentUpload(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationID, ok := pathUUID(w, r, "conversation_id")
		if !ok {
			return
		}
		file, err := receiveAttachmentFile(w, r)
		if err != nil {
			writeSiteChatErr(w, err)
			return
		}
		scope, err := s.ResolvePublicScope(
			r.Context(),
			firstFormValue(r, "public_key"),
			firstFormValue(r, "visitor_id"),
			firstFormValue(r, "visit_session_id"),
			conversationID.String(),
		)
		if writeSiteChatErr(w, err) {
			_ = os.Remove(file.StoragePath)
			return
		}
		callSessionID := firstFormValue(r, "call_session_id")
		out, err := s.CreateAttachment(r.Context(), sitechat.AttachmentInput{
			Scope:          scope,
			CallSessionID:  callSessionID,
			UploadedByKind: "visitor",
			FileName:       file.FileName,
			ContentType:    file.ContentType,
			SizeBytes:      file.SizeBytes,
			StoragePath:    file.StoragePath,
			DownloadToken:  randomDownloadToken(),
			Kind:           classifyAttachmentKind(file.ContentType, file.FileName),
			Metadata: map[string]any{
				"source": "public_widget",
			},
		})
		if writeSiteChatErr(w, err) {
			_ = os.Remove(file.StoragePath)
			return
		}
		if strings.TrimSpace(callSessionID) != "" {
			out.PublicURL = ""
		}
		problem.JSON(w, http.StatusCreated, out)
	}
}

func OperatorAttachmentUpload(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		conversationID, ok := pathUUID(w, r, "conversation_id")
		if !ok {
			return
		}
		file, err := receiveAttachmentFile(w, r)
		if err != nil {
			writeSiteChatErr(w, err)
			return
		}
		scope, err := s.OperatorAttachmentScope(r.Context(), conversationID, claims.UserID)
		if writeSiteChatErr(w, err) {
			_ = os.Remove(file.StoragePath)
			return
		}
		out, err := s.CreateAttachment(r.Context(), sitechat.AttachmentInput{
			Scope:            scope,
			CallSessionID:    firstFormValue(r, "call_session_id"),
			UploadedByKind:   "operator",
			UploadedByUserID: claims.UserID.String(),
			FileName:         file.FileName,
			ContentType:      file.ContentType,
			SizeBytes:        file.SizeBytes,
			StoragePath:      file.StoragePath,
			DownloadToken:    randomDownloadToken(),
			Kind:             classifyAttachmentKind(file.ContentType, file.FileName),
			Metadata: map[string]any{
				"source": "operator_console",
			},
		})
		if writeSiteChatErr(w, err) {
			_ = os.Remove(file.StoragePath)
			return
		}
		problem.JSON(w, http.StatusCreated, out)
	}
}

func OperatorCallRecordings(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callID, ok := pathUUID(w, r, "call_session_id")
		if !ok {
			return
		}
		rows, err := s.CallRecordings(r.Context(), callID)
		writeSiteChat(w, http.StatusOK, rows, err)
	}
}

func AttachmentDownload(s *sitechatsvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attachmentID := strings.TrimSpace(chiParam(r, "attachment_id"))
		token := strings.TrimSpace(chiParam(r, "download_token"))
		row, err := s.GetAttachmentByToken(r.Context(), attachmentID, token)
		if writeSiteChatErr(w, err) {
			return
		}
		w.Header().Set("Content-Type", row.ContentType)
		w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(row.FileName, `"`, "")+`"`)
		http.ServeFile(w, r, row.StoragePath)
	}
}

type receivedAttachment struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	StoragePath string
}

func receiveAttachmentFile(w http.ResponseWriter, r *http.Request) (receivedAttachment, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSiteChatAttachmentBytes+1024)
	if err := r.ParseMultipartForm(maxSiteChatAttachmentBytes + 1024); err != nil {
		return receivedAttachment{}, sitechat.ErrBadInput
	}
	uploaded, header, err := r.FormFile("file")
	if err != nil {
		return receivedAttachment{}, sitechat.ErrBadInput
	}
	defer uploaded.Close()
	if header.Size > maxSiteChatAttachmentBytes {
		return receivedAttachment{}, sitechat.ErrBadInput
	}
	name := safeAttachmentFileName(header.Filename)
	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	root, err := attachmentUploadRoot()
	if err != nil {
		return receivedAttachment{}, err
	}
	day := time.Now().Format("2006-01-02")
	dir := filepath.Join(root, day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return receivedAttachment{}, err
	}
	stored := uuid.NewString() + strings.ToLower(filepath.Ext(name))
	path := filepath.Join(dir, stored)
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return receivedAttachment{}, err
	}
	written, copyErr := io.Copy(dst, io.LimitReader(uploaded, maxSiteChatAttachmentBytes+1))
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return receivedAttachment{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return receivedAttachment{}, closeErr
	}
	if written > maxSiteChatAttachmentBytes {
		_ = os.Remove(path)
		return receivedAttachment{}, sitechat.ErrBadInput
	}
	return receivedAttachment{FileName: name, ContentType: contentType, SizeBytes: written, StoragePath: path}, nil
}

func attachmentUploadRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("RIDERS_PULSE_UPLOAD_DIR")); root != "" {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(wd) == "api" && filepath.Base(filepath.Dir(wd)) == "apps" {
		return filepath.Clean(filepath.Join(wd, "..", "..", "uploads", "site-chat")), nil
	}
	return filepath.Clean(filepath.Join(wd, "uploads", "site-chat")), nil
}

func safeAttachmentFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		base = "attachment"
	}
	base = unsafeFileNameRE.ReplaceAllString(base, "_")
	base = strings.Trim(base, " .")
	if base == "" {
		base = "attachment"
	}
	if len(base) > 180 {
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		if len(ext) > 24 {
			ext = ext[:24]
		}
		limit := 180 - len(ext)
		if limit < 24 {
			limit = 24
		}
		base = stem[:limit] + ext
	}
	return base
}

func classifyAttachmentKind(contentType, name string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case strings.HasPrefix(ct, "image/") || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp":
		return "image"
	case strings.HasPrefix(ct, "audio/") || ext == ".webm" || ext == ".ogg" || ext == ".mp3" || ext == ".wav":
		return "audio"
	case strings.HasPrefix(ct, "video/") || ext == ".mp4" || ext == ".mov":
		return "video"
	default:
		return "file"
	}
}

func firstFormValue(r *http.Request, name string) string {
	if value := strings.TrimSpace(r.FormValue(name)); value != "" {
		return value
	}
	return strings.TrimSpace(r.URL.Query().Get(name))
}

func randomDownloadToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewString()
	}
	return hex.EncodeToString(buf)
}

func chiParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
