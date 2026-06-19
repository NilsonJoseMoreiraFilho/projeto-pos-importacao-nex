package xlsx

import "testing"

func TestParseDecimalAcceptsBrazilianMoneyFormat(t *testing.T) {
	value, err := parseDecimal("1.234,56")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if value != 1234.56 {
		t.Fatalf("unexpected value: %.2f", value)
	}
}

func TestParseDecimalRejectsInvalidValue(t *testing.T) {
	_, err := parseDecimal("abc")
	if err == nil {
		t.Fatal("expected parse error")
	}
}
