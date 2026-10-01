package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/relentlessworks/mathkit/internal/model"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	auth   *Auth
	config *Config
}

// Config is a minimal config interface for the handler.
type Config struct {
	Degree bool
	NoAuth bool
}

// NewHandler creates a new API handler.
func NewHandler(auth *Auth, degree, noAuth bool) *Handler {
	return &Handler{
		auth: auth,
		config: &Config{Degree: degree, NoAuth: noAuth},
	}
}

// Routes returns the HTTP mux with all routes registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/health", h.health)

	// Auth endpoints
	mux.HandleFunc("/auth/request", h.authRequest)
	mux.HandleFunc("/auth/verify", h.authVerify)

	// Core endpoints
	mux.HandleFunc("/eval", h.eval)
	mux.HandleFunc("/eval/batch", h.evalBatch)
	mux.HandleFunc("/functions", h.functions)
	mux.HandleFunc("/constants", h.constants)
	mux.HandleFunc("/operators", h.operators)

	// MCP endpoint
	mux.HandleFunc("/mcp", h.mcp)

	// Root
	mux.HandleFunc("/", h.root)

	return mux
}

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, r, http.StatusNotFound,
			fmt.Sprintf("unknown endpoint: %s", r.URL.Path),
			"call GET /help for the full list of available endpoints")
		return
	}
	writeOK(w, r,
		"mathkit — agentic-first math expression evaluation service\nCall GET /help for the operating manual.",
		`{"service":"mathkit","version":"0.1.0","endpoints":["/help","/eval","/eval/batch","/functions","/constants","/operators","/mcp"]}`)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeOK(w, r, "ok", `{"status":"ok"}`)
}

func (h *Handler) authRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed,
			"this endpoint requires POST",
			"use POST /auth/request with a body like: email=user@example.com")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		writeError(w, r, http.StatusBadRequest,
			"missing email parameter",
			"send POST /auth/request with email=user@example.com")
		return
	}

	code := h.auth.RequestOTP(email)
	_ = code // In production, this would be emailed

	if wantsJSON(r) {
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"message":"OTP sent to %s","code":"%s","hint":"use this code with POST /auth/verify"}`,
				escapeJSON(email), code))
		return
	}
	writeText(w, http.StatusOK,
		fmt.Sprintf("OTP sent to %s | code: %s | hint: call POST /auth/verify with email and code to get a bearer token", email, code))
}

func (h *Handler) authVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed,
			"this endpoint requires POST",
			"use POST /auth/verify with email and code")
		return
	}

	email := r.FormValue("email")
	code := r.FormValue("code")
	if email == "" || code == "" {
		writeError(w, r, http.StatusBadRequest,
			"missing email or code parameter",
			"send POST /auth/verify with email=user@example.com and code=xxxxxx")
		return
	}

	token, err := h.auth.VerifyOTP(email, code)
	if err != nil {
		if ae, ok := err.(*authError); ok {
			writeError(w, r, http.StatusUnauthorized, ae.msg, ae.hint)
			return
		}
		writeError(w, r, http.StatusUnauthorized, "verification failed", "request a new OTP via POST /auth/request")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"token":"%s","hint":"use this token as Authorization: Bearer %s"}`,
				token, token))
		return
	}
	writeText(w, http.StatusOK,
		fmt.Sprintf("token: %s | hint: use this token as Authorization: Bearer %s", token, token))
}

