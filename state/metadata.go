package state

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"local/erago/state/csv"
	"local/erago/util"
)

// MetaData is saved with game data.
// it is referered to validate reading file.
type MetaData struct {
	Identifier  string
	GameVersion int32
	Title       string
}

const (
	saveFileIdent    = "erago"
	saveFileIdentLen = 5

	headerTitleLimit = 120 // 30 * 4byte char
)

var (
	ErrHeaderTitleTooLarge = errors.New("セーブデータのコメントが長過ぎます")

	ErrUnknownIdentifier = errors.New("正しいセーブデータではありません")
	ErrDifferentVersion  = errors.New("バージョンが異なります")
)

// return error if invalid
func (md MetaData) validate(csv *csv.CsvManager) error {
	if md.Identifier != saveFileIdent {
		return ErrUnknownIdentifier
	}
	if md.GameVersion != csv.GameBase.Version {
		return ErrDifferentVersion
	}
	return nil
}

// write header to io.Writer
func (md *MetaData) write(w io.Writer) error {
	ewriter := util.NewErrWriter(w)
	ewriter.Write([]byte(md.Identifier)) // fixed len 5B

	if bver, err := int32ToBytes(md.GameVersion); err != nil {
		return err
	} else {
		ewriter.Write(bver) // int32 4B
	}

	// variable size, prefixing bite length
	btitle := []byte(md.Title)
	blen, err := int32ToBytes(int32(len(btitle)))
	if err != nil {
		return err
	}
	if len(blen) > headerTitleLimit {
		return ErrHeaderTitleTooLarge
	}

	ewriter.Write(blen)
	ewriter.Write(btitle)

	return ewriter.Err()
}

// read into md from io.Reader
// handled errors are just io problems.
// validation of md is other task.
func (md *MetaData) read(r io.Reader) error {
	buf := make([]byte, headerTitleLimit)
	ereader := util.NewErrReader(r)

	buf = buf[:saveFileIdentLen] // 5bytes
	ereader.Read(buf)
	md.Identifier = string(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	md.GameVersion, _ = bytesToInt32(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	blen, _ := bytesToInt32(buf)
	if headerTitleLimit < blen {
		blen = headerTitleLimit
	}

	buf = buf[:blen] // variable size
	ereader.Read(buf)
	md.Title = string(buf)

	return ereader.Err()
}

var binaryEndian = binary.LittleEndian

func int32ToBytes(num int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binaryEndian, num)
	return buf.Bytes(), err
}

func bytesToInt32(p []byte) (int32, error) {
	var num int32
	r := bytes.NewReader(p)
	err := binary.Read(r, binaryEndian, &num)
	return num, err
}
