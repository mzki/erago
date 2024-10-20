// Package csv provides csv-parser for game parameter names.
// The parser read csv file and store internal data.
//
package csv

import (
	"fmt"
	"strings"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/util/errutil"
)

const (
	// user defines user variables in this file.
	variableSpecFile = "VariableSpec.csv"

	// BuiltinXXXName is builtin variable names which always exist in system.
	// User can modify its size only, and not use these names to user defined variable name.
	//
	// These names must exist in CsvManager,
	// but its size is unstable.
	BuiltinTrainName  = "Train"  // Scope CSV
	BuiltinSourceName = "Source" // Scope CSV

	BuiltinParamName  = "Param"  // Scope Chara
	BuiltinJuelName   = "Juel"   // Scope Chara
	BuiltinAblName    = "Abl"    // Scope Chara
	BuiltinTalentName = "Talent" // Scope Chara
	BuiltinMarkName   = "Mark"   // Scope Chara
	BuiltinExpName    = "Exp"    // Scope Chara

	BuiltinItemStockName = "ItemStock" // Scope System
	BuiltinMoneyName     = "Money"     // Scope System

	// it shares csv defined Item-Names with but separated as variable data.
	BuiltinItemName      = "Item"      // Scope System
	BuiltinItemPriceName = "ItemPrice" // Scope CSV

	exceptItemName      = BuiltinItemName
	exceptItemPriceName = BuiltinItemPriceName
)

const (
	aliasFileName           = "_Alias.csv"
	gameBaseFileName        = "_GameBase.csv"
	numberConstantsFileName = "_NumberConstants.csv"
)

// CSV Manager manages CSV variables.
//
// To use this, first initialize.:
//  csv := &CsvManager{}
//	err := csv.Initialize(Config{})
//
type CsvManager struct {
	encording string

	config Config

	// user defined csv names and indexes.
	constants map[string]Constant

	// the exceptional constants for reading csv file.
	Item       Constant
	ItemPrices []int64

	// default definition of the characters.
	// These are defined by "Chara/*.csv".
	// Each chara is identified by chara No.
	CharaMap map[int64]*Character

	// the spec of the allocating variables.
	vspecs variableSpecInternalMap

	// these are cached since character variables are referenced frequently.
	vspecsCharaInt variableSpecInternalMap
	vspecsCharaStr variableSpecInternalMap

	// some optional data, GameBase, Replace, and aliasMap,
	// are loaded from _{filename}.csv to configure
	// some constant parameters.

	// _GameBase.csv
	GameBase

	// _Replace.csv
	NumberConstants

	// alias for reading character defined csv, chara*.csv
	aliasMap map[string]string

	// Define buffers for reading csv fields.
	//
	// NOTE: These are allocated at the first of CsvManager.Initialize(),
	// and released at the last.
	// After released, accessing these occurs panic().
	readIntBuffer    []int
	readStringBuffer []string
}

// return empty CsvManager same as &CsvManager{} .
func NewCsvManager() *CsvManager {
	return &CsvManager{}
}

// get index using group name: BASE, ABL ... , and param name: 体力　... ,
// if not found returns -1
func (csv *CsvManager) NameIndexOf(group, name string) int {
	if c, has := csv.constants[group]; has {
		if i, ok := c.NameIndex[name]; ok {
			return i
		}
	}
	return -1
}

// get constant variables map. modifying that map
// breaks constant.
func (csv *CsvManager) Constants() map[string]Constant {
	if cs := csv.constants; cs != nil {
		return cs
	}
	panic("csv: get constants before not initialized CsvManager")
}

// get a Constant by a variable name. if vname is not found,
// it will panic.
func (csv *CsvManager) MustConst(vname string) Constant {
	c, err := csv.Const(vname)
	if err != nil {
		panic(err)
	}
	return c
}

// get a Constant by a variable name. return Constant and not found error.
func (csv *CsvManager) Const(vname string) (Constant, error) {
	if c, ok := csv.constants[vname]; ok {
		return c, nil
	}
	return Constant{}, fmt.Errorf("csv: constant variable(%s) is not found", vname)
}

// VarScope indicates where are the variables used.
type VarScope uint8

const (
	// System variables is system-wide global variables.
	ScopeSystem = VarScope(scopeSystem)
	// Share variables is global variables shared with other savedata.
	ScopeShare = VarScope(scopeShare)
	// Chara variables is character specific variables.
	ScopeChara = VarScope(scopeChara)
)

