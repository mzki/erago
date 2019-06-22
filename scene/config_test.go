package scene

import (
	"strings"
	"testing"
)

func TestConfigReplaceTextEmptyValidate(t *testing.T) {
	replace := ConfigReplaceText{}

	if err := replace.Validate(); err != nil {
		t.Fatalf("empty replace must be accepted, but %v", err)
	}
}

func TestConfigReplaceTextPlainTextValidate(t *testing.T) {
	replace := ConfigReplaceText{
		LoadingMessage: strings.Repeat("a", MaxReplaceTextLen),
		NewGame:        strings.Repeat("b", MaxReplaceTextLen),
		LoadGame:       strings.Repeat("c", MaxReplaceTextLen),
		QuitGame:       strings.Repeat("d", MaxReplaceTextLen),
		ReturnMenu:     strings.Repeat("e", MaxReplaceTextLen),
	}

	if err := replace.Validate(); err != nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}

	// chech error case for each field separately because struct fields are not iteratable.
	invalidLenText := strings.Repeat("a", MaxReplaceTextLen+1)
	replace = ConfigReplaceText{
		LoadingMessage: invalidLenText,
	}
	if err := replace.Validate(); err == nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}
	replace = ConfigReplaceText{
		NewGame: invalidLenText,
	}
	if err := replace.Validate(); err == nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}
	replace = ConfigReplaceText{
		LoadGame: invalidLenText,
	}
	if err := replace.Validate(); err == nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}
	replace = ConfigReplaceText{
		QuitGame: invalidLenText,
	}
	if err := replace.Validate(); err == nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}
	replace = ConfigReplaceText{
		QuitGame: invalidLenText,
	}
	if err := replace.Validate(); err == nil {
		t.Errorf("the text length %d should be accepted", MaxReplaceTextLen)
	}
}

func TestConfigReplaceTextMoneyFormatValidate(t *testing.T) {
	correct_cases := []string{
		"12345",
		"12345円",
		"12345 円",
		"￥12345",
		"12345$",
		"$12345",
		"€12345",
		"abcded12345",
		"12345678",
		"12345" + strings.Repeat("a", MaxReplaceTextLen-10),
		strings.Repeat("a", MaxReplaceTextLen-10) + "12345",
	}
	for _, format := range correct_cases {
		parsed, err := ParseMoneyFormat(format)
		if err != nil {
			t.Errorf("%s should be accpeted, but %v", format, err)
		}
		replace := ConfigReplaceText{
			MoneyFormat: parsed,
		}
		if err := replace.Validate(); err != nil {
			t.Errorf("should be accepted after parsed money format, format: %q", parsed)
		}
	}

	wrong_cases := []string{
		strings.Repeat("a", MaxReplaceTextLen-9) + "12345",
		"12345" + strings.Repeat("a", MaxReplaceTextLen-9),
		"1234",
		"1 2345",
		"1234 5",
		"12345 %",
		"% 12345",
		"%d 12345",
		"12345 %d",
		"12345 %s",
		"%s 12345",
		"%s 12345",
	}
	for _, format := range wrong_cases {
		parsed, err := ParseMoneyFormat(format)
		if err == nil {
			t.Errorf("%s should NOT be accpeted", format)
		}
		replace := ConfigReplaceText{
			MoneyFormat: parsed,
		}
		if err := replace.Validate(); err == nil {
			t.Errorf("should NOT be accepted parsed string %q if invalid", parsed)
		}
	}
}
