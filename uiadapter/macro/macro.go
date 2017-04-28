// macro syntax is that
// 1. starts with TokenSeparator,
// 2. "e" after TokenSeparator such as "\e" skips any wait() command.
// 3. "n[number]" after TokenSeparator such as "\n100" represents inputting number 100.
//
// Thus, valid macro "\e\n100" means
// 1. skip any wait until input required,
// 2. then input number 100.
package macro

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Macro is command sequence.
// Each command is represented as token.
type Macro struct {
	Tokens []Token
}

var (
	// empty macro has tokens are zero length.
	EmptyMacro *Macro = &Macro{make([]Token, 0)}
)

// return new macro that has one skip token.
func OneSkip() *Macro {
	return &Macro{[]Token{Token{Type: TokenTypeSkipWait}}}
}

// Token is a unit of macro.
type Token struct {
	Command string
	Type    TokenType
}

// represents type of token
type TokenType int8

const (
	TokenTypeNone TokenType = iota
	TokenTypeSkipWait
	TokenTypeNumber
)

func (tt TokenType) new(cmd string) Token {
	return Token{
		Command: cmd,
		Type:    tt,
	}
}

const (
	// represents token separator. e.g. "\Token1\Token2"
	TokenSeparator = `\`

	// represents skip any wait().
	TokenSkipWait = "e"

	// represents input number e.g. "n100" means 100.
	TokenNumberPrefix = "n"
)

// the error indicates that command is not macro.
var ErrorNotMacro = errors.New("command is not macro")

// parse command as Macro and return it.
// it also retrun error parsed correctly?
// Macro is nil if error occured.
func Parse(command string) (*Macro, error) {
	if !strings.HasPrefix(command, TokenSeparator) {
		return nil, ErrorNotMacro
	}

	token_strs := strings.Split(command, TokenSeparator)
	token_strs = token_strs[1:] // [0] is always ""

	tokens := make([]Token, 0, len(token_strs))
	for i, tok_str := range token_strs {
		token, err := parseToken(tok_str)
		if err != nil && err != ErrorEmptyToken {
			return nil, fmt.Errorf("token:%d: %v", i+1, err)
		}
		tokens = append(tokens, token)
	}
	return &Macro{tokens}, nil
}

// error Token is empty
var ErrorEmptyToken = errors.New("empry token")

func parseToken(tok_str string) (Token, error) {
	if len(tok_str) == 0 {
		return Token{}, ErrorEmptyToken
	}

	switch prefix := string(tok_str[0]); prefix {
	case TokenSkipWait:
		return TokenTypeSkipWait.new(tok_str), nil

	case TokenNumberPrefix:
		nstr := strings.TrimPrefix(tok_str, TokenNumberPrefix)
		if _, err := strconv.Atoi(nstr); err != nil {
			return Token{}, errors.New(nstr + " is not a int number")
		}
		return TokenTypeNumber.new(nstr), nil

	default:
		return Token{}, errors.New(tok_str + " is invalid token")
	}
	panic("never reached")
}
