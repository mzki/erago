package csv

import (
	"errors"

	"local/erago/util"
)

// alias descrilbs as like;
//	# Alias, Original
//  ...
//

// return alias map.
func readAliases(file string) (map[string]string, error) {
	aliasMap := make(map[string]string)

	// not existing return empty.
	if exist := util.FileExists(file); !exist {
		return aliasMap, nil
	}

	err := readCsv(file, func(record []string) error {
		if len(record) < 2 {
			return errors.New("alias: Each line must have at least 2 fields.")
		}

		alias, original := record[0], record[1]
		aliasMap[alias] = original
		return nil
	})
	return aliasMap, err
}
