package ai

import (
	"strings"
	"unicode/utf8"
)

const fallbackDescription = "A user-uploaded image was processed successfully, but only a brief safe description is available."

type GuardrailsConfig struct {
	Enabled             bool
	MaxDescriptionChars int
}

func applyDescriptionGuardrails(raw string, cfg GuardrailsConfig) string {
	cleaned := strings.TrimSpace(raw)
	if !cfg.Enabled {
		return trimToLimit(normalizeWhitespace(cleaned), cfg.MaxDescriptionChars)
	}

	cleaned = normalizeWhitespace(cleaned)
	if cleaned == "" {
		return fallbackDescription
	}

	lower := strings.ToLower(cleaned)
	blockedPhrases := []string{
		"as an ai",
		"i can't",
		"i cannot",
		"i can’t",
		"sorry,",
		"sorry but",
		"i am unable",
		"cannot help with",
		"prompt",
		"policy",
	}
	for _, phrase := range blockedPhrases {
		if strings.Contains(lower, phrase) {
			return fallbackDescription
		}
	}

	return trimToLimit(cleaned, cfg.MaxDescriptionChars)
}

func normalizeWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func trimToLimit(input string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(input) <= limit {
		return input
	}

	runes := []rune(input)
	trimmed := strings.TrimSpace(string(runes[:limit]))
	trimmed = strings.TrimRight(trimmed, ",;:-")
	if trimmed == "" {
		return fallbackDescription
	}
	return trimmed + "…"
}
