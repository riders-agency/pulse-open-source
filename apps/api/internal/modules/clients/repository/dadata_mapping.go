package repository

import (
	"fmt"
	"strconv"
	"strings"
)

func MapDaDataSuggestions(raw map[string]any) []map[string]any {
	items, _ := raw["suggestions"].([]any)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		obj, _ := item.(map[string]any)
		if obj == nil {
			continue
		}
		data, _ := obj["data"].(map[string]any)
		name, _ := data["name"].(map[string]any)
		address, _ := data["address"].(map[string]any)
		state, _ := data["state"].(map[string]any)
		inn := stringMapValue(data, "inn")
		kpp := stringMapValue(data, "kpp")
		ogrn := stringMapValue(data, "ogrn")
		clientType := "legal_entity"
		ogrnip := ""
		if strings.EqualFold(stringMapValue(data, "type"), "INDIVIDUAL") || (kpp == "" && len(ogrn) == 15) {
			clientType = "individual_entrepreneur"
			ogrnip = ogrn
			ogrn = ""
		}
		shortName := firstNonEmpty(stringMapValue(name, "short_with_opf"), stringMapValue(name, "short"), stringMapValue(obj, "value"))
		fullName := firstNonEmpty(stringMapValue(name, "full_with_opf"), stringMapValue(name, "full"), stringMapValue(obj, "unrestricted_value"))
		legalAddress := firstNonEmpty(stringMapValue(address, "unrestricted_value"), stringMapValue(address, "value"))
		out = append(out, map[string]any{
			"client_type":     clientType,
			"name":            firstNonEmpty(shortName, fullName, stringMapValue(obj, "value")),
			"short_name":      shortName,
			"full_name":       fullName,
			"inn":             inn,
			"kpp":             kpp,
			"ogrn":            ogrn,
			"ogrnip":          ogrnip,
			"legal_address":   legalAddress,
			"actual_address":  legalAddress,
			"okved":           stringMapValue(data, "okved"),
			"dadata_status":   stringMapValue(state, "status"),
			"dadata_raw_json": obj,
		})
	}
	return out
}

func stringMapValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	switch v := m[key].(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