// VariableSpec defines the spec of user defined varables.
type VariableSpec struct {
	VarName string
	Scope   VarScope
	Size    uint64
}

// IntVariableSpecs returns slice of VariableSpecs which mathces
// given VarScope and data type Int.
func (cm *CsvManager) IntVariableSpecs(where VarScope) []VariableSpec {
	return cm.selectVariableSpecs(where, dTypeInt)
}

// StrVariableSpecs returns slice of VariableSpecs which mathces
// given VarScope and data type Str.
func (cm *CsvManager) StrVariableSpecs(where VarScope) []VariableSpec {
	return cm.selectVariableSpecs(where, dTypeStr)
}

func (cm *CsvManager) selectVariableSpecs(where VarScope, dtype vspecIdent) []VariableSpec {
	vs := cm.vspecs.selectByScopeAndDType(vspecIdent(where), dtype)
	vspecs := make([]VariableSpec, len(vs))
	var i int = 0
	for _, v := range vs {
		vspecs[i] = VariableSpec{
			VarName: v.VarName,
			Scope:   VarScope(v.Scope),
			Size:    uint64(v.Size[0]),
		}
		i++
	}
	return vspecs
}

// return variable maps, which type are DataType string and
// scope where, where = {System, Share}.
// It allocates new valiables every call.
func (cm *CsvManager) BuildStrUserVars(where VarScope) map[string][]string {
	if scope := vspecIdent(where); scope == scopeChara {
		return newStrMapByVSpecs(cm.vspecsCharaStr)
	} else {
		vs := cm.vspecs.selectByScopeAndDType(scope, dTypeStr)
		return newStrMapByVSpecs(vs)
	}
}

// return variable maps, which type are DataType int64 and
// scope where, where = {System, Share}.
// It allocates new valiables every call.
func (cm *CsvManager) BuildIntUserVars(where VarScope) map[string][]int64 {
	if scope := vspecIdent(where); scope == scopeChara {
		return newIntMapByVSpecs(cm.vspecsCharaInt)
	} else {
		vs := cm.vspecs.selectByScopeAndDType(scope, dTypeInt)
		return newIntMapByVSpecs(vs)
	}
}

func newIntMapByVSpecs(vspecs variableSpecInternalMap) map[string][]int64 {
	int_map := make(map[string][]int64, len(vspecs))
	for _, vs := range vspecs {
		int_map[vs.VarName] = make([]int64, vs.Size[0])
	}
	return int_map
}

func newStrMapByVSpecs(vspecs variableSpecInternalMap) map[string][]string {
	str_map := make(map[string][]string, len(vspecs))
	for _, vs := range vspecs {
		str_map[vs.VarName] = make([]string, vs.Size[0])
	}
	return str_map
}

