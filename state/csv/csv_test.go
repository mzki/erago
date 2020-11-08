package csv

import (
	"fmt"
	"strings"
	"testing"
)

var validConfig = Config{
	Dir:          "../../stub/CSV/",
	CharaPattern: "Chara/Chara*",
}

func TestCsvInit(t *testing.T) {
	config := validConfig
	csv := &CsvManager{}
	if err := csv.Initialize(config); err != nil {
		t.Fatal(err)
	}
	if i := csv.NameIndexOf("Base", "体力"); i == -1 {
		t.Error("Base 体力 is not parsed. See Base.csv to check that existance.")
	} else if i != 0 {
		t.Error("Base 体力 0 is not parsed. See Base.csv to check that existance.")
	}

	if _, ok := csv.CharaMap[1]; !ok {
		t.Fatal("chara of ID 1 is not parsed")
	}
	if c := csv.CharaMap[1]; c.Name != "霊夢" {
		t.Errorf("ID 1 chara name is got: %v, expect %v", c.Name, "霊夢")
	}
}

var invalidConfig = Config{
	Dir:          "path/to/not/found/CSV",
	CharaPattern: "Chara/Chara*",
}

func TestCsvInitInvalidConfig(t *testing.T) {
	config := invalidConfig
	csv := &CsvManager{}
	if err := csv.Initialize(config); err == nil {
		t.Fatalf("invalid config %v given, but no error!", config)
	}
}

var (
	ValidVSpecs = variableSpecInternalMap{
		"Base": {scopeSystem, dTypeInt, "Base", "./../../stub/CSV/Base.csv", []int{100}},
	}
	InValidVSpecs = variableSpecInternalMap{
		"Base": {scopeSystem, dTypeInt, "Base", "Unkown/Dir/Base.csv", []int{100}},
	}
)

func TestCsvInitVariableSpecs(t *testing.T) {
	csv := &CsvManager{}
	if err := csv.initVariableSpecs(InValidVSpecs); err == nil {
		t.Fatal("must be error but nil")
	}
}

func TestCsvBuildConstants(t *testing.T) {
	csv := &CsvManager{}

	err := csv.buildConstants(ValidVSpecs)
	if err != nil {
		t.Fatal(err)
	}

	if has := csv.constants["Base"].NameIndex.Has("体力"); !has {
		t.Error("Base 体力 is not parsed. See Base.csv to check that existance.")
	}

	csv = &CsvManager{}
	if err := csv.buildConstants(InValidVSpecs); err == nil {
		t.Fatal("must be error. but nil")
	}
}

func newCsvManagerInited() (*CsvManager, error) {
	config := validConfig
	csv := &CsvManager{}
	err := csv.Initialize(config)
	return csv, err
}

func TestBuiltinVariables(t *testing.T) {
	cm, err := newCsvManagerInited()
	if err != nil {
		t.Fatal(err)
	}

	constMap := cm.Constants()
	constNames := []string{
		BuiltinTrainName,
		BuiltinSourceName,
	}
	for _, name := range constNames {
		if _, ok := constMap[name]; !ok {
			t.Errorf("Missing Builtin Constant: %v", name)
		}
	}
	if cm.ItemPrices == nil {
		t.Errorf("Missing Builtin Constant: %v", BuiltinItemPriceName)
	}

	intMap := cm.BuildIntUserVars(ScopeSystem)
	intNames := []string{
		BuiltinItemName,
		BuiltinItemStockName,
		BuiltinMoneyName,
	}
	for _, name := range intNames {
		if _, ok := intMap[name]; !ok {
			t.Errorf("Missing Builtin system value: %v", name)
		}
	}

	charaMap := cm.BuildIntUserVars(ScopeChara)
	charaNames := []string{
		BuiltinParamName,
		BuiltinJuelName,
		BuiltinAblName,
		BuiltinTalentName,
		BuiltinMarkName,
		BuiltinExpName,
	}
	for _, name := range charaNames {
		if _, ok := charaMap[name]; !ok {
			t.Errorf("Missing Builtin chara value: %v", name)
		}
	}
}

func TestBuiltinCustomFieldsExist(t *testing.T) {
	csv, err := newCsvManagerInited()
	if err != nil {
		t.Fatal(err)
	}

	Item := csv.constants["Item"]
	if !Item.CustomFields.Has(HeaderFieldItemPrice) {
		t.Error("Item.price is not parsed. See Item.csv to check that existance.")
	}
	if got := Item.CustomFields.TypeOf(HeaderFieldItemPrice); got != CFIntType {
		t.Errorf("Item.price type is invalid. expect int(%v), got %v", CFIntType, got)
	}
}

func TestBuildVariables(t *testing.T) {
	cm, err := newCsvManagerInited()
	if err != nil {
		t.Fatal(err)
	}

	intMap := cm.BuildIntUserVars(ScopeSystem)
	intNames := []string{
		BuiltinItemName,
		BuiltinItemStockName,
		BuiltinMoneyName,
		"Number",
	}
	if len(intMap) != len(intNames) {
		t.Errorf("different int usr variable size, expect %v, got %v", len(intNames), len(intMap))
	}
	for _, name := range intNames {
		if _, ok := intMap[name]; !ok {
			t.Errorf("Missing int system value: %v", name)
		}
	}

	strMap := cm.BuildStrUserVars(ScopeSystem)
	strNames := []string{
		"Str",
	}
	if len(strMap) != len(strNames) {
		t.Errorf("different str usr variable size, expect %v, got %v", len(strNames), len(strMap))
	}
	for _, name := range strNames {
		if _, ok := strMap[name]; !ok {
			t.Errorf("Missing str system value: %v", name)
		}
	}
}

func TestDuplicateBuildinVariables(t *testing.T) {
	VSPEC := fmt.Sprintf(`
CSV,Int,%s, ,100
System,Int,%s, ,100
Share,Int,%s, ,100
Chara,Int,%s, ,100
`, BuiltinTrainName, BuiltinSourceName, BuiltinParamName, BuiltinJuelName)

	reader := strings.NewReader(VSPEC)
	all_specs, err := readVariableSpecs(reader)
	if err != nil {
		t.Fatal(err)
	}

	notAppended := appendBuiltinVSpecs(all_specs)
	if len(notAppended) == 0 {
		t.Error("duplicate builtin but not detected")
	}
}
