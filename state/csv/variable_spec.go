package csv

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mzki/erago/filesystem"
	"github.com/mzki/erago/util/errutil"
)

// variableSpecInternal defines the spec of user defined varables.
type variableSpecInternal struct {
	Scope    vspecIdent // where the variable is used.
	DataType vspecIdent // type of the variable.
	VarName  string     // name for the variable.
	FileName string     // file to read each name of the variables.
	Size     []int      // DEPLECATED: auto-detected by csv file. the variable sizes
}

// identifer for variableSpec
type vspecIdent uint8

const (
	scopeSystem vspecIdent = iota
	scopeShare
	scopeChara
	scopeCSV
)

const (
	dTypeInt vspecIdent = scopeCSV + 1 + iota
	dTypeStr
)

var parseScopeMap = map[string]vspecIdent{
	"System": scopeSystem,
	"Share":  scopeShare,
	"Chara":  scopeChara,
	"CSV":    scopeCSV,
}

var parseDTypeMap = map[string]vspecIdent{
	"Int": dTypeInt,
	"Str": dTypeStr,
}

var builtinVSpecs = variableSpecInternalMap{
	// CSV scope, dType is not used and not allocated.
	BuiltinTrainName:  {scopeCSV, dTypeStr, BuiltinTrainName, BuiltinTrainName + ".csv", []int{0}},
	BuiltinSourceName: {scopeCSV, dTypeInt, BuiltinSourceName, BuiltinSourceName + ".csv", []int{0}},

	// Chara scope, to be used as character variable.
	BuiltinParamName: {scopeChara, dTypeInt, BuiltinParamName, BuiltinParamName + ".csv", []int{0}},
	// Juel shares with Param.csv
	BuiltinJuelName:   {scopeChara, dTypeInt, BuiltinJuelName, BuiltinParamName + ".csv", []int{0}},
	BuiltinAblName:    {scopeChara, dTypeInt, BuiltinAblName, BuiltinAblName + ".csv", []int{0}},
	BuiltinTalentName: {scopeChara, dTypeInt, BuiltinTalentName, BuiltinTalentName + ".csv", []int{0}},
	BuiltinMarkName:   {scopeChara, dTypeInt, BuiltinMarkName, BuiltinMarkName + ".csv", []int{0}},
	BuiltinExpName:    {scopeChara, dTypeInt, BuiltinExpName, BuiltinExpName + ".csv", []int{0}},

	// System scope, to be used as global variable.
	BuiltinItemName:      {scopeSystem, dTypeInt, BuiltinItemName, BuiltinItemName + ".csv", []int{0}},
	BuiltinItemStockName: {scopeSystem, dTypeInt, BuiltinItemStockName, BuiltinItemName + ".csv", []int{0}},

	// System scope, with no csv
	BuiltinMoneyName: {scopeSystem, dTypeInt, BuiltinMoneyName, "", []int{1}},
}

// read new specs of user variables from file, "VariableSpec.csv".
func readVariableSpecsFile(fname string) (variableSpecInternalMap, error) {
	fp, err := filesystem.Load(fname)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return readVariableSpecs(fp)
}

// read new specs of user variables from io.Reader
func readVariableSpecs(r io.Reader) (variableSpecInternalMap, error) {
	vspecs := make(variableSpecInternalMap, len(builtinVSpecs)+16) // 16 is arbitrary value

	err := ReadFunc(r, func(record []string) error {
		vs, err := parseVariableSpec(record)
		if err != nil {
			return err
		}

		varname := vs.VarName
		if _, has := vspecs[varname]; has {
			return fmt.Errorf("duplicate VarName(%s)", varname)
		}
		vspecs[varname] = vs
		return nil
	})
	return vspecs, err
}

// append builtin variablespecs into given vspec map.
// Some buitin vspec are not appended if the variable name for these
// already exist in the given vspec map.
// It returns appended number of builtin variable spec and the maximum.
func appendBuiltinVSpecs(vspecs variableSpecInternalMap) []string {
	notApeendedKeys := make([]string, 0, 4)
	for vname, v := range builtinVSpecs {
		if _, has := vspecs[vname]; has {
			notApeendedKeys = append(notApeendedKeys, vname)
		} else {
			vspecs[vname] = v
		}
	}
	return notApeendedKeys
}

func parseVariableSpec(record []string) (variableSpecInternal, error) {
	vspec := variableSpecInternal{}
	if len(record) < 5 {
		return vspec, errors.New(`Variables must be defined by at least 5 columns:
		Scope, DataType, VarName, FileName, Size, (Size2, ...)].`)
	}

	// parse each record
	merr := errutil.NewMultiError()
	scope, err := parseIdent(record[0], parseScopeMap)
	merr.Add(err)
	dtype, err := parseIdent(record[1], parseDTypeMap)
	merr.Add(err)

	if len(record[2]) == 0 {
		if len(record[3]) == 0 {
			return vspec, errors.New("VarName or FileName must not be empty.")
		}
		record[2] = basenameWithoutExt(record[3])
	}

	// TODO: dimention := len(record[4:])
	// cover 2D or 3D array
	var var_size int
	if record[4] == "" {
		var_size = 0
	} else {
		var_size, err = strconv.Atoi(record[4])
		merr.Add(err)
	}

	if err := merr.Err(); err != nil {
		return vspec, err
	}

	return variableSpecInternal{
		Scope:    scope,
		DataType: dtype,
		VarName:  record[2],
		FileName: record[3],
		Size:     []int{var_size},
	}, nil
}

func parseIdent(field string, m map[string]vspecIdent) (vspecIdent, error) {
	if ident, ok := m[field]; ok {
		return ident, nil
	}
	return vspecIdent(0), fmt.Errorf("unkown vspec field (%v). must be in %v", field, identKey(m))
}

func identKey(m map[string]vspecIdent) []string {
	ks := make([]string, 0, len(m))
	for k, _ := range m {
		ks = append(ks, k)
	}
	return ks
}

// func fieldInRange(query string, candidates ...string) error {
// 	for _, c := range candidates {
// 		if query == c {
// 			return nil // match
// 		}
// 	}
// 	return fmt.Errorf("The field must be %v. now: %v", candidates, query)
// }

func basenameWithoutExt(fname string) string {
	base := filepath.Base(fname)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// define type that has a method selelct_by()
type variableSpecInternalMap map[string]variableSpecInternal

func (vspecs variableSpecInternalMap) selectBy(f func(variableSpecInternal) bool) variableSpecInternalMap {
	new_vspecs := make(variableSpecInternalMap, len(vspecs)/2)
	for k, v := range vspecs {
		if f(v) {
			new_vspecs[k] = v
		}
	}
	return new_vspecs
}

func (vspecs variableSpecInternalMap) selectByScopeAndDType(scope, dtype vspecIdent) variableSpecInternalMap {
	return vspecs.selectBy(func(vs variableSpecInternal) bool {
		return vs.Scope == scope && vs.DataType == dtype
	})
}

func (vspecs variableSpecInternalMap) find(vname string) (vspec variableSpecInternal, ok bool) {
	v, ok := vspecs[vname]
	return v, ok
}

func (vspecs variableSpecInternalMap) Map(f func(variableSpecInternal) variableSpecInternal) variableSpecInternalMap {
	new_vs := make(variableSpecInternalMap, len(vspecs))
	for k, v := range vspecs {
		new_vs[k] = f(v)
	}
	return new_vs
}
