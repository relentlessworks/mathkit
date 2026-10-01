package api

import (
	"fmt"
	"net/http"
	"strings"
)

// wantsJSON checks if the client wants JSON responses.
func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	return false
}

// writeText writes a plain text response.
func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

// writeError writes an error response in the appropriate format.
func writeError(w http.ResponseWriter, r *http.Request, status int, msg, hint string) {
	if wantsJSON(r) {
		body := fmt.Sprintf(`{"error":"%s","hint":"%s"}`, escapeJSON(msg), escapeJSON(hint))
		writeJSON(w, status, body)
		return
	}
	writeText(w, status, fmt.Sprintf("error: %s | hint: %s", msg, hint))
}

// writeOK writes a success response in the appropriate format.
func writeOK(w http.ResponseWriter, r *http.Request, text, json string) {
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json)
		return
	}
	writeText(w, http.StatusOK, text)
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
