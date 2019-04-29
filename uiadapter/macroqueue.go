package uiadapter

import (
	"github.com/mzki/erago/uiadapter/macro"
)

type macroQ struct {
	macro *macro.Macro
}

func newMacroQ() *macroQ {
	return &macroQ{
		macro: macro.EmptyMacro,
	}
}

func (mq macroQ) Size() int {
	return len(mq.macro.Tokens)
}

func (mq *macroQ) SetMacro(m *macro.Macro) {
	mq.macro = m
}

func (mq *macroQ) Clear() {
	mq.macro = macro.EmptyMacro
}

// return first token and get succeed?.
// if macro is empty, return false.
func (mq macroQ) GetFirst() (macro.Token, bool) {
	return mq.getFirst()
}

func (mq macroQ) getFirst() (macro.Token, bool) {
	if len(mq.macro.Tokens) == 0 {
		return macro.Token{}, false
	}
	return mq.macro.Tokens[0], true
}

// return skip found?
// it deques tokens until skip token is found.
func (mq macroQ) DequeUntilSkip() bool {
	for {
		tok, ok := mq.getFirst()
		if !ok {
			return false
		}
		if tok.Type == macro.TokenTypeSkipWait {
			return true
		}
		mq.deque()
	}
}

// get token type is number.
// it deques token until get valid token.
func (mq *macroQ) DequeCommand() (string, bool) {
	for {
		tok, ok := mq.deque()
		if !ok {
			return "", false
		}
		if tok.Type == macro.TokenTypeNumber {
			return tok.Command, true
		}
	}
	panic("macro.DequeCommand: never reached")
}

func (mq *macroQ) deque() (macro.Token, bool) {
	first, ok := mq.getFirst()
	if !ok {
		return first, false
	}
	toks := mq.macro.Tokens
	mq.macro.Tokens = toks[1:]
	return first, true
}
