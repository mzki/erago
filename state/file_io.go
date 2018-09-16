package state

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ugorji/go/codec"

	"local/erago/util"
)

// write metadata which fields are setted by given data. return write error.
func writeMetaDataFromState(fp io.Writer, data GameState) error {
	metadata := &MetaData{
		GameVersion: int32(data.CSV.GameBase.Version),
		Identifier:  saveFileIdent,
		Title:       data.SaveComment,
	}
	return metadata.write(fp)
}

// read metadata from fp and validate it. return read metadata and validation result.
func readAndCheckMetaDataByState(fp io.Reader, data *GameState) (*MetaData, error) {
	metadata := &MetaData{}
	if err := metadata.read(fp); err != nil {
		return nil, err
	}
	if err := metadata.validate(data.CSV); err != nil {
		if err != ErrDifferentVersion { // TODO: notify error of different Version?
			return nil, err
		}
	}
	return metadata, nil
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

// load only metadata from file prefixed No and Default.
func loadMetaDataNo(No int, data *GameState) (*MetaData, error) {
	file := defaultFileOf(No)
	return loadMetaData(file, data)
}

// Load only metadata by file name
func loadMetaData(file string, data *GameState) (*MetaData, error) {
	file = data.config.savePath(file)
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return readAndCheckMetaDataByState(fp, data)
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

func (repo *FileRepository) LoadMetaList(ctx context.Context, ids ...int) ([]*MetaData, error) {
	// context is not used.

	// fetch metalist
	metalist := make([]*MetaData, 0, len(ids))
	for _, id := range ids {
		path := repo.config.savePath(defaultFileOf(id))
		header, err := loadMetaData(path, &GameState{} /* TODO use config only */)
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

	if err := writeMetaDataFromState(fp, data); err != nil {
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

	if _, err := readAndCheckMetaDataByState(fp, data); err != nil {
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
func loadSystemDataNo(No int, data *GameState) (*MetaData, error) {
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

	if err := writeMetaDataFromState(fp, data); err != nil {
		return err
	}
	return serialize(fp, data.SystemData) // return encode ok?
}

// Load game system data from file.
func loadSystemData(file string, data *GameState) (*MetaData, error) {
	file = data.config.savePath(file)
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	metadata, err := readAndCheckMetaDataByState(fp, data)
	if err != nil {
		return metadata, err
	}

	data.LastLoadVer = metadata.GameVersion
	data.LastLoadComment = metadata.Title

	return metadata, deserialize(fp, data.SystemData)
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
