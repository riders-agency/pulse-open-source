package http

import (
	"github.com/go-chi/chi/v5"
	clientssvc "riders.agency/pulse/api/internal/modules/clients/service"
)

type RunNotifier interface {
	Notify()
}

func MountCompanyRoutes(r chi.Router, clients *clientssvc.Commands, dadata DaDataService) {
	r.Get("/companies/{company_id}/clients", List(clients))
	r.Post("/companies/{company_id}/clients", Create(clients))
	r.Post("/companies/{company_id}/clients/dadata/find-party", FindDaDataParty(dadata))
	r.Get("/companies/{company_id}/integrations/dadata", GetDaDataIntegration(clients))
	r.Put("/companies/{company_id}/integrations/dadata", SaveDaDataIntegration(dadata))
	r.Get("/companies/{company_id}/legal-entities", ListCompanyLegalEntities(clients))
	r.Post("/companies/{company_id}/legal-entities", CreateCompanyLegalEntity(clients))
}

func MountGlobalRoutes(r chi.Router, clients *clientssvc.Commands, engine RunNotifier) {
	r.Get("/client-service-types", ListServiceTypes(clients))
	r.Get("/company-legal-entities/{company_legal_entity_id}", GetCompanyLegalEntity(clients))
	r.Patch("/company-legal-entities/{company_legal_entity_id}", PatchCompanyLegalEntity(clients))
	r.Delete("/company-legal-entities/{company_legal_entity_id}", DeleteCompanyLegalEntity(clients))

	r.Get("/client-services/{client_service_id}", GetService(clients))
	r.Patch("/client-services/{client_service_id}", PatchService(clients))
	r.Delete("/client-services/{client_service_id}", DeleteService(clients))
	r.Get("/payment-actions/{payment_action_id}", GetPaymentAction(clients))
	r.Post("/payment-actions/{payment_action_id}/confirm", ConfirmPaymentAction(clients, engine))
	r.Get("/payment-actions/{payment_action_id}/documents", ListPaymentActionDocuments(clients))
	r.Post("/payment-actions/{payment_action_id}/documents", UploadPaymentActionDocument(clients))
	r.Get("/payment-action-documents/{payment_action_document_id}", DownloadPaymentActionDocument(clients))
	r.Post("/payment-actions/{payment_action_id}/snooze", SnoozePaymentAction(clients))
}

func MountClientRoutes(r chi.Router, clients *clientssvc.Commands, engine RunNotifier) {
	r.Route("/clients/{client_id}", func(cr chi.Router) {
		cr.Get("/", Get(clients))
		cr.Patch("/", Patch(clients))
		cr.Delete("/", Delete(clients))
		cr.Get("/responsibles", ListResponsibles(clients))
		cr.Post("/responsibles", CreateResponsible(clients))
		cr.Patch("/responsibles/{client_responsible_id}", PatchResponsible(clients))
		cr.Delete("/responsibles/{client_responsible_id}", DeleteResponsible(clients))
		cr.Get("/contacts", ListContacts(clients))
		cr.Post("/contacts", CreateContact(clients))
		cr.Patch("/contacts/{client_contact_id}", PatchContact(clients))
		cr.Delete("/contacts/{client_contact_id}", DeleteContact(clients))
		cr.Get("/projects", ListProjects(clients))
		cr.Post("/projects/{project_id}", LinkProject(clients))
		cr.Delete("/projects/{project_id}", UnlinkProject(clients))
		cr.Get("/services", ListServices(clients))
		cr.Post("/services", CreateService(clients))
		cr.Get("/payment-actions", ClientPaymentActions(clients))
		cr.Post("/payments/manual", RecordManualPayment(clients, engine))
	})
}

func MountProjectRoutes(r chi.Router, clients *clientssvc.Commands, dadata DaDataIntegrationSaver, engine RunNotifier) {
	r.Get("/client-service-types", ListProjectServiceTypes(clients))
	r.Post("/client-service-types", CreateProjectServiceType(clients))
	r.Patch("/client-service-types/{client_service_type_id}", PatchProjectServiceType(clients))
	r.Delete("/client-service-types/{client_service_type_id}", ArchiveProjectServiceType(clients))
	r.Get("/payment-actions", ProjectPaymentActions(clients))
	r.Post("/payment-actions/manual", CreateManualPaymentAction(clients, engine))
	r.Get("/payment-schedule", ProjectPaymentSchedule(clients))
	r.Get("/integrations/dadata", GetProjectDaDataIntegration(clients))
	r.Put("/integrations/dadata", SaveProjectDaDataIntegration(dadata))
}
