package csv

import (
	"fmt"
	"github.com/mzki/erago/util"
	"strconv"
	"strings"
)

// Replacing a part of system scene flow.
type GameReplace struct {
	Currency    string
	CurrencyPos string

	LoadingMessage string

	DefaultComAble bool

	SoldItemNum int

	StainLvs []int64
	ParamLvs []int64
	ExpLvs   []int64

	PBandIndex int
}

func newGameReplace(file string) (*GameReplace, error) {
	rp := &GameReplace{
		Currency:       "円",
		CurrencyPos:    "後",
		LoadingMessage: "Now Loading...",
		DefaultComAble: true,
		SoldItemNum:    100,
		StainLvs:       []int64{0, 0, 2, 1, 4},
		ParamLvs: []int64{0, 100, 500, 3000, 10000, 30000, 60000, 100000,
			150000, 250000, 500000, 1000000, 5000000, 10000000},
		ExpLvs: []int64{0, 1, 4, 20, 50, 200, 400,
			700, 1000, 1500, 2000},
		PBandIndex: 4,
	}
	err := ReadFileFunc(file, func(record []string) error {
		i := 0
		var err error = nil
		switch record[0] {
		case "お金の単位":
			rp.Currency = record[1]
		case "単位の位置":
			rp.CurrencyPos = record[1]
		case "起動時簡略表示":
			rp.LoadingMessage = record[1]
		case "COM_ABLE初期値":
			var b bool
			b, err = strconv.ParseBool(record[1])
			rp.DefaultComAble = b
		case "販売アイテム数":
			i, err = strconv.Atoi(record[1])
			rp.SoldItemNum = i
		case "汚れの初期値":
			rp.StainLvs, err = levelsFrom(record[1])
		case "PALAMLVの初期値":
			rp.ParamLvs, err = levelsFrom(record[1])
		case "EXPLVの初期値":
			rp.ExpLvs, err = levelsFrom(record[1])
		case "PBANDの初期値":
			i, err = strconv.Atoi(record[1])
			rp.PBandIndex = i
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
	if ok := util.FileExists(file); !ok {
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
