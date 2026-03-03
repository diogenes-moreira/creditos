package validator

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateCUIT validates an Argentine CUIT/CUIL number.
// Format: XX-XXXXXXXX-X (11 digits total).
func ValidateCUIT(cuit string) error {
	cleaned := strings.ReplaceAll(cuit, "-", "")
	if len(cleaned) != 11 {
		return fmt.Errorf("CUIT must have 11 digits, got %d", len(cleaned))
	}

	for _, c := range cleaned {
		if c < '0' || c > '9' {
			return fmt.Errorf("CUIT must contain only digits")
		}
	}

	prefix := cleaned[:2]
	validPrefixes := map[string]bool{
		"20": true, "23": true, "24": true, "27": true,
		"30": true, "33": true, "34": true,
	}
	if !validPrefixes[prefix] {
		return fmt.Errorf("invalid CUIT prefix: %s", prefix)
	}

	weights := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i, w := range weights {
		d, _ := strconv.Atoi(string(cleaned[i]))
		sum += d * w
	}

	remainder := sum % 11
	var expectedDigit int
	switch remainder {
	case 0:
		expectedDigit = 0
	case 1:
		expectedDigit = 9
	default:
		expectedDigit = 11 - remainder
	}

	actualDigit, _ := strconv.Atoi(string(cleaned[10]))
	if actualDigit != expectedDigit {
		return fmt.Errorf("invalid CUIT check digit: expected %d, got %d", expectedDigit, actualDigit)
	}

	return nil
}

// ValidateDNI validates an Argentine DNI number (7 or 8 digits).
func ValidateDNI(dni string) error {
	cleaned := strings.TrimSpace(dni)
	if len(cleaned) < 7 || len(cleaned) > 8 {
		return fmt.Errorf("DNI must have 7 or 8 digits, got %d", len(cleaned))
	}
	for _, c := range cleaned {
		if c < '0' || c > '9' {
			return fmt.Errorf("DNI must contain only digits")
		}
	}
	return nil
}

// FormatCUIT formats a CUIT as XX-XXXXXXXX-X.
func FormatCUIT(cuit string) string {
	cleaned := strings.ReplaceAll(cuit, "-", "")
	if len(cleaned) != 11 {
		return cuit
	}
	return cleaned[:2] + "-" + cleaned[2:10] + "-" + cleaned[10:]
}
