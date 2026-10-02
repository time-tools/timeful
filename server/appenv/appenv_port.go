package appenv

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolvePort stays handwritten beside the generated appenv.go because GALA has
// no two-value return syntax.
func ResolvePort(env Environment, override string) (string, error) {
	value := strings.TrimSpace(override)
	if value == "" {
		return Port(env), nil
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("APP_PORT must be an integer between 1 and 65535, got %q", override)
	}

	return value, nil
}
