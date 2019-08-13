package csv

import "testing"

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

var (
	ValidVSpecs = variableSpecs{
		"Base": {scopeSystem, dTypeInt, "Base", "./../../stub/CSV/Base.csv", []int{100}},
	}
	InValidVSpecs = variableSpecs{
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

func TestReadNames(t *testing.T) {
	_, err := readNames("not-exists.file", make([]int, 0), make([]string, 0))
	if err == nil {
		t.Fatal("Fatal Error: processing non exist file but non-error")
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
		BuiltinExName,
		// TODO BuiltinItemPriceName,; it's not string type, can't integrate Const map.
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

func TestBuildVariables(t *testing.T) {
	cm, err := newCsvManagerInited()
	if err != nil {
		t.Fatal(err)
	}

	intMap := cm.BuildIntUserVars(ScopeSystem)
	intNames := []string{
		BuiltinItemName,
		BuiltinItemStockName,
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

func BenchmarkNameIndexOf(b *testing.B) {
	cm, err := newCsvManagerInited()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := cm.NameIndexOf("Base", "体力")
		_ = idx
	}
}

func BenchmarkDirectNameIndex(b *testing.B) {
	cm, err := newCsvManagerInited()
	if err != nil {
		b.Fatal(err)
	}
	base_idxs := cm.constants["Base"].NameIndex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := base_idxs["体力"]
		_ = idx
	}
}
