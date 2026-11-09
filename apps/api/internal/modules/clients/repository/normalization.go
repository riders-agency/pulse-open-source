package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func normalizeClientInput(in ClientInput) ClientInput {
	in.ClientType = normalizeClientType(in.ClientType)
	in.Status = normalizeDefault(in.Status, "active")
	in.Name = strings.TrimSpace(firstNonEmpty(in.Name, in.ShortName, in.FullName))
	in.ShortName = strings.TrimSpace(in.ShortName)
	in.FullName = strings.TrimSpace(in.FullName)
	in.INN = normalizeDigits(in.INN)
	in.KPP = normalizeDigits(in.KPP)
	in.OGRN = normalizeDigits(in.OGRN)
	in.OGRNIP = normalizeDigits(in.OGRNIP)
	in.LegalAddress = strings.TrimSpace(in.LegalAddress)
	in.ActualAddress = strings.TrimSpace(in.ActualAddress)
	in.OKVED = strings.TrimSpace(in.OKVED)
	in.Comment = strings.TrimSpace(in.Comment)
	if in.Name == "" {
		in.Name = "Новый клиент"
	}
	return in
}

func normalizeCompanyLegalEntityInput(in CompanyLegalEntityInput) CompanyLegalEntityInput {
	in.EntityType = normalizeLegalEntityType(in.EntityType)
	in.Status = normalizeDefault(in.Status, "active")
	in.Name = strings.TrimSpace(firstNonEmpty(in.Name, in.ShortName, in.FullName))
	in.ShortName = strings.TrimSpace(in.ShortName)
	in.FullName = strings.TrimSpace(in.FullName)
	in.INN = normalizeDigits(in.INN)
	in.KPP = normalizeDigits(in.KPP)
	in.OGRN = normalizeDigits(in.OGRN)
	in.OGRNIP = normalizeDigits(in.OGRNIP)
	in.LegalAddress = strings.TrimSpace(in.LegalAddress)
	in.ActualAddress = strings.TrimSpace(in.ActualAddress)
	in.OKVED = strings.TrimSpace(in.OKVED)
	in.Comment = strings.TrimSpace(in.Comment)
	if in.Name == "" {
		in.Name = "Наше юрлицо"
	}
	return in
}

func normalizeClientType(value string) string {
	switch strings.TrimSpace(value) {
	case "individual":
		return "individual"
	case "individual_entrepreneur":
		return "individual_entrepreneur"
	default:
		return "legal_entity"
	}
}

func normalizeLegalEntityType(value string) string {
	if strings.TrimSpace(value) == "individual_entrepreneur" {
		return "individual_entrepreneur"
	}
	return "legal_entity"
}

func normalizeResponsibleRole(value string) string {
	switch strings.TrimSpace(value) {
	case "technical_lead", "finance", "legal", "owner", "other":
		return strings.TrimSpace(value)
	default:
		return "account_manager"
	}
}

func normalizeServiceInput(in ServiceInput) ServiceInput {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = "Услуга"
	}
	in.Status = normalizeDefault(in.Status, "active")
	in.Currency = normalizeDefault(in.Currency, "RUB")
	in.BillingPeriod = normalizeDefault(in.BillingPeriod, "monthly")
	in.NextDueDate = strings.TrimSpace(in.NextDueDate)
	if in.RemindDaysBefore == 0 {
		in.RemindDaysBefore = 7
	}
	if in.RemindEveryDays == 0 {
		in.RemindEveryDays = 7
	}
	return in
}

func normalizeServiceTypeInput(in ServiceTypeInput) ServiceTypeInput {
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if in.DisplayName == "" {
		in.DisplayName = "Тип услуги"
	}
	in.ServiceTypeKey = normalizeServiceTypeKey(in.ServiceTypeKey)
	if in.ServiceTypeKey == "" {
		in.ServiceTypeKey = normalizeServiceTypeKey(in.DisplayName)
	}
	if in.ServiceTypeKey == "" {
		in.ServiceTypeKey = "service_" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	}
	in.Status = normalizeServiceTypeStatus(in.Status)
	if in.SortOrder <= 0 {
		in.SortOrder = 100
	}
	in.Description = strings.TrimSpace(in.Description)
	return in
}

func normalizeServiceTypeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastUnderscore := false
	for _, ch := range value {
		ok := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if ok {
			b.WriteRune(ch)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore && b.Len() > 0 {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func normalizeServiceTypeStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "archived":
		return "archived"
	default:
		return "active"
	}
}

func ValidateClientIdentifiers(clientType, inn, kpp, ogrn, ogrnip string) error {
	clientType = normalizeClientType(clientType)
	inn = normalizeDigits(inn)
	kpp = normalizeDigits(kpp)
	ogrn = normalizeDigits(ogrn)
	ogrnip = normalizeDigits(ogrnip)
	switch clientType {
	case "legal_entity":
		if inn != "" && len(inn) != 10 {
			return fmt.Errorf("%w: legal entity INN must contain 10 digits", ErrInvalidIdentifiers)
		}
		if kpp != "" && len(kpp) != 9 {
			return fmt.Errorf("%w: KPP must contain 9 digits", ErrInvalidIdentifiers)
		}
		if ogrn != "" && len(ogrn) != 13 {
			return fmt.Errorf("%w: OGRN must contain 13 digits", ErrInvalidIdentifiers)
		}
	case "individual_entrepreneur":
		if inn != "" && len(inn) != 12 {
			return fmt.Errorf("%w: individual entrepreneur INN must contain 12 digits", ErrInvalidIdentifiers)
		}
		if ogrnip != "" && len(ogrnip) != 15 {
			return fmt.Errorf("%w: OGRNIP must contain 15 digits", ErrInvalidIdentifiers)
		}
	case "individual":
		if inn != "" && len(inn) != 12 {
			return fmt.Errorf("%w: individual INN must contain 12 digits", ErrInvalidIdentifiers)
		}
	}
	return nil
}
