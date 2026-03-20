package aisuites

import (
	"reflect"
	"strings"
)

const (
	langEnglish         = "Respond entirely in English."
	langBahasaIndonesia = "Respond entirely in Bahasa Indonesia."
	langJapanese        = "Respond entirely in Japanese."
	langChinese         = "Respond entirely in Chinese."
	langKorean          = "Respond entirely in Korean."
)

// toInt64 converts various numeric types to int64.
func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int64:
		return t, true
	case float64:
		return int64(t), true
	case int32:
		return int64(t), true
	}
	return 0, false
}

// extractTotalTokens tries to get TotalTokens from GenerationInfo.
// Supports:
//   - Flat map keys: "TotalTokens" (OpenAI, Google)
//   - Nested struct: info["usage"].TotalTokens (Mistral: UsageInfo struct)
//
//nolint:gocognit,nestif // extraction logic handles complex nested map structures output by LLMs
func extractTotalTokens(info map[string]any) int64 {
	// 1. Try flat key (OpenAI, Google, etc.)
	if v, ok := info["TotalTokens"]; ok {
		if n, ok := toInt64(v); ok {
			return n
		}
	}

	// 2. Try nested "usage" (Mistral SDK stores UsageInfo struct)
	if usage, ok := info["usage"]; ok && usage != nil {
		// If it's a map (JSON-like)
		if m, ok := usage.(map[string]any); ok {
			if v, ok := m["total_tokens"]; ok {
				if n, ok := toInt64(v); ok {
					return n
				}
			}
			if v, ok := m["TotalTokens"]; ok {
				if n, ok := toInt64(v); ok {
					return n
				}
			}
		}
		// If it's a struct (e.g. sdk.UsageInfo), use reflection
		rv := reflect.ValueOf(usage)
		if rv.Kind() == reflect.Struct {
			f := rv.FieldByName("TotalTokens")
			if f.IsValid() && f.CanInterface() {
				if n, ok := toInt64(f.Interface()); ok {
					return n
				}
			}
		}
	}

	return 0
}

// truncateString helper ensures strings do not exceed maxRunes to prevent HTTP 400 Payload Too Large errors.
func truncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "...\n[Content truncated due to length limits]"
	}
	return s
}

func stripMarkdownFences(s string) string {
	cleaned := strings.TrimSpace(s)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Robustly extract only the JSON part if there's extra text
	// Find the first occurrence of { or [
	start := -1
	var opener rune
	var closer rune

	for i, r := range cleaned {
		if r == '{' {
			start = i
			opener = '{'
			closer = '}'
			break
		}
		if r == '[' {
			start = i
			opener = '['
			closer = ']'
			break
		}
	}

	if start != -1 {
		cleaned = extractJSONBlock(cleaned, start, opener, closer)
	}

	return strings.TrimSpace(cleaned)
}

func extractJSONBlock(cleaned string, start int, opener, closer rune) string {
	count := 0
	end := -1
	inString := false
	escaped := false

	for i := start; i < len(cleaned); i++ {
		char := rune(cleaned[i])

		if inString {
			if char == '\\' && !escaped {
				escaped = true
			} else {
				if char == '"' && !escaped {
					inString = false
				}
				escaped = false
			}
			continue
		}

		if char == '"' {
			inString = true
			continue
		}

		if char == opener {
			count++
		} else if char == closer {
			count--
			if count == 0 {
				end = i
				break
			}
		}
	}

	if end != -1 {
		return cleaned[start : end+1]
	}
	return cleaned
}
