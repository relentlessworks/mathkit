package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// DegreeMode controls whether trig functions use degrees (true) or radians (false).
type DegreeMode bool

// Evaluator is a safe recursive-descent math expression evaluator.
// It supports arithmetic, functions, constants, variables, factorial, and power.
type Evaluator struct {
	vars map[string]float64
	mode DegreeMode
}

// NewEvaluator creates an evaluator with the given degree mode.
func NewEvaluator(degree bool) *Evaluator {
	return &Evaluator{
		vars: make(map[string]float64),
		mode: DegreeMode(degree),
	}
}

// SetVar sets a variable for use in expressions.
func (e *Evaluator) SetVar(name string, val float64) {
	e.vars[strings.ToLower(name)] = val
}

// Eval parses and evaluates the given expression string.
func (e *Evaluator) Eval(expr string) (float64, error) {
	p := newParser(expr, e)
	return p.parseExpr()
}

// --- Tokenizer ---

type tokenType int

const (
	tokNumber tokenType = iota
	tokIdent
	tokPlus
	tokMinus
	tokMul
	tokDiv
	tokMod
	tokPow
	tokFact
	tokLParen
	tokRParen
	tokComma
	tokEOF
	tokError
)

type token struct {
	typ  tokenType
	val  string
	num  float64
	pos  int
}

type tokenizer struct {
	src  string
	pos  int
}

func newTokenizer(src string) *tokenizer {
	return &tokenizer{src: src}
}

func (t *tokenizer) next() (token, error) {
	// Skip whitespace
	for t.pos < len(t.src) && unicode.IsSpace(rune(t.src[t.pos])) {
		t.pos++
	}
	if t.pos >= len(t.src) {
		return token{typ: tokEOF, pos: t.pos}, nil
	}

	start := t.pos
	ch := t.src[t.pos]

	switch ch {
	case '+':
		t.pos++
		return token{typ: tokPlus, val: "+", pos: start}, nil
	case '-':
		t.pos++
		return token{typ: tokMinus, val: "-", pos: start}, nil
	case '*':
		t.pos++
		// Check for ** (power)
		if t.pos < len(t.src) && t.src[t.pos] == '*' {
			t.pos++
			return token{typ: tokPow, val: "**", pos: start}, nil
		}
		return token{typ: tokMul, val: "*", pos: start}, nil
	case '/':
		t.pos++
		return token{typ: tokDiv, val: "/", pos: start}, nil
	case '%':
		t.pos++
		return token{typ: tokMod, val: "%", pos: start}, nil
	case '^':
		t.pos++
		return token{typ: tokPow, val: "^", pos: start}, nil
	case '!':
		t.pos++
		return token{typ: tokFact, val: "!", pos: start}, nil
	case '(':
		t.pos++
		return token{typ: tokLParen, val: "(", pos: start}, nil
	case ')':
		t.pos++
		return token{typ: tokRParen, val: ")", pos: start}, nil
	case ',':
		t.pos++
		return token{typ: tokComma, val: ",", pos: start}, nil
	}

	// Number
	if ch == '.' || (ch >= '0' && ch <= '9') {
		return t.readNumber(start)
	}

	// Identifier (function or variable or constant)
	if isIdentStart(ch) {
		return t.readIdent(start)
	}

	return token{typ: tokError, val: string(ch), pos: start}, fmt.Errorf("unexpected character '%c' at position %d", ch, start)
}

func (t *tokenizer) readNumber(start int) (token, error) {
	for t.pos < len(t.src) {
		ch := t.src[t.pos]
		if ch >= '0' && ch <= '9' || ch == '.' {
			t.pos++
		} else if ch == 'e' || ch == 'E' {
			// Scientific notation
			t.pos++
			if t.pos < len(t.src) && (t.src[t.pos] == '+' || t.src[t.pos] == '-') {
				t.pos++
			}
		} else {
			break
		}
	}
	numStr := t.src[start:t.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number '%s' at position %d", numStr, start)
	}
	return token{typ: tokNumber, val: numStr, num: num, pos: start}, nil
}

