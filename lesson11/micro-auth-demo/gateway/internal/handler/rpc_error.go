package handler

import "strings"

func statusFromMessage(message string) int {
	switch {
	case strings.Contains(message, "unauthorized:"):
		return 401
	case strings.Contains(message, "forbidden:"):
		return 403
	case strings.Contains(message, "not_found:"):
		return 404
	default:
		return 500
	}
}

func presentableMessage(message string) string {
	for _, prefix := range []string{"unauthorized: ", "forbidden: ", "not_found: "} {
		if idx := strings.Index(message, prefix); idx >= 0 {
			return message[idx+len(prefix):]
		}
	}
	return message
}
