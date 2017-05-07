package csv

import (
	"fmt"
	"strconv"
)

// Character defined by csv
type Character struct {
	CharaInfo
	Parameter
}

func newCharacter(cm *CsvManager) *Character {
	c := &Character{
		Parameter: newParameter(cm),
	}
	return c
}

type CharaInfo struct {
	ID int64

	Name       string // formal name
	CallName   string // name when is called
	NickName   string // friendly name
	MasterName string // call for you
}

// Parameter is variables which are numbers or strings.
type Parameter struct {
	strMap map[string][]string
	intMap map[string][]int64
}

// default Parameter instance
func newParameter(cm *CsvManager) Parameter {
	p := Parameter{}
	p.intMap = cm.BuildIntUserVars(ScopeChara)
	p.strMap = cm.BuildStrUserVars(ScopeChara)
	return p
}

// return string variable map which can affect original values.
func (p Parameter) GetStrMap() map[string][]string {
	return p.strMap
}

// return int64 variable map which can affect original values.
func (p Parameter) GetIntMap() map[string][]int64 {
	return p.intMap
}

// it returns new data conserving contents pf given data.
func CloneInt64(data []int64) []int64 {
	new_data := make([]int64, len(data))
	copy(new_data, data)
	return new_data
}

// it returns new data conserving contents pf given data.
func CloneStr(data []string) []string {
	new_data := make([]string, len(data))
	copy(new_data, data)
	return new_data
}

const (
	// Parsed keys for chara's CSV.
	// 2 field record
	keyID         = "ID"
	keyNAME       = "Name"
	keyCALLNAME   = "CallName"
	keyNICKNAME   = "NickName"
	keyMASTERNAME = "MasterName"
)

// TODO: split into external file _Alias.csv
var nameAlias = map[string]string{
	// 2 field record
	"番号":     keyID,
	"名前":     keyNAME,
	"呼び名":    keyCALLNAME,
	"あだ名":    keyNICKNAME,
	"主人の呼び方": keyMASTERNAME,

	"ID":         keyID,
	"NAME":       keyNAME,
	"CALLNAME":   keyCALLNAME,
	"NICKNAME":   keyNICKNAME,
	"MASTERNAME": keyMASTERNAME,

	// 3 field record
	"能力":    "Abl",
	"経験":    "Exp",
	"快感":    "Ex",
	"基礎":    "Base",
	"文字":    "CStr",
	"刻印":    "Mark",
	"珠":     "Juel",
	"フラグ":   "CFlag",
	"装着物":   "Equip",
	"素質":    "Talent",
	"パラメータ": "Param",
	"相性":    "Relation",
	"ソース":   "Source",
	"汚れ":    "Stain",

	"ABL":      "Abl",
	"EXP":      "Exp",
	"EX":       "Ex",
	"BASE":     "Base",
	"CSTR":     "CStr",
	"MARK":     "Mark",
	"JUEL":     "Juel",
	"FLAG":     "CFlag",
	"EQUIP":    "Equip",
	"TALENT":   "Talent",
	"PARAM":    "Param",
	"RELATION": "Relation",
	"SOURCE":   "Source",
	"STAIN":    "Stain",
}

// read characrer deifinition from csv file,
// and return it.
func readCharacter(file string, cm *CsvManager) (*Character, error) {
	character := newCharacter(cm)
	err := ReadFileFunc(file, func(record []string) error {
		key := record[0]
		if alias, has := cm.aliasMap[key]; has {
			key = alias
		}

		switch key {
		case keyID:
			return parseIDField(&character.CharaInfo, record)

		case keyNAME, keyCALLNAME, keyNICKNAME, keyMASTERNAME:
			return parseStr2Field(&character.CharaInfo, key, record)

		default:
			return parseUserField(&character.Parameter, cm, key, record)
		}
	})
	return character, err
}

func parseIDField(chara *CharaInfo, record []string) error {
	id, err := strconv.ParseInt(record[1], 10, 64)
	if err != nil {
		return err
	} else if id < 0 {
		return fmt.Errorf("Must be Character's ID >= 0. But %d", id)
	}
	chara.ID = id
	return nil
}

func parseStr2Field(character *CharaInfo, key string, record []string) error {
	data := record[1]
	switch key {
	case keyNAME:
		character.Name = data
	case keyCALLNAME:
		character.CallName = data
	case keyNICKNAME:
		character.NickName = data
	case keyMASTERNAME:
		character.MasterName = data
	default:
		panic("csv.parseStr2Field: This code should not be printed")
	}
	return nil
}

func parseUserField(p *Parameter, cm *CsvManager, key string, record []string) error {
	// parse string index for the variables specified by key
	index := cm.NameIndexOf(key, record[1])
	if index < 0 {
		// string index is not defined, try parsing as int
		i, err := strconv.ParseInt(record[1], 0, 64)
		if err != nil {
			return fmt.Errorf("Second field(%v) is not defined in %v.", record[1], key)
		}
		index = int(i)
	}

	// set value for the variables specified by key.
	if vars, has := p.strMap[key]; has {
		if !InRangeStr(vars, index) {
			return fmt.Errorf("Second field %v is invalid index (Max %v).", record[1], len(vars))
		}
		vars[index] = record[2]

	} else if vars, has := p.intMap[key]; has {
		if !InRangeInt64(vars, index) {
			return fmt.Errorf("Second field %v is invalid index (Max %v).", record[1], len(vars))
		}

		num, err := strconv.ParseInt(record[2], 0, 64)
		if err != nil {
			num = 1 // NOTE default data if no defined
		}
		vars[index] = num

	} else {
		return fmt.Errorf("First field %v is not defined.", record[0])
	}
	return nil
}

// check wheather index is in valid range?
func InRangeStr(data []string, index int) bool {
	return 0 <= index && index < len(data)
}

// check wheather index is in valid range?
func InRangeInt64(data []int64, index int) bool {
	return 0 <= index && index < len(data)
}
