package service

import (
	"errors"
	"strings"
)

// MoneyScale is the fixed-point multiplier for monetary amounts.
// Values stored as int64 represent yuan * MoneyScale.
// e.g. 12.3456 yuan = 123456 (int64).
const MoneyScale int64 = 10000

// ParseMoneyX10000 parses a decimal string into a fixed-point int64.
// Input must be non-negative, at most 4 decimal places.
// Returns an error for empty, negative, non-numeric, or too-many-decimal inputs.
// Uses base-10 integer arithmetic only — no float32/float64/strconv.ParseFloat.
func ParseMoneyX10000(input string) (int64, error) {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return 0, errors.New("empty input")
	}

	// Check for negative sign
	if input[0] == '-' {
		return 0, errors.New("negative values are not allowed")
	}

	// Split on decimal point
	parts := strings.SplitN(input, ".", 2)

	// Parse integer part
	intPart := parts[0]
	var whole int64
	if len(intPart) == 0 {
		// Leading dot case: ".5" -> integer part is 0
		whole = 0
	} else {
		if !isAllDigits(intPart) {
			return 0, errors.New("non-numeric integer part")
		}
		for _, c := range intPart {
			whole = whole*10 + int64(c-'0')
		}
	}

	// Parse fractional part
	if len(parts) == 1 {
		// No decimal point: multiply by scale
		return whole * MoneyScale, nil
	}

	fracPart := parts[1]
	if len(fracPart) == 0 {
		return 0, errors.New("trailing decimal point without fractional digits")
	}
	if len(fracPart) > 4 {
		return 0, errors.New("more than 4 decimal places")
	}
	if !isAllDigits(fracPart) {
		return 0, errors.New("non-numeric fractional part")
	}

	// Pad fractional part to 4 digits
	for len(fracPart) < 4 {
		fracPart += "0"
	}

	var frac int64
	for _, c := range fracPart {
		frac = frac*10 + int64(c-'0')
	}

	return whole*MoneyScale + frac, nil
}

// FormatMoneyX10000 formats a fixed-point int64 back to a decimal string.
// Trailing zeros are stripped and the decimal point is omitted when fraction is zero.
func FormatMoneyX10000(value int64) string {
	if value == 0 {
		return "0"
	}

	whole := value / MoneyScale
	frac := value % MoneyScale
	if frac < 0 {
		frac = -frac
	}

	if frac == 0 {
		// Integer value
		// Use simple conversion without fmt.Sprintf to avoid float
		return int64ToStr(whole)
	}

	// Convert fractional part to string padded to 4 digits
	fracStr := ""
	tmp := frac
	for i := 0; i < 4; i++ {
		fracStr = string(rune('0')+rune(tmp%10)) + fracStr
		tmp /= 10
	}

	// Strip trailing zeros
	fracStr = strings.TrimRight(fracStr, "0")

	wholeStr := int64ToStr(whole)
	return wholeStr + "." + fracStr
}

// isAllDigits returns true if every character in s is an ASCII digit.
func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// int64ToStr converts an int64 to a decimal string without using fmt.Sprintf.
func int64ToStr(v int64) string {
	if v == 0 {
		return "0"
	}

	neg := v < 0
	if neg {
		v = -v
	}

	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}

	if neg {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
