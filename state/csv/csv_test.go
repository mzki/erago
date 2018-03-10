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