func (t *tokenizer) readIdent(start int) (token, error) {
	for t.pos < len(t.src) && isIdentPart(t.src[t.pos]) {
		t.pos++
	}
	name := t.src[start:t.pos]
	return token{typ: tokIdent, val: name, pos: start}, nil
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// --- Parser (recursive descent) ---

type parser struct {
	tok  *tokenizer
	cur  token
	peek token
	eval *Evaluator
}

func newParser(expr string, eval *Evaluator) *parser {
	t := newTokenizer(expr)
	p := &parser{tok: t, eval: eval}
	// Prime the pump
	p.cur, _ = t.next()
	p.peek, _ = t.next()
	return p
}

func (p *parser) advance() {
	p.cur = p.peek
	p.peek, _ = p.tok.next()
}

// Grammar:
//   expr    := term (('+' | '-') term)*
//   term    := factor (('*' | '/' | '%') factor)*
//   factor  := unary
//   unary   := ('+' | '-') unary | power   // unary minus has lower precedence than power
//   power   := postfix ('^' unary)?        // right-associative, allows unary in exponent
//   postfix := primary ('!')*
//   primary := number | ident | ident '(' args ')' | '(' expr ')'

func (p *parser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for p.cur.typ == tokPlus || p.cur.typ == tokMinus {
		op := p.cur
		p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op.typ == tokPlus {
			left = left + right
		} else {
			left = left - right
		}
	}
	return left, nil
}

func (p *parser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for p.cur.typ == tokMul || p.cur.typ == tokDiv || p.cur.typ == tokMod {
		op := p.cur
		p.advance()
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		switch op.typ {
		case tokMul:
			left = left * right
		case tokDiv:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left = left / right
		case tokMod:
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		}
	}
	return left, nil
}

func (p *parser) parseFactor() (float64, error) {
	return p.parseUnary()
}

func (p *parser) parseUnary() (float64, error) {
	if p.cur.typ == tokPlus {
		p.advance()
		return p.parseUnary()
	}
	if p.cur.typ == tokMinus {
		p.advance()
		val, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return -val, nil
	}
	return p.parsePower()
}

func (p *parser) parsePower() (float64, error) {
	base, err := p.parsePostfix()
	if err != nil {
		return 0, err
	}
	if p.cur.typ == tokPow {
		p.advance()
		exp, err := p.parseUnary() // allows 2^-2, right-associative via unary→power
		if err != nil {
			return 0, err
		}
		// Check for negative base with fractional exponent
		if base < 0 && exp != math.Trunc(exp) {
			return 0, fmt.Errorf("negative base with fractional exponent")
		}
		if base == 0 && exp < 0 {
			return 0, fmt.Errorf("zero raised to negative power")
		}
		return math.Pow(base, exp), nil
	}
	return base, nil
}

func (p *parser) parsePostfix() (float64, error) {
	val, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	// Handle postfix factorial
	for p.cur.typ == tokFact {
		p.advance()
		if val < 0 || val != math.Trunc(val) {
			return 0, fmt.Errorf("factorial requires non-negative integer, got %g", val)
		}
		if val > 170 {
			return 0, fmt.Errorf("factorial overflow: %g! is too large", val)
		}
		val = factorial(int(val))
	}
	return val, nil
}

func (p *parser) parsePrimary() (float64, error) {
	switch p.cur.typ {
	case tokNumber:
		val := p.cur.num
		p.advance()
		return val, nil

	case tokLParen:
		p.advance()
		val, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.cur.typ != tokRParen {
			return 0, fmt.Errorf("expected ')' at position %d", p.cur.pos)
		}
		p.advance()
		return val, nil

	case tokIdent:
		name := strings.ToLower(p.cur.val)
		p.advance()

		// Function call?
		if p.cur.typ == tokLParen {
			p.advance()
			args, err := p.parseArgs()
			if err != nil {
				return 0, err
			}
			if p.cur.typ != tokRParen {
				return 0, fmt.Errorf("expected ')' after function arguments at position %d", p.cur.pos)
			}
			p.advance()
			return p.callFunction(name, args)
		}

		// Constant or variable
		return p.lookupIdent(name)

	case tokError:
		return 0, fmt.Errorf("unexpected character '%s' at position %d", p.cur.val, p.cur.pos)
	}

	return 0, fmt.Errorf("unexpected token '%s' at position %d", p.cur.val, p.cur.pos)
}

func (p *parser) parseArgs() ([]float64, error) {
	var args []float64

	// Empty args
	if p.cur.typ == tokRParen {
		return args, nil
	}

	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	args = append(args, val)

	for p.cur.typ == tokComma {
		p.advance()
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	return args, nil
}

func (p *parser) callFunction(name string, args []float64) (float64, error) {
	// Single-argument functions
	fns1 := map[string]func(float64) float64{
		"sqrt":   math.Sqrt,
		"abs":    math.Abs,
		"floor":  math.Floor,
		"ceil":   math.Ceil,
		"round":  math.Round,
		"trunc":  math.Trunc,
		"ln":     math.Log,
		"log":    math.Log10,
		"log10":  math.Log10,
		"log2":   math.Log2,
		"exp":    math.Exp,
		"sign":   nil, // handled separately
		"deg":    radToDeg,
		"rad":    degToRad,
		"sinh":   math.Sinh,
		"cosh":   math.Cosh,
		"tanh":   math.Tanh,
		"asinh":  math.Asinh,
		"acosh":  math.Acosh,
		"atanh":  math.Atanh,
	}

	// Trig functions (respect degree mode)
	trigFns := map[string]func(float64) float64{
		"sin":   math.Sin,
		"cos":   math.Cos,
		"tan":   math.Tan,
		"asin":  math.Asin,
		"acos":  math.Acos,
		"atan":  math.Atan,
	}

	// Multi-arg functions
	switch name {
	case "min":
		if len(args) < 1 {
			return 0, fmt.Errorf("min() requires at least 1 argument")
		}
		result := args[0]
		for _, a := range args[1:] {
			if a < result {
				result = a
			}
		}
		return result, nil

	case "max":
		if len(args) < 1 {
			return 0, fmt.Errorf("max() requires at least 1 argument")
		}
		result := args[0]
		for _, a := range args[1:] {
			if a > result {
				result = a
			}
		}
		return result, nil

	case "avg", "mean":
		if len(args) < 1 {
			return 0, fmt.Errorf("avg() requires at least 1 argument")
		}
		sum := 0.0
		for _, a := range args {
			sum += a
		}
		return sum / float64(len(args)), nil

	case "sum":
		if len(args) < 1 {
			return 0, fmt.Errorf("sum() requires at least 1 argument")
		}
		sum := 0.0
		for _, a := range args {
			sum += a
		}
		return sum, nil

	case "gcd":
		if len(args) < 2 {
			return 0, fmt.Errorf("gcd() requires at least 2 arguments")
		}
		result := args[0]
		for _, a := range args[1:] {
			result = gcd(result, a)
		}
		return result, nil

	case "lcm":
		if len(args) < 2 {
			return 0, fmt.Errorf("lcm() requires at least 2 arguments")
		}
		result := args[0]
		for _, a := range args[1:] {
			result = lcm(result, a)
		}
		return result, nil

	case "pow":
		if len(args) != 2 {
			return 0, fmt.Errorf("pow() requires exactly 2 arguments, got %d", len(args))
		}
		if args[0] < 0 && args[1] != math.Trunc(args[1]) {
			return 0, fmt.Errorf("pow() with negative base requires integer exponent")
		}
		return math.Pow(args[0], args[1]), nil

	case "atan2":
		if len(args) != 2 {
			return 0, fmt.Errorf("atan2() requires exactly 2 arguments, got %d", len(args))
		}
		result := math.Atan2(args[0], args[1])
		if p.eval.mode {
			result = radToDeg(result)
		}
		return result, nil

	case "log_":
		// log_(base, x) = log(x) / log(base)
		if len(args) != 2 {
			return 0, fmt.Errorf("log_() requires exactly 2 arguments, got %d", len(args))
		}
		if args[0] <= 0 || args[0] == 1 {
			return 0, fmt.Errorf("log_() base must be positive and not 1")
		}
		if args[1] <= 0 {
			return 0, fmt.Errorf("log_() argument must be positive")
		}
		return math.Log(args[1]) / math.Log(args[0]), nil

	case "clamp":
		// clamp(value, min, max)
		if len(args) != 3 {
			return 0, fmt.Errorf("clamp() requires exactly 3 arguments, got %d", len(args))
		}
		return math.Max(args[1], math.Min(args[0], args[2])), nil

	case "hypot":
		// hypot(x, y) = sqrt(x^2 + y^2)
		if len(args) != 2 {
			return 0, fmt.Errorf("hypot() requires exactly 2 arguments, got %d", len(args))
		}
		return math.Hypot(args[0], args[1]), nil
	}

	// Single-arg functions
	if fn, ok := fns1[name]; ok {
		if len(args) != 1 {
			return 0, fmt.Errorf("%s() requires exactly 1 argument, got %d", name, len(args))
		}
		if name == "sign" {
			// sign(x) returns -1, 0, or 1
			if args[0] > 0 {
				return 1, nil
			} else if args[0] < 0 {
				return -1, nil
			}
			return 0, nil
		}
		return fn(args[0]), nil
	}

	// Trig functions with degree mode
	if fn, ok := trigFns[name]; ok {
		if len(args) != 1 {
			return 0, fmt.Errorf("%s() requires exactly 1 argument, got %d", name, len(args))
		}
		arg := args[0]
		// Inverse trig: result is in radians, convert to degrees if mode is degree
		if strings.HasPrefix(name, "a") {
			result := fn(arg)
			if p.eval.mode {
				result = radToDeg(result)
			}
			return result, nil
		}
		// Forward trig: if degree mode, convert input from degrees to radians
		if p.eval.mode {
			arg = degToRad(arg)
		}
		return fn(arg), nil
	}

	return 0, fmt.Errorf("unknown function '%s'", name)
}

func (p *parser) lookupIdent(name string) (float64, error) {
	// Constants
	constants := map[string]float64{
		"pi":    math.Pi,
		"e":     math.E,
		"phi":   math.Phi,
		"tau":   2 * math.Pi,
		"inf":   math.Inf(1),
		"nan":   math.NaN(),
		"true":  1,
		"false": 0,
	}

	if val, ok := constants[name]; ok {
		return val, nil
	}

	// User variables
	if val, ok := p.eval.vars[name]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("unknown identifier '%s'", name)
}

// --- Helper functions ---

func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
	}
	return result
}

