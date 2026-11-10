package repository

type ContactRow struct {
	ClientContactID  string `json:"client_contact_id"`
	ClientID         string `json:"client_id"`
	FirstName        string `json:"first_name,omitempty"`
	LastName         string `json:"last_name,omitempty"`
	MiddleName       string `json:"middle_name,omitempty"`
	Position         string `json:"position,omitempty"`
	Email            string `json:"email,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Telegram         string `json:"telegram,omitempty"`
	PreferredChannel string `json:"preferred_channel"`
	Comment          string `json:"comment,omitempty"`
	IsPrimary        bool   `json:"is_primary"`
	Status           string `json:"status"`
}

type ContactInput struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	MiddleName       string `json:"middle_name"`
	Position         string `json:"position"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	Telegram         string `json:"telegram"`
	PreferredChannel string `json:"preferred_channel"`
	Comment          string `json:"comment"`
	IsPrimary        bool   `json:"is_primary"`
	Status           string `json:"status"`
}

type ContactPatch struct {
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	MiddleName       *string `json:"middle_name"`
	Position         *string `json:"position"`
	Email            *string `json:"email"`
	Phone            *string `json:"phone"`
	Telegram         *string `json:"telegram"`
	PreferredChannel *string `json:"preferred_channel"`
	Comment          *string `json:"comment"`
	IsPrimary        *bool   `json:"is_primary"`
	Status           *string `json:"status"`
}
