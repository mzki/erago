// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strings"
)

const (
	InputFileName  = "./css-color-names.json"
	OutputFileName = "./colortable.go"
)

func main() {
	table, err := loadTable(InputFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	src, err := writeTable(table)
	if err != nil {
		fmt.Println(err)
		return
	}

	fp, err := os.Create(OutputFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fp.Close()

	if _, err := fp.Write(src); err != nil {
		fmt.Println(err)
		return
	}
}

func loadTable(fname string) (map[string]interface{}, error) {
	fp, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	table := make(map[string]interface{})
	if err := json.NewDecoder(fp).Decode(&table); err != nil {
		return nil, err
	}
	return table, nil
}

const sourcePrefix = `// generated by make-colortable.go and css-color-name.json
// json file is retrieved from https://github.com/bahamas10/css-color-names/
// DO NOT EDIT

package attribute
`

func writeTable(table map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(sourcePrefix + `
// html color name to hex color table
var HTMLColorTable = map[string]uint32 {
	`)

	for name, color := range table {
		hex := parse(color)
		buf.WriteString(fmt.Sprintf("\"%s\": %s,\n", name, hex))
	}

	buf.WriteString(`
}
`)

	src := buf.Bytes()
	return format.Source(src)
}

func parse(any interface{}) string {
	s := any.(string)
	return strings.Replace(s, "#", "0x", 1)
}