func gcd(a, b float64) float64 {
	ai, bi := int64(math.Abs(a)), int64(math.Abs(b))
	for bi != 0 {
		ai, bi = bi, ai%bi
	}
	if ai == 0 {
		return 0
	}
	return float64(ai)
}

func lcm(a, b float64) float64 {
	if a == 0 || b == 0 {
		return 0
	}
	g := gcd(a, b)
	if g == 0 {
		return 0
	}
	return math.Abs(a*b) / g
}

func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

// FormatResult formats a float64 result for display.
// Removes trailing zeros and unnecessary decimal points.
func FormatResult(val float64) string {
	if math.IsInf(val, 1) {
		return "Infinity"
	}
	if math.IsInf(val, -1) {
		return "-Infinity"
	}
	if math.IsNaN(val) {
		return "NaN"
	}
	// Use %g for compact representation, but with enough precision
	s := strconv.FormatFloat(val, 'g', -1, 64)
	// If the result is an integer, show without decimal
	if val == math.Trunc(val) && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
		return strconv.FormatFloat(val, 'f', 0, 64)
	}
	return s
}

// ListFunctions returns all available function names grouped by category.
func ListFunctions() map[string][]string {
	return map[string][]string{
		"arithmetic": {"abs", "sign", "floor", "ceil", "round", "trunc", "min", "max", "avg", "sum", "pow", "mod", "gcd", "lcm", "clamp", "hypot"},
		"logarithmic": {"sqrt", "exp", "ln", "log", "log10", "log2", "log_"},
		"trigonometric": {"sin", "cos", "tan", "asin", "acos", "atan", "atan2"},
		"hyperbolic": {"sinh", "cosh", "tanh", "asinh", "acosh", "atanh"},
		"conversion": {"deg", "rad"},
	}
}

// ListConstants returns all available constants.
func ListConstants() map[string]string {
	return map[string]string{
		"pi":  "3.141592653589793 (ratio of circle circumference to diameter)",
		"e":   "2.718281828459045 (Euler's number, base of natural logarithm)",
		"phi": "1.618033988749895 (golden ratio)",
		"tau": "6.283185307179586 (2*pi, full circle in radians)",
		"inf": "positive infinity",
		"nan": "not a number",
	}
}

// ListOperators returns all available operators.
func ListOperators() map[string]string {
	return map[string]string{
		"+":  "addition",
		"-":  "subtraction (binary or unary negation)",
		"*":  "multiplication",
		"/":  "division",
		"%":  "modulo",
		"^":  "power (right-associative)",
		"**": "power (alias for ^)",
		"!":  "factorial (postfix, e.g. 5! = 120)",
		"()": "grouping / function call",
		",":  "argument separator",
	}
}
