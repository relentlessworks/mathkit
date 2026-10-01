package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/mathkit/internal/model"
)

// MCPRequest is a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse is a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError is a JSON-RPC 2.0 error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed,
			"this endpoint requires POST",
			"send a JSON-RPC 2.0 request body")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, http.StatusBadRequest,
			"failed to read request body",
			"ensure the request body is valid")
		return
	}
	defer r.Body.Close()

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, r, http.StatusBadRequest,
			"invalid JSON-RPC request",
			"send a valid JSON-RPC 2.0 request with jsonrpc, method, and params fields")
		return
	}

	var resp MCPResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "mathkit",
				"version": "0.1.0",
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": mcpTools(),
		}

	case "tools/call":
		result, err := h.handleMCPToolCall(req.Params)
		if err != nil {
			resp.Error = &MCPError{Code: -32603, Message: err.Error()}
		} else {
			resp.Result = map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": result},
				},
			}
		}

	case "ping":
		resp.Result = map[string]interface{}{}

	default:
		resp.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("unknown method: %s", req.Method),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func mcpTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "eval",
			"description": "Evaluate a math expression and return the result.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expr": map[string]interface{}{
						"type":        "string",
						"description": "Math expression to evaluate (e.g. 2+3*4, sqrt(16), sin(pi/2))",
					},
					"mode": map[string]interface{}{
						"type":        "string",
						"description": "Angle mode: 'degree' or 'radian' (default: radian)",
						"enum":        []string{"degree", "radian"},
					},
				},
				"required": []string{"expr"},
			},
		},
		{
			"name":        "eval_batch",
			"description": "Evaluate multiple math expressions at once.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"exprs": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Array of math expressions to evaluate",
					},
					"mode": map[string]interface{}{
						"type":        "string",
						"description": "Angle mode: 'degree' or 'radian'",
						"enum":        []string{"degree", "radian"},
					},
				},
				"required": []string{"exprs"},
			},
		},
		{
			"name":        "list_functions",
			"description": "List all available math functions by category.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "list_constants",
			"description": "List all available math constants.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "list_operators",
			"description": "List all available operators.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (h *Handler) handleMCPToolCall(params json.RawMessage) (string, error) {
	var p struct {
		Name      string   `json:"name"`
		Arguments struct {
			Expr  string   `json:"expr"`
			Exprs []string `json:"exprs"`
			Mode  string   `json:"mode"`
		} `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("invalid arguments: %s", err)
	}

	switch p.Name {
	case "eval":
		if p.Arguments.Expr == "" {
			return "", fmt.Errorf("missing 'expr' argument")
		}
		degree := h.config.Degree
		if p.Arguments.Mode == "degree" {
			degree = true
		}
		eval := model.NewEvaluator(degree)
		result, err := eval.Eval(p.Arguments.Expr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s = %s", p.Arguments.Expr, model.FormatResult(result)), nil

	case "eval_batch":
		if len(p.Arguments.Exprs) == 0 {
			return "", fmt.Errorf("missing 'exprs' argument")
		}
		degree := h.config.Degree
		if p.Arguments.Mode == "degree" {
			degree = true
		}
		var lines []string
		for _, expr := range p.Arguments.Exprs {
			expr = strings.TrimSpace(expr)
			if expr == "" {
				continue
			}
			eval := model.NewEvaluator(degree)
			result, err := eval.Eval(expr)
			if err != nil {
				lines = append(lines, fmt.Sprintf("%s = error: %s", expr, err.Error()))
			} else {
				lines = append(lines, fmt.Sprintf("%s = %s", expr, model.FormatResult(result)))
			}
		}
		return strings.Join(lines, "\n"), nil

	case "list_functions":
		fns := model.ListFunctions()
		var lines []string
		for cat, list := range fns {
			lines = append(lines, fmt.Sprintf("%s: %s", cat, strings.Join(list, ", ")))
		}
		return strings.Join(lines, "\n"), nil

	case "list_constants":
		consts := model.ListConstants()
		var lines []string
		for name, desc := range consts {
			lines = append(lines, fmt.Sprintf("%s = %s", name, desc))
		}
		return strings.Join(lines, "\n"), nil

	case "list_operators":
		ops := model.ListOperators()
		var lines []string
		for op, desc := range ops {
			lines = append(lines, fmt.Sprintf("%s — %s", op, desc))
		}
		return strings.Join(lines, "\n"), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", p.Name)
	}
}