func (h *Handler) eval(w http.ResponseWriter, r *http.Request) {
	expr := getExpr(r)
	if expr == "" {
		writeError(w, r, http.StatusBadRequest,
			"missing expression",
			"send the expression via query param ?expr=2+2 or POST body expr=2+2")
		return
	}

	// Check for degree mode override
	degree := h.config.Degree
	if d := r.URL.Query().Get("mode"); d == "degree" {
		degree = true
	} else if d == "radian" {
		degree = false
	}

	// Parse variables from query params (x=1, y=2, etc.)
	vars := parseVars(r)

	eval := model.NewEvaluator(degree)
	for k, v := range vars {
		eval.SetVar(k, v)
	}

	result, err := eval.Eval(expr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest,
			fmt.Sprintf("evaluation error: %s", err.Error()),
			"check the expression syntax. Call GET /help for supported operators and functions")
		return
	}

	formatted := model.FormatResult(result)

	if wantsJSON(r) {
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"expr":"%s","result":"%s","value":%s}`,
				escapeJSON(expr), formatted, formatted))
		return
	}
	writeText(w, http.StatusOK,
		fmt.Sprintf("expr=%s result=%s", expr, formatted))
}

func (h *Handler) evalBatch(w http.ResponseWriter, r *http.Request) {
	// Expressions can be passed as multiple expr= params or newline-separated in body
	var exprs []string

	if r.Method == http.MethodPost {
		r.ParseForm()
		exprs = r.Form["expr"]
		if len(exprs) == 0 {
			// Try body as newline-separated
			body := r.FormValue("expr")
			if body != "" {
				exprs = strings.Split(body, "\n")
			}
		}
	} else {
		q := r.URL.Query()
		exprs = q["expr"]
	}

	if len(exprs) == 0 {
		writeError(w, r, http.StatusBadRequest,
			"missing expressions",
			"send expressions via ?expr=2+2&expr=3*3 or POST body with multiple expr= fields")
		return
	}

	degree := h.config.Degree
	if d := r.URL.Query().Get("mode"); d == "degree" {
		degree = true
	} else if d == "radian" {
		degree = false
	}

	vars := parseVars(r)

	var textResults []string
	var jsonResults []string

	for _, expr := range exprs {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			continue
		}
		eval := model.NewEvaluator(degree)
		for k, v := range vars {
			eval.SetVar(k, v)
		}
		result, err := eval.Eval(expr)
		if err != nil {
			textResults = append(textResults, fmt.Sprintf("expr=%s error=%s", expr, err.Error()))
			jsonResults = append(jsonResults, fmt.Sprintf(`{"expr":"%s","error":"%s"}`,
				escapeJSON(expr), escapeJSON(err.Error())))
		} else {
			formatted := model.FormatResult(result)
			textResults = append(textResults, fmt.Sprintf("expr=%s result=%s", expr, formatted))
			jsonResults = append(jsonResults, fmt.Sprintf(`{"expr":"%s","result":"%s","value":%s}`,
				escapeJSON(expr), formatted, formatted))
		}
	}

	if wantsJSON(r) {
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"results":[%s]}`, strings.Join(jsonResults, ",")))
		return
	}
	writeText(w, http.StatusOK, strings.Join(textResults, "\n"))
}

func (h *Handler) functions(w http.ResponseWriter, r *http.Request) {
	fns := model.ListFunctions()

	if wantsJSON(r) {
		var parts []string
		for cat, list := range fns {
			var quoted []string
			for _, f := range list {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, f))
			}
			parts = append(parts, fmt.Sprintf(`"%s":[%s]`, cat, strings.Join(quoted, ",")))
		}
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"functions":{%s}}`, strings.Join(parts, ",")))
		return
	}

	var lines []string
	for cat, list := range fns {
		lines = append(lines, fmt.Sprintf("%s: %s", cat, strings.Join(list, ", ")))
	}
	writeText(w, http.StatusOK, strings.Join(lines, "\n"))
}

func (h *Handler) constants(w http.ResponseWriter, r *http.Request) {
	consts := model.ListConstants()

	if wantsJSON(r) {
		var parts []string
		for name, desc := range consts {
			parts = append(parts, fmt.Sprintf(`"%s":"%s"`, name, escapeJSON(desc)))
		}
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"constants":{%s}}`, strings.Join(parts, ",")))
		return
	}

	var lines []string
	for name, desc := range consts {
		lines = append(lines, fmt.Sprintf("%s = %s", name, desc))
	}
	writeText(w, http.StatusOK, strings.Join(lines, "\n"))
}

