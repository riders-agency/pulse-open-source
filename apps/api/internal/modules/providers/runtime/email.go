package runtime

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	emailprovider "riders.agency/pulse/api/internal/modules/providers/email"
)

func (r *Runner) executeEmail(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	if strings.TrimSpace(node.Config.SMTPActionDependencyID) == "" {
		err := errors.New("smtp integration is not selected")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	dep, err := r.loadDependency(ctx, node.Config.SMTPActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "smtp.server" {
		err := fmt.Errorf("email.send requires smtp.server dependency")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "error": err.Error()}, false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err := fmt.Errorf("smtp dependency is %s", dep.Status)
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "status": dep.Status}, false, "dependency_status", err)
	}
	host := strings.TrimSpace(stringMetadata(dep.Metadata, "host"))
	if host == "" {
		perr := errors.New("SMTP host in metadata_json is required")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": perr.Error()}, false, "validation", perr)
	}
	port := intFromMetadata(dep.Metadata, "port", 587)
	if port <= 0 || port > 65535 {
		perr := errors.New("SMTP port metadata invalid")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": perr.Error()}, false, "validation", perr)
	}
	username := strings.TrimSpace(stringMetadata(dep.Metadata, "username"))
	fromEmail := strings.TrimSpace(stringMetadata(dep.Metadata, "from_email"))
	fromName := strings.TrimSpace(stringMetadata(dep.Metadata, "from_name"))
	if fromEmail == "" {
		perr := errors.New("from_email metadata is required")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": perr.Error()}, false, "validation", perr)
	}
	subjectTemplate := node.Config.Subject
	bodyTemplate := node.Config.Body
	bodyFormat := strings.TrimSpace(node.Config.BodyFormat)
	if (strings.TrimSpace(subjectTemplate) == "" || strings.TrimSpace(bodyTemplate) == "") && strings.TrimSpace(node.Config.EmailTemplateID) != "" {
		if subject, body, format, terr := r.loadEmailTemplate(ctx, node.Config.EmailTemplateID); terr == nil {
			if strings.TrimSpace(subjectTemplate) == "" {
				subjectTemplate = subject
			}
			if strings.TrimSpace(bodyTemplate) == "" {
				bodyTemplate = body
			}
			if bodyFormat == "" {
				bodyFormat = format
			}
		}
	}

	toRaw := strings.TrimSpace(renderTemplate(node.Config.To, runInput, steps, tpl))
	rcpt := emailRecipients(toRaw)
	if len(rcpt) == 0 {
		message := "email to is empty"
		if strings.Contains(node.Config.To, "client.primary_contact.email") {
			message = "у клиента нет email ответственного заказчика. Добавьте активный контакт с email в карточке клиента"
		}
		perr := errors.New(message)
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": perr.Error()}, false, "validation", perr)
	}
	subject := strings.TrimSpace(renderTemplate(subjectTemplate, runInput, steps, tpl))
	if subject == "" {
		perr := errors.New("email subject is required")
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": perr.Error()}, false, "validation", perr)
	}
	body := renderTemplate(bodyTemplate, runInput, steps, tpl)
	if strings.EqualFold(bodyFormat, "html") || strings.Contains(strings.ToLower(body), "<html") {
		bodyFormat = "html"
	}

	encryption := strings.TrimSpace(stringMetadata(dep.Metadata, "encryption"))
	fromLine, err := formatEmailFromDisplay(fromName, fromEmail)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "error": err.Error()}, false, "validation", err)
	}
	var msg []byte
	attachmentURLsJSON := strings.TrimSpace(renderTemplate(node.Config.DocumentURLsJSON, runInput, steps, tpl))
	attachments, attachmentErr := r.downloadEmailAttachments(ctx, attachmentURLsJSON, node.Config.DocumentAllowedHosts)
	if attachmentErr != nil {
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "error": attachmentErr.Error()}, true, "document_download", attachmentErr)
	}
	if attachmentErr = requireDocumentAttachments("email", node.Config.DocumentRequired, len(attachments)); attachmentErr != nil {
		return failedAction(input, "{}", map[string]any{"provider": "smtp", "ok": false, "error": attachmentErr.Error()}, false, "validation", attachmentErr)
	}
	if len(attachments) > 0 {
		msg = buildEmailMessageWithAttachments(fromLine, rcpt, subject, body, bodyFormat, attachments)
	} else if bodyFormat == "html" {
		msg = buildHTMLEmailMessage(fromLine, rcpt, subject, body)
	} else {
		msg = buildPlainEmailMessage(fromLine, rcpt, subject, body)
	}

	reqMeta, _ := json.Marshal(map[string]any{
		"to":                rcpt,
		"subject":           subject,
		"body":              body,
		"host":              host,
		"port":              port,
		"run_url":           tpl.RunURL,
		"system":            tpl.Map(),
		"template":          map[string]any{"to": node.Config.To, "subject": subjectTemplate, "body": bodyTemplate, "email_template_id": node.Config.EmailTemplateID},
		"success_criteria":  node.Config.SuccessCriteria,
		"attachments_count": len(attachments),
	})
	providerMap, sErr := smtpDialSendDetailed(ctx, host, port, encryption, username, strings.TrimSpace(dep.Plaintext), fromEmail, rcpt, msg)
	providerJSON, _ := json.Marshal(providerMap)
	if sErr != nil {
		retryable := emailprovider.IsRetryableSMTPFailure(providerMap, sErr)
		report := diagnosticsForFailure(node.ActionTypeKey, providerMap, retryable, "smtp", sErr)
		output := providerdiagnostics.OutputWithReport("{}", report)
		return actionExecution{InputJSON: string(reqMeta), OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "smtp", Err: sErr}
	}

	delivered := map[string]any{
		"delivered":          true,
		"provider":           "smtp",
		"provider_ok":        true,
		"recipients_count":   len(rcpt),
		"attachments_count":  len(attachments),
		"documents_attached": len(attachments) > 0,
	}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(reqMeta), providerMap, delivered, smtpExpose(len(rcpt), host, port))
}

