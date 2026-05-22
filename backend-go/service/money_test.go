package service

import (
	"testing"
)

func TestParseMoneyX10000_Zero(t *testing.T) {
	v, err := ParseMoneyX10000("0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Fatalf("expected 0, got %d", v)
	}
}

func TestParseMoneyX10000_FourDecimals(t *testing.T) {
	v, err := ParseMoneyX10000("12.3456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 123456 {
		t.Fatalf("expected 123456, got %d", v)
	}
}

func TestParseMoneyX10000_Integer(t *testing.T) {
	v, err := ParseMoneyX10000("5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 50000 {
		t.Fatalf("expected 50000, got %d", v)
	}
}

func TestParseMoneyX10000_TooManyDecimals(t *testing.T) {
	_, err := ParseMoneyX10000("12.34567")
	if err == nil {
		t.Fatal("expected error for >4 decimals, got nil")
	}
}

func TestParseMoneyX10000_Empty(t *testing.T) {
	_, err := ParseMoneyX10000("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestParseMoneyX10000_Negative(t *testing.T) {
	_, err := ParseMoneyX10000("-0.0001")
	if err == nil {
		t.Fatal("expected error for negative, got nil")
	}
}

func TestParseMoneyX10000_NonNumeric(t *testing.T) {
	_, err := ParseMoneyX10000("abc")
	if err == nil {
		t.Fatal("expected error for non-numeric, got nil")
	}
}

func TestParseMoneyX10000_TrailingDot(t *testing.T) {
	_, err := ParseMoneyX10000("1.")
	if err == nil {
		t.Fatal("expected error for trailing dot, got nil")
	}
}

func TestParseMoneyX10000_LeadingDot(t *testing.T) {
	v, err := ParseMoneyX10000(".5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 5000 {
		t.Fatalf("expected 5000, got %d", v)
	}
}

func TestFormatMoneyX10000(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"12.34", 123400, "12.34"},
		{"1", 10000, "1"},
		{"0.0001", 1, "0.0001"},
		{"0.12 trailing zero", 1200, "0.12"},
		{"0", 0, "0"},
		{"100", 1000000, "100"},
		{"0.001", 10, "0.001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMoneyX10000(tt.input)
			if got != tt.want {
				t.Errorf("FormatMoneyX10000(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