func (h *Handler) operators(w http.ResponseWriter, r *http.Request) {
	ops := model.ListOperators()

	if wantsJSON(r) {
		var parts []string
		for op, desc := range ops {
			parts = append(parts, fmt.Sprintf(`"%s":"%s"`, op, escapeJSON(desc)))
		}
		writeJSON(w, http.StatusOK,
			fmt.Sprintf(`{"operators":{%s}}`, strings.Join(parts, ",")))
		return
	}

	var lines []string
	for op, desc := range ops {
		lines = append(lines, fmt.Sprintf("%s — %s", op, desc))
	}
	writeText(w, http.StatusOK, strings.Join(lines, "\n"))
}

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	manual := `mathkit — Agentic-First Math Expression Evaluation Service

Evaluate mathematical expressions safely. Supports arithmetic, functions,
constants, factorial, power, modulo, variables, and batch evaluation.

== AUTH ==
1. POST /auth/request  email=user@example.com     → sends OTP code
2. POST /auth/verify    email=... code=...         → returns bearer token
3. Use: Authorization: Bearer <token>

== ENDPOINTS ==

POST/GET /eval?expr=<expression>
  Evaluate a single expression.
  Optional: &mode=degree (trig in degrees), &x=1&y=2 (variables)
  Example: /eval?expr=2+3*4           → result=14
  Example: /eval?expr=sqrt(16)+pi     → result=7.141592653589793
  Example: /eval?expr=sin(90)&mode=degree → result=1
  Example: /eval?expr=x*y&x=3&y=4     → result=12

POST/GET /eval/batch?expr=<e1>&expr=<e2>
  Evaluate multiple expressions. One result per line.
  Example: /eval/batch?expr=2+2&expr=3*3 → two results

GET /functions    — list all available functions by category
GET /constants    — list all available constants
GET /operators    — list all available operators
GET /health       — health check
GET /help         — this manual (also at /.well-known/agent.md)
POST /mcp         — MCP JSON-RPC 2.0 endpoint

== OPERATORS ==
+  -  *  /  %  ^ (power)  ** (power alias)  ! (factorial)  ()  ,

== FUNCTIONS ==
Arithmetic:  abs sign floor ceil round trunc min max avg sum pow gcd lcm clamp hypot
Logarithmic: sqrt exp ln log log10 log2 log_(base, x)
Trigonometric: sin cos tan asin acos atan atan2
Hyperbolic:  sinh cosh tanh asinh acosh atanh
Conversion:  deg(rad) rad(deg)

== CONSTANTS ==
pi = 3.14159...   e = 2.71828...   phi = 1.61803...   tau = 6.28318...

== RESPONSES ==
Plain text by default (one labeled line per result).
JSON on demand: Accept: application/json or ?format=json

== ERRORS ==
error: <message> | hint: <what to do next>
`
	writeText(w, http.StatusOK, manual)
}

// --- Helpers ---

func getExpr(r *http.Request) string {
	// Try query param first
	expr := r.URL.Query().Get("expr")
	if expr != "" {
		return expr
	}
	// Try POST body
	if r.Method == http.MethodPost {
		r.ParseForm()
		return r.FormValue("expr")
	}
	return ""
}

func parseVars(r *http.Request) map[string]float64 {
	vars := make(map[string]float64)
	// Known query params to skip
	skip := map[string]bool{
		"expr": true, "mode": true, "format": true,
	}
	for key, vals := range r.URL.Query() {
		if skip[key] {
			continue
		}
		if len(vals) > 0 {
			var f float64
			_, err := fmt.Sscanf(vals[0], "%g", &f)
			if err == nil {
				vars[key] = f
			}
		}
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		for key, vals := range r.Form {
			if skip[key] {
				continue
			}
			if len(vals) > 0 {
				var f float64
				_, err := fmt.Sscanf(vals[0], "%g", &f)
				if err == nil {
					vars[key] = f
				}
			}
		}
	}
	return vars
}
