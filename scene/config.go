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
	// for boot
	LoadingMessage string

	// for title
	NewGame  string
	LoadGame string
	QuitGame string

	// for shop
	ReturnMenu  string
	MoneyFormat string

	// for save/load
	SelectSaveData   string
	SelectLoadData   string
	ConfirmOverwrite string
}

const (
	// Max length for replace text length used for plain text.
	// Affects to LoadingMessage.
	MaxReplacePlainTextLen = 32
	// Max length for replace text length used for command. -5 means the length of command prefix "[NN] ".
	// Affects to NewGame, LoadGame, QuitGame and ReturnMenu.
	MaxReplaceCmdTextLen = DefaultPrintCWidth - 5
)

func (c *ConfigReplaceText) Validate() error {
	// Palin text
	for _, text := range []string{
		c.LoadingMessage,
		c.SelectSaveData,
		c.SelectLoadData,
		c.ConfirmOverwrite,
	} {
		if width.StringWidth(text) > MaxReplacePlainTextLen {
			return fmt.Errorf("text length should be < %d for %q", MaxReplacePlainTextLen, text)
		}
	}
	// Command text
	for _, text := range []string{
		c.NewGame,
		c.LoadGame,
		c.QuitGame,
		c.ReturnMenu,
	} {
		if width.StringWidth(text) > MaxReplaceCmdTextLen {
			return fmt.Errorf("text length should be < %d for %q", MaxReplaceCmdTextLen, text)
		}
	}
	// Format string
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
	if filled := fmt.Sprintf(format, 1234567890); width.StringWidth(filled) > MaxReplaceCmdTextLen {
		return fmt.Errorf("text length should be < %d for %q", MaxReplaceCmdTextLen-10, format)
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