func smtpExpose(recipientCount int, host string, port int) map[string]any {
	return map[string]any{"recipient_count": recipientCount, "smtp_host": host, "smtp_port": port}
}

func intFromMetadata(metadata map[string]any, key string, def int) int {
	if metadata == nil {
		return def
	}
	v, exists := metadata[key]
	if !exists || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return def
	}
}

func formatEmailFromDisplay(name, addr string) (string, error) {
	name = strings.TrimSpace(name)
	addr = strings.TrimSpace(addr)
	if strings.ContainsAny(name, "\r\n") || strings.ContainsAny(addr, "\r\n") {
		return "", errors.New("SMTP sender contains a line break")
	}
	parsed, err := mail.ParseAddress(addr)
	if err != nil || parsed.Address != addr {
		return "", errors.New("SMTP from_email metadata is invalid")
	}
	parsed.Name = name
	return parsed.String(), nil
}

func buildPlainEmailMessage(fromLine string, recipients []string, subject, plainBody string) []byte {
	b := strings.Builder{}
	fmt.Fprintf(&b, "From: %s\r\n", fromLine)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(recipients, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", encodeEmailHeader(subject))
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(plainBody)
	return []byte(b.String())
}

func buildHTMLEmailMessage(fromLine string, recipients []string, subject, htmlBody string) []byte {
	b := strings.Builder{}
	fmt.Fprintf(&b, "From: %s\r\n", fromLine)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(recipients, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", encodeEmailHeader(subject))
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(htmlBody)
	return []byte(b.String())
}

func encodeEmailHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if utf8.RuneCountInString(value) == len(value) {
		return value
	}
	return mime.QEncoding.Encode("UTF-8", value)
}

func emailRecipients(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func smtpDialSendDetailed(ctx context.Context, host string, port int, encrypt, username, password, envelopeFrom string, rcpt []string, msg []byte) (map[string]any, error) {
	stages := []string{}
	fail := func(stage string, err error) (map[string]any, error) {
		st := append(append([]string{}, stages...), stage)
		code, enhanced := emailprovider.ExtractSMTPStatus(err.Error())
		env := map[string]any{"accepted": false, "stages": st, "failed_stage": stage, "last_server_reply": err.Error()}
		if code != "" {
			env["smtp_code"] = code
		}
		if enhanced != "" {
			env["enhanced_status_code"] = enhanced
		}
		return map[string]any{"provider": "smtp", "ok": false, "smtp": env, "error": err.Error(), "provider_response_redacted": true}, err
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	mode := strings.ToLower(strings.TrimSpace(encrypt))
	if mode == "" {
		if port == 465 {
			mode = "tls"
		} else {
			mode = "starttls"
		}
	}
	dialer := &net.Dialer{Timeout: 45 * time.Second}
	dialCtx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	var tcp net.Conn
	var err error
	switch mode {
	case "tls":
		tcp, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
	case "none":
		tcp, err = dialer.DialContext(dialCtx, "tcp", addr)
	default:
		tcp, err = dialer.DialContext(dialCtx, "tcp", addr)
	}
	if err != nil {
		return fail("tcp_dial", err)
	}
	defer tcp.Close()
	stages = append(stages, "tcp_connected")

	client, err := smtp.NewClient(tcp, host)
	if err != nil {
		return fail("smtp_client", err)
	}
	defer client.Close()
	stages = append(stages, "smtp_banner")
	if mode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return fail("starttls", err)
			}
			stages = append(stages, "starttls")
		}
	}
	if strings.TrimSpace(username) != "" {
		auth := emailprovider.NewSMTPAuth(username, password, host)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fail("auth", err)
			}
			stages = append(stages, "auth")
		}
	}
	if err := client.Mail(envelopeFrom); err != nil {
		return fail("mail_from", err)
	}
	stages = append(stages, "mail_from")
	for _, r := range rcpt {
		if err := client.Rcpt(r); err != nil {
			return fail("rcpt_to", err)
		}
	}
	stages = append(stages, "rcpt")
	w, err := client.Data()
	if err != nil {
		return fail("data", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fail("data_write", err)
	}
	if err := w.Close(); err != nil {
		return fail("data_close", err)
	}
	stages = append(stages, "data_sent")
	if err := client.Quit(); err != nil {
		return fail("quit", err)
	}
	stages = append(stages, "quit")
	return map[string]any{
		"provider":                   "smtp",
		"ok":                         true,
		"provider_response_redacted": true,
		"smtp": map[string]any{
			"accepted":          true,
			"stages":            append([]string(nil), stages...),
			"last_server_reply": "session completed",
		},
	}, nil
}
