# mathkit

Agentic-first math expression evaluation service. Evaluate arithmetic expressions with functions (sqrt, sin, cos, log, etc.), constants (pi, e), factorial, power, modulo, variables, and batch evaluation. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
# Build
make build

# Run (defaults to :7100)
./mathkit

# Evaluate an expression
curl "http://localhost:7100/eval?expr=2+3*4"
# expr=2+3*4 result=14

# With functions and constants
curl "http://localhost:7100/eval?expr=sqrt(16)+pi"
# expr=sqrt(16)+pi result=7.141592653589793

# Trig in degree mode
curl "http://localhost:7100/eval?expr=sin(90)&mode=degree"
# expr=sin(90)&mode=degree result=1

# With variables
curl "http://localhost:7100/eval?expr=x*y+1&x=3&y=4"
# expr=x*y+1 result=13

# Batch evaluation
curl "http://localhost:7100/eval/batch?expr=2+2&expr=3*3&expr=5!"
# expr=2+2 result=4
# expr=3*3 result=9
# expr=5! result=120

# JSON on demand
curl -H "Accept: application/json" "http://localhost:7100/eval?expr=2^10"
# {"expr":"2^10","result":"1024","value":1024}
```

## Auth

```bash
# 1. Request OTP
curl -X POST "http://localhost:7100/auth/request" -d "email=user@example.com"

# 2. Verify OTP → get bearer token
curl -X POST "http://localhost:7100/auth/verify" -d "email=user@example.com&code=xxxxxx"

# 3. Use token
curl -H "Authorization: Bearer <token>" "http://localhost:7100/eval?expr=pi*2"
```

For dev/testing, run with `--no-auth` to skip auth:
```bash
./mathkit -no-auth
```

## Operators

| Operator | Description |
|----------|-------------|
| `+` | Addition |
| `-` | Subtraction / unary negation |
| `*` | Multiplication |
| `/` | Division |
| `%` | Modulo |
| `^` or `**` | Power (right-associative) |
| `!` | Factorial (postfix, e.g. `5! = 120`) |
| `()` | Grouping / function call |
| `,` | Argument separator |

## Functions

| Category | Functions |
|----------|-----------|
| Arithmetic | abs, sign, floor, ceil, round, trunc, min, max, avg, sum, pow, gcd, lcm, clamp, hypot |
| Logarithmic | sqrt, exp, ln, log, log10, log2, log_(base, x) |
| Trigonometric | sin, cos, tan, asin, acos, atan, atan2 |
| Hyperbolic | sinh, cosh, tanh, asinh, acosh, atanh |
| Conversion | deg(radians→degrees), rad(degrees→radians) |

## Constants

| Constant | Value | Description |
|----------|-------|-------------|
| pi | 3.14159... | Ratio of circle circumference to diameter |
| e | 2.71828... | Euler's number |
| phi | 1.61803... | Golden ratio |
| tau | 6.28318... | 2*pi |

## Configuration

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `-addr` | `MATHKIT_ADDR` | `:7100` | Listen address |
| `-secret` | `MATHKIT_SECRET` | auto | Token signing secret |
| `-degree` | `MATHKIT_DEGREE` | `false` | Use degrees for trig by default |
| `-no-auth` | `MATHKIT_NO_AUTH` | `false` | Disable auth (dev/testing) |

## MCP

Model Context Protocol endpoint at `POST /mcp`:
- `initialize` — handshake
- `tools/list` — list available tools
- `tools/call` — call a tool (eval, eval_batch, list_functions, list_constants, list_operators)

## Build

```bash
make build    # CGO_ENABLED=0, single static binary
make test     # go test -race
make vet      # go vet
```

## License

MIT
