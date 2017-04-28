package csv

import (
	"fmt"
	"local/erago/util"
	"strconv"
	"strings"
)

// DEPLECATED: no longer used
type Replace struct {
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

func newReplace(file string) (*Replace, error) {
	rp := &Replace{
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
	err := readCsv(file, func(record []string) error {
		i := 0
		var err error = nil
		switch record[0] {
		case "お金の単位":
			rp.Currency = record[1]
		case "単位の位置":
			rp.CurrencyPos = record[1]
		case "起動時簡略表示":
			rp.LoadingMessage = record[1]
		case "COM_ABLEの初期値":
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

type GameBase struct {
	Code           string
	Version        int32
	Title          string
	Author         string
	Develop        string
	DifferentVer   string
	AdditionalInfo string
}

func newGameBase(file string) (*GameBase, error) {
	gb := &GameBase{}

	// not found, then return empty.
	if ok := util.FileExists(file); !ok {
		return gb, nil
	}

	err := readCsv(file, func(record []string) error {
		switch record[0] {
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
		case "制作年":
			gb.Develop = record[1]
		case "バージョン違い認める":
			gb.DifferentVer = record[1]
		case "追加情報":
			gb.AdditionalInfo = record[1]
		}
		return nil
	})
	return gb, err
}
