package validator_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCUIT(t *testing.T) {
	tests := []struct {
		name    string
		cuit    string
		wantErr bool
		errMsg  string
	}{
		{name: "valid CUIT prefix 20 with dashes", cuit: "20-12345678-6"},
		{name: "valid CUIT prefix 20 without dashes", cuit: "20123456786"},
		{name: "valid CUIT prefix 27", cuit: "27-10000000-3"},
		{name: "valid CUIT prefix 30 company", cuit: "30-71234567-1"},
		{name: "valid CUIT prefix 20 repeating digits", cuit: "20-33333333-4"},
		{name: "invalid check digit zero instead of six", cuit: "20-12345678-0", wantErr: true, errMsg: "invalid CUIT check digit"},
		{name: "invalid check digit off by one", cuit: "20-12345678-5", wantErr: true, errMsg: "invalid CUIT check digit"},
		{name: "invalid check digit nine instead of four", cuit: "20-33333333-9", wantErr: true, errMsg: "invalid CUIT check digit"},
		{name: "invalid prefix 21", cuit: "21-12345678-6", wantErr: true, errMsg: "invalid CUIT prefix: 21"},
		{name: "invalid prefix 10", cuit: "10-12345678-6", wantErr: true, errMsg: "invalid CUIT prefix: 10"},
		{name: "invalid prefix 50", cuit: "50-12345678-6", wantErr: true, errMsg: "invalid CUIT prefix: 50"},
		{name: "invalid prefix 99", cuit: "99-12345678-6", wantErr: true, errMsg: "invalid CUIT prefix: 99"},
		{name: "too short 10 digits", cuit: "20-1234567-6", wantErr: true, errMsg: "CUIT must have 11 digits"},
		{name: "too long 12 digits", cuit: "20-123456789-6", wantErr: true, errMsg: "CUIT must have 11 digits"},
		{name: "empty string", cuit: "", wantErr: true, errMsg: "CUIT must have 11 digits"},
		{name: "single digit", cuit: "2", wantErr: true, errMsg: "CUIT must have 11 digits"},
		{name: "contains letters", cuit: "20-1234567A-6", wantErr: true, errMsg: "CUIT must contain only digits"},
		{name: "contains spaces after removing dashes", cuit: "20 12345678 6", wantErr: true, errMsg: "CUIT must have 11 digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCUIT(tt.cuit)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCUIT_AllValidPrefixes(t *testing.T) {
	validPrefixCUITs := []string{
		"20-00000000-1",
		"23-00000000-0",
		"24-00000000-7",
		"27-00000000-6",
		"30-00000000-7",
		"33-00000000-6",
		"34-00000000-2",
	}

	for _, cuit := range validPrefixCUITs {
		t.Run("prefix_"+cuit[:2], func(t *testing.T) {
			err := validator.ValidateCUIT(cuit)
			assert.NoError(t, err, "CUIT %s should be valid", cuit)
		})
	}
}

func TestValidateDNI(t *testing.T) {
	tests := []struct {
		name    string
		dni     string
		wantErr bool
		errMsg  string
	}{
		{name: "valid 8 digit DNI", dni: "12345678"},
		{name: "valid 7 digit DNI", dni: "1234567"},
		{name: "too short 6 digits", dni: "123456", wantErr: true, errMsg: "DNI must have 7 or 8 digits"},
		{name: "too long 9 digits", dni: "123456789", wantErr: true, errMsg: "DNI must have 7 or 8 digits"},
		{name: "empty string", dni: "", wantErr: true, errMsg: "DNI must have 7 or 8 digits"},
		{name: "contains letters", dni: "1234567A", wantErr: true, errMsg: "DNI must contain only digits"},
		{name: "contains special chars", dni: "12345-78", wantErr: true, errMsg: "DNI must contain only digits"},
		{name: "all zeros 7 digits is valid", dni: "0000000"},
		{name: "all nines 8 digits is valid", dni: "99999999"},
		{name: "whitespace trimmed to valid 8 digits", dni: " 12345678 "},
		{name: "only whitespace", dni: "   ", wantErr: true, errMsg: "DNI must have 7 or 8 digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDNI(tt.dni)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatCUIT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "format raw 11 digits", input: "20123456786", expected: "20-12345678-6"},
		{name: "already formatted passes through", input: "20-12345678-6", expected: "20-12345678-6"},
		{name: "too short returns unchanged", input: "2012345", expected: "2012345"},
		{name: "too long returns unchanged", input: "201234567890", expected: "201234567890"},
		{name: "empty string returns unchanged", input: "", expected: ""},
		{name: "format CUIT prefix 27", input: "27100000003", expected: "27-10000000-3"},
		{name: "format CUIT prefix 30", input: "30712345671", expected: "30-71234567-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.FormatCUIT(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
