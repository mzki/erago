package csv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mzki/erago/filesystem"
)

// Number constants excepted from CSV names.
type NumberConstants struct {
	ParamLvs []int64
	ExpLvs   []int64
}

func newNumberConstants(file string) (*NumberConstants, error) {
	rp := &NumberConstants{
		ParamLvs: []int64{0, 100, 500, 3000, 10000, 30000, 60000, 100000,
			150000, 250000, 500000, 1000000, 5000000, 10000000},
		ExpLvs: []int64{0, 1, 4, 20, 50, 200, 400,
			700, 1000, 1500, 2000},
	}
	err := ReadFileFunc(file, func(record []string) error {
		var err error = nil
		switch record[0] {
		case "PALAMLVの初期値", "PARAMLV":
			rp.ParamLvs, err = levelsFrom(record[1])
		case "EXPLVの初期値", "EXPLV":
			rp.ExpLvs, err = levelsFrom(record[1])
		default:
			return fmt.Errorf("unknown record name %v", record)
		}
		return err
	})
	return rp, err
}

func levelsFrom(s string) ([]int64, error) {
	fragments := strings.Split(s, "/")
	lvs := make([]int64, len(fragments)+1)
	for i, fragment := range fragments {
		num, err := strconv.ParseInt(fragment, 0, 64)
		if err != nil {
			return nil, err
		}
		lvs[i] = num
	}
	return lvs, nil
}

// Base information for the Game
type GameBase struct {
	Code           string
	Version        int32
	Title          string
	Author         string
	Develop        string
	AllowDiffVer   bool
	AdditionalInfo string
}

func newGameBase(file string) (*GameBase, error) {
	gb := &GameBase{}

	// not found, then return empty.
	if ok := filesystem.Exist(file); !ok {
		return gb, nil
	}

	err := ReadFileFunc(file, func(record []string) error {
		switch key := record[0]; key {
		case "コード":
			gb.Code = record[1]
		case "バージョン":
			ver, err := strconv.ParseInt(record[1], 10, 32)
			if err != nil {
				return fmt.Errorf("version is required int number: %s", record[1])
			}
			gb.Version = int32(ver)
		case "タイトル":
			gb.Title = record[1]
		case "作者":
			gb.Author = record[1]
		case "製作年":
			gb.Develop = record[1]
		case "バージョン違い認める":
			if strings.Contains(record[1], "はい") {
				gb.AllowDiffVer = true
			} else {
				gb.AllowDiffVer = false
			}
		case "追加情報":
			gb.AdditionalInfo = record[1]
		default:
			return fmt.Errorf("unknown GameBase key: %s", key)
		}
		return nil
	})
	return gb, err
}
