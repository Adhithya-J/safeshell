package validator

import (
	"errors"
	"strings"
)

// Validate checks if a shell script is basically sound and safe.
func Validate(script string) error {
	if strings.TrimSpace(script) == "" {
		return errors.New("script is empty")
	}

	// Basic safety check: don't allow common escape attempts if they are obviously malicious
	// This is a very basic validator, most safety is handled by AI and the container itself.
	dangerous := []string{"rm -rf /", "chmod -R 777 /", ":(){ :|:& };:"}
	for _, d := range dangerous {
		if strings.Contains(script, d) {
			return errors.New("script contains dangerous command: " + d)
		}
	}

	return nil
}
