package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, `
Usage: %v text-file
  this command will print string from binary result shown at console when Zip archive testing failed,
  e.g. ArchiveAsZipWriter() = [80 75 3 4 2]
  The argument should be text file containing binary result without square bracket.
  That means text file content should be "80 75 3 4 2" as an example.
`, os.Args[0])
		os.Exit(1)
	}
	file := os.Args[1]
	str, err := dumpAsString(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed dump as string: %v", err)
	}
	fmt.Println(str)
	os.Exit(0)
}

func dumpAsString(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	content := string(bs)

	content = strings.ReplaceAll(content, "\n", "")
	bytesAsStr := strings.Split(content, " ")
	byteList := make([]byte, 0, 32)
	for _, s := range bytesAsStr {
		bi, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return "", err
		}
		byteList = append(byteList, byte(bi))
	}

	return string(byteList), nil
}
