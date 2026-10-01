package model

import (
	"math"
	"testing"
)

func TestBasicArithmetic(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"2+3", 5},
		{"10-4", 6},
		{"3*4", 12},
		{"20/4", 5},
		{"10%3", 1},
		{"2+3*4", 14},
		{"(2+3)*4", 20},
		{"2^10", 1024},
		{"2**8", 256},
		{"-5+10", 5},
		{"-(3+2)", -5},
		{"+5", 5},
		{"10/3*3", 10},
		{"2^3^2", 512}, // right-associative: 2^(3^2) = 2^9 = 512
		{"(2^3)^2", 64},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"0!", 1},
		{"1!", 1},
		{"5!", 120},
		{"3!", 6},
		{"10!", 3628800},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestFunctions(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"sqrt(16)", 4},
		{"abs(-5)", 5},
		{"abs(5)", 5},
		{"floor(3.7)", 3},
		{"ceil(3.2)", 4},
		{"round(3.5)", 4},
		{"round(3.4)", 3},
		{"trunc(3.9)", 3},
		{"exp(0)", 1},
		{"ln(e)", 1},
		{"log(100)", 2},
		{"log10(1000)", 3},
		{"log2(8)", 3},
		{"min(3,1,2)", 1},
		{"max(3,1,2)", 3},
		{"avg(1,2,3)", 2},
		{"sum(1,2,3,4)", 10},
		{"gcd(12,8)", 4},
		{"lcm(4,6)", 12},
		{"pow(2,10)", 1024},
		{"hypot(3,4)", 5},
		{"clamp(5,1,10)", 5},
		{"clamp(15,1,10)", 10},
		{"clamp(-5,1,10)", 1},
		{"sign(5)", 1},
		{"sign(-5)", -1},
		{"sign(0)", 0},
		{"deg(pi)", 180},
		{"rad(180)", math.Pi},
		{"log_(2,8)", 3},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestTrigRadian(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"sin(0)", 0},
		{"cos(0)", 1},
		{"sin(pi/2)", 1},
		{"cos(pi)", -1},
		{"tan(0)", 0},
		{"asin(1)", math.Pi / 2},
		{"acos(0)", math.Pi / 2},
		{"atan(1)", math.Pi / 4},
	}
	for _, tt := range tests {
		e := NewEvaluator(false) // radian mode
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestTrigDegree(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"sin(0)", 0},
		{"sin(90)", 1},
		{"cos(0)", 1},
		{"cos(180)", -1},
		{"sin(30)", 0.5},
		{"cos(60)", 0.5},
		{"asin(1)", 90},
		{"acos(0)", 90},
		{"atan(1)", 45},
	}
	for _, tt := range tests {
		e := NewEvaluator(true) // degree mode
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"pi", math.Pi},
		{"e", math.E},
		{"phi", math.Phi},
		{"tau", 2 * math.Pi},
		{"pi*2", 2 * math.Pi},
		{"e*1", math.E},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestVariables(t *testing.T) {
	e := NewEvaluator(false)
	e.SetVar("x", 10)
	e.SetVar("y", 3)

	tests := []struct {
		expr string
		want float64
	}{
		{"x", 10},
		{"y", 3},
		{"x+y", 13},
		{"x-y", 7},
		{"x*y", 30},
		{"x/y", 10.0 / 3.0},
		{"x^2+y^2", 109},
		{"sqrt(x^2+y^2)", math.Sqrt(109)},
	}
	for _, tt := range tests {
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestHyperbolic(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"sinh(0)", 0},
		{"cosh(0)", 1},
		{"tanh(0)", 0},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestErrors(t *testing.T) {
	tests := []string{
		"",           // empty
		"2+",         // incomplete
		"2++3",       // double operator (actually valid: 2 + (+3) = 5, so this should NOT error)
		"sin()",      // empty args
		"unknown(5)", // unknown function
		"xyz",        // unknown identifier
		"1/0",        // division by zero
		"10%0",       // modulo by zero
		"(-1)^0.5",   // negative base with fractional exponent
		"0^-1",       // zero to negative power
		"5.5!",       // factorial of non-integer
		"(-1)!",      // factorial of negative
		"171!",       // factorial overflow
		"@",          // invalid character
	}
	for _, expr := range tests {
		e := NewEvaluator(false)
		_, err := e.Eval(expr)
		if err == nil && expr != "2++3" {
			t.Errorf("Eval(%q) expected error, got nil", expr)
		}
	}
}

func TestScientificNotation(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"1e3", 1000},
		{"1.5e2", 150},
		{"2E3", 2000},
		{"1.2e-1", 0.12},
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}

func TestFormatResult(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{5, "5"},
		{3.14, "3.14"},
		{100, "100"},
		{0.5, "0.5"},
		{math.Inf(1), "Infinity"},
		{math.Inf(-1), "-Infinity"},
		{math.NaN(), "NaN"},
	}
	for _, tt := range tests {
		got := FormatResult(tt.val)
		if got != tt.want {
			t.Errorf("FormatResult(%g) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"2+3*4-1", 13},
		{"(2+3)*(4-1)", 15},
		{"sqrt(16)+abs(-3)", 7},
		{"max(1,2,3)*min(4,5,6)", 12},
		{"sin(pi/2)*cos(0)", 1},
		{"2^3+sqrt(9)", 11},
		{"floor(3.7)+ceil(2.1)", 6},
		{"log(100)+ln(e)", 3},
		{"5!-3!", 114},
		{"(1+2)*(3+4)", 21},
		{"10/2+3*4", 17},
		{"-2^2", -4}, // unary minus has lower precedence than power: -(2^2) = -4
	}
	for _, tt := range tests {
		e := NewEvaluator(false)
		got, err := e.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %s", tt.expr, err)
			continue
		}
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.want)
		}
	}
}
