package runtime

import "strings"

func projectPaymentsURL(projectID string) string {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return publicAppBaseURL() + "/projects"
	}
	return publicAppBaseURL() + "/projects/" + projectID + "/payments"
}

func clientURL(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return publicAppBaseURL() + "/clients"
	}
	return publicAppBaseURL() + "/clients/" + clientID
}

func projectURL(projectID string) string {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return publicAppBaseURL() + "/projects"
	}
	return publicAppBaseURL() + "/projects/" + projectID
}

func splitTelegramHTML(text string, maxRunes int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= maxRunes {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	blocks := strings.Split(text, "\n\n")
	if len(blocks) <= 1 {
		return splitTelegramHTMLHard(runes, maxRunes)
	}
	chunks := make([]string, 0, len(blocks))
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(current.String()))
		current.Reset()
	}
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		extra := block
		if current.Len() > 0 {
			extra = "\n\n" + block
		}
		if len([]rune(current.String()+extra)) > maxRunes {
			flush()
			if len([]rune(block)) > maxRunes {
				chunks = append(chunks, splitTelegramHTMLHard([]rune(block), maxRunes)...)
				continue
			}
			current.WriteString(block)
			continue
		}
		current.WriteString(extra)
	}
	flush()
	if len(chunks) == 0 {
		return splitTelegramHTMLHard(runes, maxRunes)
	}
	return chunks
}

func splitTelegramHTMLHard(runes []rune, maxRunes int) []string {
	if maxRunes <= 0 {
		maxRunes = 3900
	}
	out := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
	}
	return out
}
