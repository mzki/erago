package repo

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ugorji/go/codec"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/state"
	"github.com/mzki/erago/state/csv"
	"github.com/mzki/erago/util/errutil"
)

const (
	defaultSavePrefix = "save"
	defaultSaveExt    = ".sav"

	shareSaveFileName = "share" + defaultSaveExt
)

func defaultFileOf(No int) string {
	return defaultSavePrefix + fmt.Sprintf("%02d", No) + defaultSaveExt
}

// implements local/erago/state.Repository
type FileRepository struct {
	config     Config
	expectMeta state.MetaData
}

func NewFileRepository(csvdb *csv.CsvManager, config Config) *FileRepository {
	return &FileRepository{
		config: config,
		expectMeta: state.MetaData{
			Identifier:  state.DefaultMetaIdent,
			GameVersion: csvdb.GameBase.Version,
			Title:       "", // not used
		},
	}
}

func (repo *FileRepository) Exist(ctx context.Context, id int) bool {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))
	return filesystem.Exist(path)
}

// save game system data to file.
func (repo *FileRepository) SaveSystemData(ctx context.Context, id int, gs *state.GameState) error {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))

	fp, err := filesystem.Store(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	// save metadata with comment
	var metadata state.MetaData = repo.expectMeta // deep copy
	metadata.Title = gs.SaveInfo.SaveComment
	if err := writeMetaDataTo(fp, &metadata); err != nil {
		return err
	}

	// save system data
	return serialize(fp, gs.SystemData) // return encode ok?
}

// Load game system data from file.
func (repo *FileRepository) LoadSystemData(ctx context.Context, id int, gs *state.GameState) error {
	// context is not used.
	path := repo.config.savePath(defaultFileOf(id))

	fp, err := filesystem.Load(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	metadata, err := readAndCheckMetaDataByState(fp, repo.expectMeta)
	if err != nil {
		return err
	}

	gs.SaveInfo.LastLoadVer = metadata.GameVersion
	gs.SaveInfo.LastLoadComment = metadata.Title

	return deserialize(fp, gs.SystemData)
}

// save share data to file
func (repo *FileRepository) SaveShareData(ctx context.Context, gs *state.GameState) error {
	// context is not used.
	path := repo.config.savePath(shareSaveFileName)

	fp, err := filesystem.Store(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	var metadata state.MetaData = repo.expectMeta // deep copy
	if err := writeMetaDataTo(fp, &metadata); err != nil {
		return err
	}
	return serialize(fp, gs.ShareData)
}

// load shared data from file
func (repo *FileRepository) LoadShareData(ctx context.Context, gs *state.GameState) error {
	// context is not used.
	path := repo.config.savePath(shareSaveFileName)

	fp, err := filesystem.Load(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err := readAndCheckMetaDataByState(fp, repo.expectMeta); err != nil {
		return err
	}
	return deserialize(fp, gs.ShareData)
}

func (repo *FileRepository) LoadMetaList(ctx context.Context, ids ...int) ([]*state.MetaData, error) {
	// context is not used.

	// fetch metalist
	metalist := make([]*state.MetaData, 0, len(ids))
	for _, id := range ids {
		path := repo.config.savePath(defaultFileOf(id))
		header, err := loadMetaData(path, repo.expectMeta)
		if err != nil {
			return nil, fmt.Errorf("repo: failed to fetch meta data for %s, err: %v", path, err)
		}

		metalist = append(metalist, header)
	}

	return metalist, nil
}

// Load only metadata by file path
func loadMetaData(path string, expectMeta state.MetaData) (*state.MetaData, error) {
	fp, err := filesystem.Load(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return readAndCheckMetaDataByState(fp, expectMeta)
}

// read metadata from fp and validate it. return read metadata and validation result.
func readAndCheckMetaDataByState(fp io.Reader, expectMeta state.MetaData) (*state.MetaData, error) {
	metadata := &state.MetaData{}
	if err := readMetaDataFrom(fp, metadata); err != nil {
		return nil, err
	}
	if err := validateMetaData(metadata, expectMeta); err != nil {
		if err != state.ErrDifferentVersion { // TODO: notify error of different Version?
			return nil, err
		}
	}
	return metadata, nil
}

// return error if invalid
func validateMetaData(md *state.MetaData, expectMeta state.MetaData) error {
	if md.Identifier != expectMeta.Identifier {
		return state.ErrUnknownIdentifier
	}
	if md.GameVersion != expectMeta.GameVersion {
		return state.ErrDifferentVersion
	}
	// Title is ignored for validation
	return nil
}

// writeMetaDataTo writes metadata into io.Writer
func writeMetaDataTo(w io.Writer, md *state.MetaData) error {
	ewriter := errutil.NewErrWriter(w)
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
	if len(blen) > state.MetaTitleLimit {
		return state.ErrTitleTooLarge
	}

	ewriter.Write(blen)
	ewriter.Write(btitle)

	return ewriter.Err()
}

// readMetaDataFrom reads metadata from io.Reader
// handled errors are just io problems.
// validation of md is other task.
func readMetaDataFrom(r io.Reader, md *state.MetaData) error {
	buf := make([]byte, state.MetaTitleLimit)
	ereader := errutil.NewErrReader(r)

	buf = buf[:state.DefaultMetaIdentLen] // 5bytes
	ereader.Read(buf)
	md.Identifier = string(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	md.GameVersion, _ = bytesToInt32(buf)

	buf = buf[:4] // 4bytes
	ereader.Read(buf)
	blen, _ := bytesToInt32(buf)
	if state.MetaTitleLimit < blen {
		blen = state.MetaTitleLimit
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

var (
	codecHandler = &codec.MsgpackHandle{}
)

func serialize(w io.Writer, data interface{}) error {
	enc := codec.NewEncoder(w, codecHandler)
	return enc.Encode(data)
}

func deserialize(r io.Reader, data interface{}) error {
	dec := codec.NewDecoder(r, codecHandler)
	return dec.Decode(data) // return encode ok?
}
