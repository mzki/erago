package csv

import (
	"errors"
)

// alias descrilbs as like;
//	# Alias, Original
//  ...
//

// return alias map.
func readAliases(file string) (map[string]string, error) {
	aliasMap := make(map[string]string)

	err := ReadFileFunc(file, func(record []string) error {
		if len(record) < 2 {
			return errors.New("alias: Each line must have at least 2 fields.")
		}

		alias, original := record[0], record[1]
		aliasMap[alias] = original
		return nil
	})
	return aliasMap, err
}
