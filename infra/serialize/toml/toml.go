package toml

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/mzki/erago/util/log"
)

// encode data to Writer.
func Encode(w io.Writer, data interface{}) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(data)
}

// encode data to file.
func EncodeFile(file string, data interface{}) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	return Encode(fp, data)
}

// decode from reader and store it to data.
func Decode(r io.Reader, data interface{}) error {
	meta, err := toml.DecodeReader(r, data)
	if undecoded := meta.Undecoded(); undecoded != nil && len(undecoded) > 0 {
		log.Infoln("toml.Decode:", "undecoded keys exist,", undecoded)
	}
	return err
}

// decode from file and store it to data.
func DecodeFile(file string, data interface{}) error {
	meta, err := toml.DecodeFile(file, data)
	if undecoded := meta.Undecoded(); undecoded != nil && len(undecoded) > 0 {
		log.Infoln("toml.DecodeFile:", "undecoded keys exist,", undecoded)
	}
	return err
}