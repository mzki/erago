package state

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ugorji/go/codec"

	"local/erago/state/csv"
	"local/erago/util"
)

// file header is saved with game data.
// it refered to validate reading file.
type FileHeader struct {
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
	errHeaderTitleTooLarge = errors.New("セーブデータのコメントが長過ぎます")

	errUnknownIdentifier = errors.New("正しいセーブデータではありません")
	errDifferentVersion  = errors.New("バージョンが異なります")
)

// return error if invalid
func (fh FileHeader) validate(csv *csv.CsvManager) error {
	if fh.Identifier != saveFileIdent {
		return errUnknownIdentifier
	}
	if fh.GameVersion != csv.GameBase.Version {
		return errDifferentVersion
	}
	return nil
}

// write header to io.Writer
func (header *FileHeader) write(w io.Writer) error {
	ewriter := util.NewErrWriter(w)
	ewriter.Write([]byte(header.Identifier)) // fixed len 5B

	if bver, err := int32ToBytes(header.GameVersion); err != nil {
		return err
	} else {
		ewriter.Write(bver) // int32 4B
	}

	// variable size, prefixing bite length
	btitle := []byte(header.Title)
	blen, err := int32ToBytes(int32(len(btitle)))
	if err != nil {
		return err
	}
	if len(blen) > headerTitleLimit {
		return errHeaderTitleTooLarge
	}

	ewriter.Write(blen)
	ewriter.Write(btitle)

	return ewriter.Err()
}

// read into header from io.Reader
// handled errors are just io problems.
// validation of header is other task.
func (header *FileHeader) read(r io.Reader) error {
	buf := make([]byte, headerTitleLimit)
	ereader := util.NewErrReader(r)

	buf = buf[:saveFileIdentLen] // 5bytes
	ereader.Read(buf)
	header.Identifier = string(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	header.GameVersion, _ = bytesToInt32(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	blen, _ := bytesToInt32(buf)
	if headerTitleLimit < blen {
		blen = headerTitleLimit
	}

	buf = buf[:blen] // variable size
	ereader.Read(buf)
	header.Title = string(buf)

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

// write header which fields are setted by given data. return write error.
func writeHeaderFromData(fp io.Writer, data GameState) error {
	header := &FileHeader{
		GameVersion: int32(data.CSV.GameBase.Version),
		Identifier:  saveFileIdent,
		Title:       data.SaveComment,
	}
	return header.write(fp)
}

// read header from fp and validate it. return read header and validation result.
func readAndCheckHeaderByData(fp io.Reader, data *GameState) (*FileHeader, error) {
	header := &FileHeader{}
	if err := header.read(fp); err != nil {
		return nil, err
	}
	if err := header.validate(data.CSV); err != nil {
		if err != errDifferentVersion { // TODO: notify error of different Version?
			return nil, err
		}
	}
	return header, nil
}

const (
	defaultSavePrefix = "save"
	defaultSaveExt    = ".sav"

	shareSaveFileName = "share" + defaultSaveExt
)

func defaultFileOf(No int) string {
	return defaultSavePrefix + fmt.Sprintf("%02d", No) + defaultSaveExt
}

// make directory of given path. if exist do nothing.
func mkdirPath(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

// load only header from file prefixed No and Default.
func loadHeaderNo(No int, data *GameState) (*FileHeader, error) {
	file := defaultFileOf(No)
	return loadHeader(file, data)
}

// Load only header by file name
func loadHeader(file string, data *GameState) (*FileHeader, error) {
	file = data.config.savePath(file)
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return readAndCheckHeaderByData(fp, data)
}

// implements local/erago/state.Repository
type FileRepository struct {
	config Config
}

func NewFileRepository(config Config) *FileRepository {
	return &FileRepository{config}
}

func (repo *FileRepository) Exist(ctx context.Context, id int) bool {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))
	return util.FileExists(path)
}

func (repo *FileRepository) SaveSystemData(ctx context.Context, id int, state *GameState) error {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))
	return saveSystemData(path, *state)
}

func (repo *FileRepository) LoadSystemData(ctx context.Context, id int, state *GameState) error {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))
	_, err := loadSystemData(path, state)
	return err
}

func (repo *FileRepository) SaveShareData(ctx context.Context, state *GameState) error {
	// context is not used.
	path := repo.config.savePath(shareSaveFileName)
	return saveShareData(path, *state)
}

func (repo *FileRepository) LoadShareData(ctx context.Context, state *GameState) error {
	// context is not used.
	path := repo.config.savePath(shareSaveFileName)
	return loadShareData(path, state)
}

func (repo *FileRepository) LoadMetaList(ctx context.Context, ids ...int) ([]*FileHeader, error) {
	// context is not used.

	// fetch metalist
	metalist := make([]*FileHeader, 0, len(ids))
	for _, id := range ids {
		path := repo.config.savePath(defaultFileOf(id))
		header, err := loadHeader(path, &GameState{} /* TODO use config only */)
		if err != nil {
			return nil, fmt.Errorf("state: failed to fetch meta data for %s, err: %v", path, err)
		}

		metalist = append(metalist, header)
	}

	return metalist, nil
}

// save share data to file
func saveShareData(file string, data GameState) error {
	file = data.config.savePath(file)
	if err := mkdirPath(file); err != nil {
		return err
	}
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := writeHeaderFromData(fp, data); err != nil {
		return err
	}
	return serialize(fp, data.ShareData)
}

// load shared data from file
func loadShareData(file string, data *GameState) error {
	file = data.config.savePath(file)
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err := readAndCheckHeaderByData(fp, data); err != nil {
		return err
	}
	return deserialize(fp, data.ShareData)
}

// save System data to file using No. and default file prefix.
func saveSystemDataNo(No int, data GameState) error {
	file := defaultFileOf(No)
	return saveSystemData(file, data)
}

// load System data from file using No. and default file prefix.
func loadSystemDataNo(No int, data *GameState) (*FileHeader, error) {
	file := defaultFileOf(No)
	return loadSystemData(file, data)
}

// save game system data to file.
func saveSystemData(file string, data GameState) error {
	file = data.config.savePath(file)
	if err := mkdirPath(file); err != nil {
		return err
	}
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := writeHeaderFromData(fp, data); err != nil {
		return err
	}
	return serialize(fp, data.SystemData) // return encode ok?
}

// Load game system data from file.
func loadSystemData(file string, data *GameState) (*FileHeader, error) {
	file = data.config.savePath(file)
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	header, err := readAndCheckHeaderByData(fp, data)
	if err != nil {
		return header, err
	}

	data.LastLoadVer = header.GameVersion
	data.LastLoadComment = header.Title

	return header, deserialize(fp, data.SystemData)
}

var (
	codecHandler = &codec.MsgpackHandle{RawToString: true}
)

func serialize(w io.Writer, data interface{}) error {
	enc := codec.NewEncoder(w, codecHandler)
	return enc.Encode(data)
}

func deserialize(r io.Reader, data interface{}) error {
	dec := codec.NewDecoder(r, codecHandler)
	return dec.Decode(data) // return encode ok?
}

func ExistsSaveFileNo(No int, state *GameState) bool {
	path := state.config.savePath(defaultFileOf(No))
	return util.FileExists(path)
}
