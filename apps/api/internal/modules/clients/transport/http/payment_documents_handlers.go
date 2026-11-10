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

	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
	mware "riders.agency/pulse/api/internal/platform/httpserver/middleware"
	"riders.agency/pulse/api/internal/platform/problem"
)

const maxPaymentDocumentBytes int64 = 25 * 1024 * 1024

var unsafePaymentDocumentFileNameRE = regexp.MustCompile(`[^a-zA-Z0-9._ -]+`)
var errPaymentDocumentTooLarge = errors.New("file is too large")

func ListPaymentActionDocuments(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		actionID, ok := paymentActionIDParam(w, r)
		if !ok {
			return
		}
		rows, err := cl.ListPaymentActionDocuments(r.Context(), actionID, claims.UserID)
		writeRowsOrError(w, rows, err)
	}
}

func UploadPaymentActionDocument(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		actionID, ok := paymentActionIDParam(w, r)
		if !ok {
			return
		}
		file, err := receivePaymentDocumentFile(w, r)
		if err != nil {
			problem.WriteModule(w, http.StatusBadRequest, "clients", "document_upload_failed", "Документ не загружен", err.Error(), "Проверьте файл и повторите загрузку.")
			return
		}
		row, err := cl.CreatePaymentActionDocument(r.Context(), actionID, claims.UserID, clientsrepo.PaymentDocumentInput{
			FileName:    file.FileName,
			ContentType: file.ContentType,
			SizeBytes:   file.SizeBytes,
			Kind:        firstFormValue(r, "kind"),
			StoragePath: file.StoragePath,
		})
		if writeClientError(w, err) {
			_ = os.Remove(file.StoragePath)
			return
		}
		problem.JSON(w, http.StatusCreated, row)
	}
}

func DownloadPaymentActionDocument(cl *clientssvc.Commands) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := mware.SessionClaimsFrom(r.Context())
		if !ok {
			problem.Write(w, http.StatusUnauthorized, "unauthorized", "")
			return
		}
		documentID, err := uuid.Parse(chi.URLParam(r, "payment_action_document_id"))
		if err != nil {
			problem.Write(w, http.StatusBadRequest, "payment_action_document_id", "")
			return
		}
		row, storagePath, err := cl.GetPaymentActionDocument(r.Context(), documentID, claims.UserID)
		if writeClientError(w, err) {
			return
		}
		w.Header().Set("Content-Type", row.ContentType)
		w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(row.FileName, `"`, "")+`"`)
		http.ServeFile(w, r, storagePath)
	}
}

type receivedPaymentDocument struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	StoragePath string
}

func receivePaymentDocumentFile(w http.ResponseWriter, r *http.Request) (receivedPaymentDocument, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPaymentDocumentBytes+1024)
	if err := r.ParseMultipartForm(maxPaymentDocumentBytes + 1024); err != nil {
		return receivedPaymentDocument{}, err
	}
	uploaded, header, err := r.FormFile("file")
	if err != nil {
		return receivedPaymentDocument{}, err
	}
	defer uploaded.Close()
	if header.Size > maxPaymentDocumentBytes {
		return receivedPaymentDocument{}, errPaymentDocumentTooLarge
	}
	name := safePaymentDocumentFileName(header.Filename)
	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	root, err := paymentDocumentUploadRoot()
	if err != nil {
		return receivedPaymentDocument{}, err
	}
	day := time.Now().Format("2006-01-02")
	dir := filepath.Join(root, day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return receivedPaymentDocument{}, err
	}
	path := filepath.Join(dir, uuid.NewString()+strings.ToLower(filepath.Ext(name)))
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return receivedPaymentDocument{}, err
	}
	written, copyErr := io.Copy(dst, io.LimitReader(uploaded, maxPaymentDocumentBytes+1))
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return receivedPaymentDocument{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return receivedPaymentDocument{}, closeErr
	}
	if written > maxPaymentDocumentBytes {
		_ = os.Remove(path)
		return receivedPaymentDocument{}, errPaymentDocumentTooLarge
	}
	return receivedPaymentDocument{FileName: name, ContentType: contentType, SizeBytes: written, StoragePath: path}, nil
}

func paymentDocumentUploadRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("RIDERS_PULSE_UPLOAD_DIR")); root != "" {
		return filepath.Clean(filepath.Join(root, "payment-documents")), nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(wd) == "api" && filepath.Base(filepath.Dir(wd)) == "apps" {
		return filepath.Clean(filepath.Join(wd, "..", "..", "uploads", "payment-documents")), nil
	}
	return filepath.Clean(filepath.Join(wd, "uploads", "payment-documents")), nil
}

func safePaymentDocumentFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		base = "document"
	}
	base = unsafePaymentDocumentFileNameRE.ReplaceAllString(base, "_")
	base = strings.Trim(base, " .")
	if base == "" {
		base = "document"
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

func firstFormValue(r *http.Request, name string) string {
	if r.MultipartForm == nil {
		return ""
	}
	values := r.MultipartForm.Value[name]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func paymentActionIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	actionID, err := uuid.Parse(chi.URLParam(r, "payment_action_id"))
	if err != nil {
		problem.Write(w, http.StatusBadRequest, "payment_action_id", "")
		return uuid.Nil, false
	}
	return actionID, true
}