// initialize by reading csv files.
func (cm *CsvManager) Initialize(config Config) (err error) {
	// initialize reading-buffer
	cm.readIntBuffer = make([]int, 0, 1000)
	cm.readStringBuffer = make([]string, 0, 1000)

	// finalize reading-buffer
	defer func() {
		cm.readIntBuffer = nil
		cm.readStringBuffer = nil
	}()

	// to prevent nil reference.
	cm.CharaMap = make(map[int64]*Character)
	cm.aliasMap = make(map[string]string)

	cm.config = config

	// load GameBase, Replace and Alias.
	{
		errs := errutil.NewMultiError()
		var err error

		if aliasFile := config.filepath(aliasFileName); FileExists(aliasFile) {
			cm.aliasMap, err = readAliases(aliasFile)
			errs.Add(err)
		}
		if file := config.filepath(gameBaseFileName); FileExists(file) {
			base, err := newGameBase(file)
			cm.GameBase = *base
			errs.Add(err)
		}
		if file := config.filepath(numberConstantsFileName); FileExists(file) {
			// TODO: DEPRECATED?: numberConstants can be placed at script layer and is adeque for the software architecure.
			numbers, err := newNumberConstants(file)
			cm.NumberConstants = *numbers
			errs.Add(err)
		}
		if err = errs.Err(); err != nil {
			return err
		}
	}

	// load user specific variables.
	{
		var all_vspecs variableSpecInternalMap = make(variableSpecInternalMap)
		var vspec_path = config.filepath(variableSpecFile)
		if FileExists(vspec_path) {
			if vs, err := readVariableSpecsFile(vspec_path); err != nil {
				return err
			} else {
				all_vspecs = vs
			}
		}
		notAppended := appendBuiltinVSpecs(all_vspecs)
		// NOTE: duplicated varnames with builtin's are not allowed.
		if len(notAppended) > 0 {
			return fmt.Errorf("Varname %v are preserved by builtin values. Rename it", notAppended)
		}

		if err := cm.initVariableSpecs(all_vspecs); err != nil {
			return err
		}
	}

	// load builtin exceptional variables
	{
		newConst, err := readConstantFile(
			config.filepath(BuiltinItemName+".csv"),
			cm.readIntBuffer,
			cm.readStringBuffer,
		)
		if err != nil {
			return fmt.Errorf("csv: can not be initialized: %v", err)
		}
		if !newConst.CustomFields.Has(HeaderFieldItemPrice) {
			return fmt.Errorf("csv: can not be initialized: %s.csv must have `%s` field, but not exist", BuiltinItemName, HeaderFieldItemPrice)
		}

		// TODO Remove struct Field Item and ItemPrices?
		cm.Item = *newConst
		cm.ItemPrices = newConst.CustomFields.MustInts(HeaderFieldItemPrice).data

		// Publish as Constant so that it is used in
		// the same mannar as the other variables.
		// cm.constants[BuiltinItemName] = cm.Item
		cm.constants[BuiltinItemName] = *newConst
	}

	// fit variable size by Constant.Names one.
	new_vspecs := cm.vspecs.Map(func(v variableSpecInternal) variableSpecInternal {
		if v.Size[0] <= 0 {
			v.Size[0] = cm.constants[v.VarName].Names.Len()
		}
		return v
	})
	cm.vspecs = new_vspecs
	cm.vspecsCharaInt = new_vspecs.selectByScopeAndDType(scopeChara, dTypeInt)
	cm.vspecsCharaStr = new_vspecs.selectByScopeAndDType(scopeChara, dTypeStr)

	// read character
	return cm.initCharacters(config.charaPattern())
}

// initVariableSpecs build csv Constants from all the variableSpecs.
func (cm *CsvManager) initVariableSpecs(all_vspecs variableSpecInternalMap) error {
	// exclude execeptional variable, Item, and ItemPrice,
	// and built rest only.
	ordinary_vspecs := all_vspecs.selectBy(func(v variableSpecInternal) bool {
		exception := v.VarName == exceptItemName || v.VarName == exceptItemPriceName
		return !exception
	})
	if err := cm.buildConstants(ordinary_vspecs); err != nil {
		return err
	}
	cm.vspecs = all_vspecs
	return nil
}

// build Names of given variableSpecs
func (cm *CsvManager) buildConstants(vspecs variableSpecInternalMap) error {
	cm.constants = make(map[string]Constant, len(vspecs)*2)

	for _, vs := range vspecs {
		fname := vs.FileName
		if len(fname) == 0 { // ignore empty file name
			continue
		}

		constant, has := cm.constants[fname]
		if !has {
			// load new csv file
			newConst, err := readConstantFile(
				cm.config.filepath(fname),
				cm.readIntBuffer,
				cm.readStringBuffer,
			)
			if err != nil {
				return err
			}
			constant = *newConst
		}
		// register Names and its indexes.
		varname := vs.VarName
		if _, has := cm.constants[varname]; has {
			return fmt.Errorf("csv: duplicate VarName (%s)", varname)
		}
		cm.constants[varname] = constant
		cm.constants[fname] = constant // csv file name is also remindered.
	}

	// remove csv file names from constants map, which are never accessed.
	// the data accessed by csv file name can be accessed by varname.
	for key, _ := range cm.constants {
		if strings.HasSuffix(key, ".csv") {
			delete(cm.constants, key)
		}
	}
	return nil
}

// read all csv characters files matched to given pattern.
func (csv *CsvManager) initCharacters(pattern string) error {
	files, err := filesystem.Glob(pattern)
	if err != nil {
		return err
	}
	if files == nil {
		return fmt.Errorf("no find character's load-pattern %v", pattern)
	}

	charas := make(map[int64]*Character, len(files))
	for _, file := range files {
		c, err := readCharacter(file, csv)
		if err != nil {
			return err
		}
		id := c.ID
		if _, ok := charas[id]; ok {
			return fmt.Errorf("CSV Character: duplicate definition: file: %v", file)
		}
		charas[id] = c
	}

	csv.CharaMap = charas
	return nil
}
