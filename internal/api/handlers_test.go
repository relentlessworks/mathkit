package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestHandler() *Handler {
	auth := NewAuth("test-secret")
	return NewHandler(auth, false, true) // no-auth for testing
}

func TestEvalGET(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr="+url.QueryEscape("2+3*4"), nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=14") {
		t.Errorf("body = %q, want result=14", body)
	}
}

func TestEvalWithFunction(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr=sqrt(16)", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=4") {
		t.Errorf("body = %q, want result=4", body)
	}
}

func TestEvalJSON(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr=2^10", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"result":"1024"`) {
		t.Errorf("body = %q, want result 1024", body)
	}
	if !strings.Contains(body, `"value":1024`) {
		t.Errorf("body = %q, want value 1024", body)
	}
}

func TestEvalDegreeMode(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr=sin(90)&mode=degree", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=1") {
		t.Errorf("body = %q, want result=1", body)
	}
}

func TestEvalWithVariables(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr=x*y&x=3&y=4", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=12") {
		t.Errorf("body = %q, want result=12", body)
	}
}

func TestEvalPOST(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	form := url.Values{}
	form.Set("expr", "5!")
	req := httptest.NewRequest("POST", "/eval", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=120") {
		t.Errorf("body = %q, want result=120", body)
	}
}

func TestEvalBatch(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval/batch?expr="+url.QueryEscape("2+2")+"&expr="+url.QueryEscape("3*3")+"&expr="+url.QueryEscape("5!"), nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "result=4") {
		t.Errorf("body missing result=4: %q", body)
	}
	if !strings.Contains(body, "result=9") {
		t.Errorf("body missing result=9: %q", body)
	}
	if !strings.Contains(body, "result=120") {
		t.Errorf("body missing result=120: %q", body)
	}
}

func TestEvalError(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval?expr=1/0", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	body := w.Body.String()
	if !strings.Contains(body, "error:") {
		t.Errorf("body should contain error: %q", body)
	}
	if !strings.Contains(body, "hint:") {
		t.Errorf("body should contain hint: %q", body)
	}
}

func TestEvalMissingExpr(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/eval", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestFunctions(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/functions", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "sqrt") {
		t.Errorf("body should contain sqrt: %q", body)
	}
	if !strings.Contains(body, "sin") {
		t.Errorf("body should contain sin: %q", body)
	}
}

func TestConstants(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/constants", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "pi") {
		t.Errorf("body should contain pi: %q", body)
	}
}

func TestOperators(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/operators", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "factorial") {
		t.Errorf("body should contain factorial: %q", body)
	}
}

func TestHelp(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/help", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "mathkit") {
		t.Errorf("body should contain mathkit: %q", body)
	}
	if !strings.Contains(body, "/eval") {
		t.Errorf("body should contain /eval: %q", body)
	}
}

func TestHealth(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if body != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestRoot(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNotFound(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAuthFlow(t *testing.T) {
	auth := NewAuth("test-secret")
	h := NewHandler(auth, false, false)
	mux := h.Routes()

	// Request OTP
	form := url.Values{}
	form.Set("email", "test@example.com")
	req := httptest.NewRequest("POST", "/auth/request", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("auth request status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Extract code from response
	codeIdx := strings.Index(body, "code: ")
	if codeIdx == -1 {
		t.Fatalf("could not find code in response: %q", body)
	}
	code := strings.TrimSpace(body[codeIdx+6:])
	codeEnd := strings.Index(code, " ")
	if codeEnd > 0 {
		code = code[:codeEnd]
	}

	// Verify OTP
	form2 := url.Values{}
	form2.Set("email", "test@example.com")
	form2.Set("code", code)
	req2 := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("auth verify status = %d, want %d", w2.Code, http.StatusOK)
	}
	body2 := w2.Body.String()
	tokenIdx := strings.Index(body2, "token: ")
	if tokenIdx == -1 {
		t.Fatalf("could not find token in response: %q", body2)
	}
	token := strings.TrimSpace(body2[tokenIdx+7:])
	tokenEnd := strings.Index(token, " ")
	if tokenEnd > 0 {
		token = token[:tokenEnd]
	}
	if token == "" {
		t.Fatal("token is empty")
	}
}

func TestMCPInitialize(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, "mathkit") {
		t.Errorf("response should contain mathkit: %q", resp)
	}
}

func TestMCPToolsList(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, "eval") {
		t.Errorf("response should contain eval tool: %q", resp)
	}
}

func TestMCPToolsCallEval(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"eval","arguments":{"expr":"2+3*4"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, "14") {
		t.Errorf("response should contain 14: %q", resp)
	}
}

func TestMCPToolsCallListFunctions(t *testing.T) {
	h := newTestHandler()
	mux := h.Routes()

	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_functions","arguments":{}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, "sqrt") {
		t.Errorf("response should contain sqrt: %q", resp)
	}
}
