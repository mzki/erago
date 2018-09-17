package state

import (
	"errors"
)

// MetaData is saved with game data.
// it is referered to validate reading file.
type MetaData struct {
	Identifier  string
	GameVersion int32
	Title       string
}

const (
	DefaultMetaIdent    = "erago"
	DefaultMetaIdentLen = 5

	MetaTitleLimit = 120 // 30 * 4byte char
)

var (
	ErrTitleTooLarge     = errors.New("state: comment in metadata is too long")
	ErrUnknownIdentifier = errors.New("state: signature in metadata is not correct")
	ErrDifferentVersion  = errors.New("state: different game version in metadata")
)
