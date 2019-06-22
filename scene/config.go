package scene

import (
	"fmt"
	"strings"

	"github.com/mzki/erago/width"
)

// Scene Configure
type Config struct {
	// can be auto-saved in the specific scene transition
	CanAutoSave bool
}

// ConfigReplaceText holds strings for replace specific text in the bultin scenes.
// empty string is treated as no replace (use builtin).
type ConfigReplaceText struct {
	LoadingMessage string
	NewGame        string
	LoadGame       string
	QuitGame       string
	ReturnMenu     string
	MoneyFormat    string
}

const (
	// Max length for replace text length. -5 means the length of command prefix "[NN] ".
	MaxReplaceTextLen = DefaultPrintCWidth - 5
)

func (c *ConfigReplaceText) Validate() error {
	for _, text := range []string{
		c.LoadingMessage,
		c.NewGame,
		c.LoadGame,
		c.QuitGame,
		c.ReturnMenu,
	} {
		if width.StringWidth(text) > MaxReplaceTextLen {
			return fmt.Errorf("text length should be < %d for %q", MaxReplaceTextLen, text)
		}
	}
	if err := validateMoneyFormat(c.MoneyFormat); err != nil {
		return err
	}
	return nil
}

const (
	// PlaceHolderNumber is a place holder for a number instead of "%d".
	PlaceHolderNumber   = "12345"
	replaceNumberFormat = "%d"
)

// ParseMoneyFormat validates source string has
// correct monery format and converts it to be used on scene flow.
func ParseMoneyFormat(src string) (string, error) {
	parsed_text := strings.Replace(src, PlaceHolderNumber, replaceNumberFormat, 1)
	return parsed_text, validateMoneyFormat(parsed_text)
}

// ValidateMoneyFormat validates given string has correct money format?
func validateMoneyFormat(format string) error {
	// special case: no replace if empty.
	if len(format) == 0 {
		return nil
	}

	// check placement for a format string "%d"
	if c := strings.Count(format, "%"); c != 1 {
		return fmt.Errorf("containing %% is not allowed in %q", format)
	}
	if i := strings.Index(format, "%d"); i < 0 {
		return fmt.Errorf("should contain exactly one %%d in %q", format)
	}

	// check text length with using 10 digits
	if filled := fmt.Sprintf(format, 1234567890); width.StringWidth(filled) > MaxReplaceTextLen {
		return fmt.Errorf("text length should be < %d for %q", MaxReplaceTextLen-10, format)
	}
	return nil
}

// DefaultOrString is a helper function for string replacement.
// It returns _default text if replacement text is likely empty,
// otherwise returns replacement text
func DefaultOrString(_default string, replacement string) string {
	if len(replacement) == 0 {
		return _default
	}
	return replacement
}
