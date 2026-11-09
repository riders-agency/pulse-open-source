package http

import (
	"github.com/go-chi/chi/v5"

	tplrepo "riders.agency/pulse/api/internal/modules/templates/repository"
)

func MountProjectRoutes(r chi.Router, email *tplrepo.EmailTemplates, message *tplrepo.MessageTemplates) {
	r.Get("/email-templates", ListEmail(email))
	r.Post("/email-templates", CreateEmail(email))
	r.Get("/email-templates/{email_template_id}", GetEmail(email))
	r.Patch("/email-templates/{email_template_id}", PatchEmail(email))
	r.Delete("/email-templates/{email_template_id}", DeleteEmail(email))

	r.Get("/message-templates", ListMessage(message))
	r.Post("/message-templates", CreateMessage(message))
	r.Get("/message-templates/{message_template_id}", GetMessage(message))
	r.Patch("/message-templates/{message_template_id}", PatchMessage(message))
	r.Delete("/message-templates/{message_template_id}", DeleteMessage(message))
}
