package runtime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
)

const (
	maxEmailDocumentCount = 8
	maxEmailDocumentBytes = 18 << 20
)

type emailDocumentAttachment struct {
	FileName string
	Body     []byte
}

func requireDocumentAttachments(provider string, required bool, count int) error {
	if !required || count > 0 {
		return nil
	}
	return fmt.Errorf("%s PDF attachments are required", strings.TrimSpace(provider))
}

func (r *Runner) downloadEmailAttachments(ctx context.Context, raw string, allowedHosts []string) ([]emailDocumentAttachment, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var urls []string
	if err := json.Unmarshal([]byte(raw), &urls); err != nil {
		return nil, errors.New("email document_urls_json must be a JSON array")
	}
	if len(urls) > maxEmailDocumentCount {
		return nil, fmt.Errorf("email contains more than %d document attachments", maxEmailDocumentCount)
	}
	hosts := append([]string{"shop.example.test"}, allowedHosts...)
	attachments := make([]emailDocumentAttachment, 0, len(urls))
	totalBytes := 0
	for _, rawURL := range urls {
		source, err := validateMaxDocumentURL(rawURL, hosts, true)
		if err != nil {
			return nil, err
		}
		body, fileName, err := r.downloadMaxDocument(ctx, source)
		if err != nil {
			return nil, err
		}
		totalBytes += len(body)
		if totalBytes > maxEmailDocumentBytes {
			return nil, errors.New("email document attachments exceed 18 MiB")
		}
		attachments = append(attachments, emailDocumentAttachment{FileName: fileName, Body: body})
	}
	return attachments, nil
}

func buildEmailMessageWithAttachments(
	fromLine string,
	recipients []string,
	subject string,
	body string,
	bodyFormat string,
	attachments []emailDocumentAttachment,
) []byte {
	var output bytes.Buffer
	writer := multipart.NewWriter(&output)
	fmt.Fprintf(&output, "From: %s\r\n", fromLine)
	fmt.Fprintf(&output, "To: %s\r\n", strings.Join(recipients, ", "))
	fmt.Fprintf(&output, "Subject: %s\r\n", encodeEmailHeader(subject))
	fmt.Fprintf(&output, "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n", writer.Boundary())

	bodyType := "text/plain; charset=UTF-8"
	if bodyFormat == "html" {
		bodyType = "text/html; charset=UTF-8"
	}
	bodyHeader := textproto.MIMEHeader{}
	bodyHeader.Set("Content-Type", bodyType)
	bodyHeader.Set("Content-Transfer-Encoding", "8bit")
	bodyPart, _ := writer.CreatePart(bodyHeader)
	_, _ = bodyPart.Write([]byte(body))

	for _, attachment := range attachments {
		header := textproto.MIMEHeader{}
		header.Set("Content-Type", "application/pdf")
		header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", safeEmailAttachmentName(attachment.FileName)))
		header.Set("Content-Transfer-Encoding", "base64")
		part, _ := writer.CreatePart(header)
		writeMIMEBase64(part, attachment.Body)
	}
	_ = writer.Close()
	return output.Bytes()
}

func safeEmailAttachmentName(value string) string {
	value = strings.NewReplacer("\r", "", "\n", "", "\\", "-", "/", "-").Replace(strings.TrimSpace(value))
	if value == "" || !strings.HasSuffix(strings.ToLower(value), ".pdf") {
		return "demo-shop-certificate.pdf"
	}
	return value
}

type mimeBase64Writer interface {
	Write([]byte) (int, error)
}

func writeMIMEBase64(writer mimeBase64Writer, body []byte) {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(body)))
	base64.StdEncoding.Encode(encoded, body)
	for len(encoded) > 76 {
		_, _ = writer.Write(encoded[:76])
		_, _ = writer.Write([]byte("\r\n"))
		encoded = encoded[76:]
	}
	if len(encoded) > 0 {
		_, _ = writer.Write(encoded)
		_, _ = writer.Write([]byte("\r\n"))
	}
}
