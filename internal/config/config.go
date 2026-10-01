package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all service configuration.
type Config struct {
	Addr   string
	Secret string
	Degree bool
	NoAuth bool
}

// Default returns sensible zero-config defaults.
func Default() *Config {
	return &Config{
		Addr:   ":7100",
		Secret: "",
		Degree: false,
		NoAuth: false,
	}
}

// Load reads config from defaults < env < flags.
func Load() *Config {
	c := Default()

	// Env
	if v := os.Getenv("MATHKIT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("MATHKIT_SECRET"); v != "" {
		c.Secret = v
	}
	if v := os.Getenv("MATHKIT_DEGREE"); v == "1" || v == "true" {
		c.Degree = true
	}
	if v := os.Getenv("MATHKIT_NO_AUTH"); v == "1" || v == "true" {
		c.NoAuth = true
	}

	// Flags
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.StringVar(&c.Secret, "secret", c.Secret, "token signing secret (auto-generated if empty)")
	flag.BoolVar(&c.Degree, "degree", c.Degree, "use degrees for trig functions by default")
	flag.BoolVar(&c.NoAuth, "no-auth", c.NoAuth, "disable auth (for dev/testing)")
	flag.Parse()

	// Auto-generate secret if not provided
	if c.Secret == "" {
		c.Secret = randomSecret(32)
	}

	return c
}

func (c *Config) String() string {
	return fmt.Sprintf("addr=%s degree=%v no-auth=%v", c.Addr, c.Degree, c.NoAuth)
}

func randomSecret(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[fastRandN(uint64(len(chars)))]
	}
	return string(b)
}

var seed uint64 = 1

func fastRandN(n uint64) int {
	seed = seed*6364136223846793005 + 1442695040888963407
	return int(seed % n)
}
