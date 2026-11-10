package fio

import (
	"strings"
)

type Result struct {
	UName      string   `json:"u_name"`
	USurname   string   `json:"u_surname"`
	USname     string   `json:"u_sname"`
	Mode       string   `json:"mode"`
	Changed    bool     `json:"changed"`
	Comment    string   `json:"comment"`
	Split      string   `json:"split"`
	FullName   string   `json:"full_name"`
	InputName  string   `json:"input_name"`
	IsValid    bool     `json:"is_valid"`
	Confidence string   `json:"confidence"`
	Warnings   []string `json:"warnings,omitempty"`
}

func prepareName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}

func prepareSearch(raw string) string {
	raw = strings.ToUpper(raw)
	return strings.ReplaceAll(raw, "Ё", "Е")
}

// Normalize splits and corrects FIO using Orion Travel name dictionaries.
func Normalize(raw string) Result {
	input := prepareName(raw)
	if input == "" {
		return Result{
			Mode:       "empty",
			InputName:  input,
			IsValid:    false,
			Confidence: "none",
			Warnings:   []string{"empty fio input"},
		}
	}

	parts := strings.Fields(input)
	res := Result{
		InputName:  input,
		FullName:   input,
		IsValid:    true,
		Confidence: "medium",
	}

	native := struct{ name, second, last string }{}

	switch len(parts) {
	case 3:
		normalizeThreeParts(&res, parts)
		res.FullName = joinNameParts(res.UName, res.USname, res.USurname)
		return res
	case 2:
		native.name, native.last = parts[0], parts[1]
		res.UName, res.USurname = native.name, native.last
		res.Mode = "fio_normalized"
		res.Split = "ИФ в имени"
	case 1:
		res.UName = input
		res.Mode = "single_name"
		res.Split = "Одно имя"
		res.Confidence = "low"
		res.Warnings = append(res.Warnings, "surname is missing")
		return res
	default:
		res.UName = input
		res.Mode = "single_name"
		res.Split = "Много слов как одно имя"
		res.Confidence = "low"
		res.Warnings = append(res.Warnings, "too many fio parts")
		return res
	}

	applyCorrections(&res, native)
	res.FullName = joinNameParts(res.UName, res.USname, res.USurname)
	return res
}

func joinNameParts(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, " ")
}

func normalizeThreeParts(res *Result, parts []string) {
	first, second, third := parts[0], parts[1], parts[2]
	res.Mode = "fio_normalized"

	switch {
	case isLastNameLike(first) && (isFirstName(second) || isPatronymicLike(third)):
		res.UName, res.USname, res.USurname = second, third, first
		res.Split = "ФИО в имени"
		res.Changed = true
		res.Comment = "Фамилия в начале, имя и отчество после нее"
		res.Confidence = confidenceFor(isLastName(first), isFirstName(second), isPatronymicLike(third))
	case (isFirstName(first) || isPatronymicLike(second)) && isLastNameLike(third):
		res.UName, res.USname, res.USurname = first, second, third
		res.Split = "ИОФ в имени"
		res.Confidence = confidenceFor(isFirstName(first), isPatronymicLike(second), isLastName(third))
	case isLastNameLike(first) && !isLastNameLike(third):
		res.UName, res.USname, res.USurname = second, third, first
		res.Split = "ФИО в имени"
		res.Changed = true
		res.Comment = "Похоже, фамилия стоит первой"
		res.Confidence = "medium"
	default:
		res.UName, res.USname, res.USurname = first, second, third
		res.Split = "ИОФ в имени"
		res.Confidence = "low"
		res.Warnings = append(res.Warnings, "ambiguous three-part fio order")
	}
}

func applyCorrections(res *Result, native struct{ name, second, last string }) {
	name := res.UName
	second := res.USname
	last := res.USurname

	if isFirstName(last) && isLastName(name) {
		res.UName, res.USurname = native.last, native.name
		res.Changed = true
		res.Comment = "Имя на месте фамилии, фамилия на месте имени"
		res.Confidence = "high"
		return
	}
	if isFirstName(last) && isLastNameLike(name) {
		res.UName, res.USurname = native.last, native.name
		res.Changed = true
		res.Comment = "Имя на месте фамилии, фамилия похожа на первую часть"
		res.Confidence = "medium"
		return
	}
	if isLastNameLike(name) && !isLastNameLike(last) && !isFirstName(name) {
		res.UName, res.USurname = native.last, native.name
		res.Changed = true
		res.Comment = "Похоже, фамилия стоит первой"
		res.Confidence = "medium"
		return
	}
	if isFirstName(name) && isLastNameLike(last) {
		res.Confidence = confidenceFor(isFirstName(name), isLastName(last))
		return
	}
	if isFirstName(last) && name == "" {
		res.UName, res.USurname = native.last, native.name
		res.Changed = true
		res.Comment = "Имя на месте фамилии, имя пустое"
		res.Confidence = "medium"
		return
	}
	if isLastName(name) && last == "" {
		res.UName, res.USurname = native.last, native.name
		res.Changed = true
		res.Comment = "Фамилия на месте имени, имя пустое"
		res.Confidence = "medium"
		return
	}
	if isLastName(name) && isFirstName(second) {
		res.UName, res.USname, res.USurname = native.second, native.last, native.name
		res.Changed = true
		res.Comment = "Фамилия на месте имени, имя на месте отчества"
		res.Confidence = "high"
		return
	}
	if isLastName(name) && isFirstName(second) && last != "" {
		res.UName, res.USname, res.USurname = native.second, native.last, native.name
		res.Changed = true
		res.Comment = "Фамилия на месте имени, имя на месте отчества, отчество на месте фамилии"
		res.Confidence = "high"
		return
	}
	res.Confidence = "low"
	res.Warnings = append(res.Warnings, "ambiguous two-part fio order")
}

func isLastNameLike(value string) bool {
	if isLastName(value) {
		return true
	}
	s := prepareSearch(strings.TrimSpace(value))
	for _, suffix := range []string{
		"ОВ", "ЕВ", "ЕВА", "ОВА", "ИН", "ИНА", "ЫН", "ЫНА",
		"СКИЙ", "СКАЯ", "ЦКИЙ", "ЦКАЯ", "КО", "ЕНКО", "УК", "ЮК",
		"ИЧ", "ВИЧ", "ОВИЧ", "ЕВИЧ",
	} {
		if strings.HasSuffix(s, suffix) && len([]rune(s)) > len([]rune(suffix))+1 {
			return true
		}
	}
	return false
}

func isPatronymicLike(value string) bool {
	s := prepareSearch(strings.TrimSpace(value))
	for _, suffix := range []string{"ВИЧ", "ВНА", "ИЧНА", "ОВИЧ", "ЕВИЧ", "ОВНА", "ЕВНА", "ОГЛЫ", "КЫЗЫ"} {
		if strings.HasSuffix(s, suffix) && len([]rune(s)) > len([]rune(suffix))+1 {
			return true
		}
	}
	return false
}

func confidenceFor(signals ...bool) string {
	count := 0
	for _, ok := range signals {
		if ok {
			count++
		}
	}
	if count >= 2 {
		return "high"
	}
	if count == 1 {
		return "medium"
	}
	return "low"
}
