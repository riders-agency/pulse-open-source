package repository

type ProjectLinkRow struct {
	ProjectID string `json:"project_id"`
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
}
