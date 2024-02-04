package state

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mzki/erago/state/csv"
)

// initialize and finalize before and after testing.
func TestMain(m *testing.M) {
	if err := initialize(); err != nil {
		panic(err)
	}

	code := m.Run()
	// finalize()
	os.Exit(code)
}

var (
	// Global state of CSV.
	CSVDB *csv.CsvManager

	Repo Repository = &StubRepository{}
)

func initialize() error {
	CSVDB = &csv.CsvManager{}
	return CSVDB.Initialize(csv.Config{
		Dir:          "./../stub/CSV",
		CharaPattern: "Chara/Chara*",
	})
}

func TestNewGameState(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)

	base_vars, ok := gamestate.SystemData.GetInt("Number")
	if !ok {
		t.Error("Can not get System's Number variable.")
	}

	if _, ok := base_vars.GetByStr("数値１"); !ok {
		t.Error("Can not get Number[”数値１”] variable.")
	}
	if _, ok := base_vars.GetByStr("数値２"); ok {
		t.Error("Can get Number[”数値２”] variable. but it should not exist")
	}

	t.Log(base_vars)

	// compare length of
	// if base_vars.Len() != CSVDB.CharaMap[1].
}

func TestNewCharacter(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)
	const CHARA_ID = 1 // it must exist

	chara, err := gamestate.SystemData.Chara.AddID(CHARA_ID)
	if err != nil {
		t.Fatal(err)
	}
	base, ok := chara.GetInt("Base")
	if !ok {
		t.Fatal("the user variable Base is not found")
	}

	base.SetByStr("体力", 100)
	hp, ok := base.GetByStr("体力")
	if !ok {
		t.Fatal("Base［体力］is not found")
	}
	if hp != 100 {
		t.Error("different setted parametrer and got parametrer in Base[\"体力\"]")
	}

	csvc, ok := gamestate.CSV.CharaMap[CHARA_ID]
	if !ok {
		t.Fatal("no exists csv Character ID: 0.")
	}
	if csvbase := csvc.GetIntMap()["Base"]; csvbase[0] == hp {
		t.Error("setting parametrer to chara influeces csv parametrer: it breaks CSV constants")
	}
}

// testing for that system-data object can be marshalizable and unmarshalizable.
func TestMarshalSystemData(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)
	const CHARA_ID = 1 // it must exist
	const MASTER_IDX = 1

	chara, err := gamestate.SystemData.Chara.AddID(CHARA_ID)
	if err != nil {
		t.Fatal(err)
	}
	base, ok := chara.GetInt("Base")
	if !ok {
		t.Fatal("the user variable Base is not found")
	}

	base.SetByStr("体力", 100)
	t.Logf("before marshal %#v", base)

	// Set Master relation
	chara2, err := gamestate.SystemData.Chara.AddID(CHARA_ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := gamestate.SystemData.Master.Set(MASTER_IDX, chara2); err != nil {
		t.Fatal(err)
	}

	// marshal gamestate
	dump, err := json.Marshal(&gamestate.SystemData)
	if err != nil {
		t.Fatal(err)
	}

	// re-create gamestate using unmarshall
	newGamestate := NewGameState(CSVDB, Repo)
	if err := json.Unmarshal(dump, &newGamestate.SystemData); err != nil {
		t.Fatal(err)
	}
	// calling refine() is required after unmarshal.
	// It is done at (GameState).SaveSystem() and SaveShare() internally.
	newGamestate.SystemData.refine(newGamestate.CSV)

	// check unmarshal result
	loadedChara := newGamestate.SystemData.Chara.Get(0)
	if loadedChara == nil {
		t.Fatal("nil character after loaded")
	}
	loadedBase, ok := loadedChara.GetInt("Base")
	if !ok {
		t.Fatalf("cant access %s for loaded character", "Base")
	}
	if v, ok := loadedBase.GetByStr("体力"); !ok {
		t.Errorf("loaded %s has no key %s", "Base", "体力")
		t.Logf("after unmarshal %#v", loadedBase)
	} else if v != 100 {
		t.Errorf("loaded %s has differenct value, epxect %v, got %v", "Base", 100, v)
		t.Logf("after unmarshal %#v", loadedBase)
	}

	master := newGamestate.SystemData.Master.GetChara(MASTER_IDX)
	if master == nil {
		t.Fatalf("Nil Master after unmarshal at %v", MASTER_IDX)
	}
	chara2New := newGamestate.SystemData.Chara.Get(1)
	if master != chara2New {
		t.Errorf("Difference master and original chara after unmarshal: master = %p, org_chara = %p", master, chara2New)
	}
}

// testing for system-data object should not remain older data
func TestUnmarshalSystemDataNotRemainOlderData(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)
	const CharaID = 1 // it must exist

	chara, err := gamestate.SystemData.Chara.AddID(CharaID)
	if err != nil {
		t.Fatal(err)
	}
	base, ok := chara.GetInt("Base")
	if !ok {
		t.Fatal("the user variable Base is not found")
	}

	base.SetByStr("体力", 100)
	t.Logf("before marshal %#v", base)

	// marshal gamestate
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}

	// modify game state after marshall
	const BaseAfterValue = 201
	base.SetByStr("体力", BaseAfterValue)
	// 2nd chara
	if _, err := gamestate.SystemData.Chara.AddID(CharaID); err != nil {
		t.Fatal(err)
	}
	if l := gamestate.SystemData.Chara.Len(); l != 2 {
		t.Fatalf("expect 2 characters but %v", l)
	}

	//  gamestate using unmarshall
	if err := gamestate.LoadSystem(0); err != nil {
		t.Fatal(err)
	}

	// check unmarshal result
	if l := gamestate.SystemData.Chara.Len(); l != 1 {
		t.Errorf("expect old gamestate to have 1 charcter , but len(Chara)=%d", l)
	}

	unmarshalChara := gamestate.SystemData.Chara.Get(0)
	if unmarshalChara == nil {
		t.Fatalf("nil characeter at index 0 after unmarshall")
	}

	// Check whether values are equal before and after marshal/unmarshal
	oldBaseV, oldBaseOk := base.GetByStr("体力")
	newBase, _ := unmarshalChara.GetInt("Base")
	newBaseV, newBaseOk := newBase.GetByStr("体力")
	if !oldBaseOk || !newBaseOk {
		t.Fatal("missing Base value")
	}
	if oldBaseV == newBaseV {
		t.Errorf("propagate unmarshal value for old data, not expect: %v, got: %v", oldBaseV, newBaseV)
	}
}

// testing for system-data object should not remain older data
func TestUnmarshalUserDataShouldNotHasOldData(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)

	// marshal gamestate
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}
	if err := gamestate.SaveShare(); err != nil {
		t.Fatal(err)
	}

	// Add unknown key
	const UnknownValueKey = "Unknown"
	gamestate.SystemData.UserVariables.IntMap[UnknownValueKey] = intData{}
	gamestate.ShareData.IntMap[UnknownValueKey] = intData{}

	//  unmarshall gamestate
	if err := gamestate.LoadSystem(0); err != nil {
		t.Fatal(err)
	}
	if err := gamestate.LoadShare(); err != nil {
		t.Fatal(err)
	}

	// check unmarshal result
	if _, ok := gamestate.SystemData.UserVariables.GetInt(UnknownValueKey); ok {
		t.Errorf("Loaded game state should not have unknown value key, but has key: %v", UnknownValueKey)

		names := make([]string, 0, 4)
		for k := range gamestate.SystemData.UserVariables.IntMap {
			names = append(names, k)
		}
		t.Logf("Dump IntMap keys: %#v", names)
	}
	// check unmarshal result
	if _, ok := gamestate.ShareData.GetInt(UnknownValueKey); ok {
		t.Errorf("Loaded game state should not have unknown value key, but has key: %v", UnknownValueKey)

		names := make([]string, 0, 4)
		for k := range gamestate.ShareData.IntMap {
			names = append(names, k)
		}
		t.Logf("Dump IntMap keys: %#v", names)
	}
}

// testing for system-data object should have value existing in CSV database
// This occurs when update CSV database to add new value then older persistet data not have
// such values.
func TestUnmarshalUserDataShouldHaveCsvData(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)
	const CharaID = 1 // must exist

	// Drop existent key
	const DropValueKey = "CStr"
	const ShrinkValueKey = "Base"
	if _, ok := gamestate.CSV.Constants()[DropValueKey]; !ok {
		t.Fatalf("This test requires CSV to must have key %v", DropValueKey)
	}
	chara, err := gamestate.SystemData.Chara.AddID(CharaID)
	if err != nil {
		t.Fatal(err)
	}
	// Drop specific field to suppose older data
	delete(chara.UserVariables.StrMap, DropValueKey)
	// Shrink specific field to suppose older data
	shrinkData := chara.IntMap[ShrinkValueKey].Values
	shrinkDataOriginalLen := len(shrinkData)
	chara.IntMap.addEntry(ShrinkValueKey, shrinkData[:shrinkDataOriginalLen-1])

	// marshal gamestate
	if err := gamestate.SaveSystem(0); err != nil {
		t.Fatal(err)
	}

	// Remove character
	gamestate.SystemData.Chara.Remove(0)

	//  unmarshall gamestate; scenario: many time later... csv has dropped key.
	if err := gamestate.LoadSystem(0); err != nil {
		t.Fatal(err)
	}

	// check unmarshal result
	chara = gamestate.SystemData.Chara.Get(0)
	if chara == nil {
		t.Fatalf("after loaded, character(index %v) is not restored", 0)
	}
	if _, ok := chara.UserVariables.GetStr(DropValueKey); !ok {
		t.Errorf("Loaded game chara should have csv defined value, but not exist key: %v", DropValueKey)
	}
	if v, ok := chara.UserVariables.GetInt(ShrinkValueKey); !ok {
		t.Errorf("Loaded game chara should have csv defined value, but not eixst key: %v", ShrinkValueKey)
	} else if len(v.Values) != shrinkDataOriginalLen {
		t.Errorf("Loaded ShrinkData (%v) has diffrent length: %v, from csv defined: %v", ShrinkValueKey, len(v.Values), shrinkDataOriginalLen)
	}
}

const (
	IntKey  = "i"
	StrKey  = "s"
	DataKey = "1"
	VarLen  = 10
)

var (
	testDataIntVSpecs = intVariableSpecs{
		csv.VariableSpec{VarName: IntKey, Scope: csv.ScopeSystem, Size: VarLen},
	}
	testDataStrVSpecs = strVariableSpecs{
		csv.VariableSpec{VarName: StrKey, Scope: csv.ScopeSystem, Size: VarLen},
	}
)

func newTestUserVariable() UserVariables {

	imap := map[string][]int64{
		IntKey: make([]int64, VarLen),
	}
	smap := map[string][]string{
		StrKey: make([]string, VarLen),
	}
	constant := csv.Constant{Names: []string{DataKey}, NameIndex: map[string]int{DataKey: 1}}
	cmap := map[string]csv.Constant{
		IntKey: constant,
		StrKey: constant,
	}

	return newUserVariablesByMap(imap, smap, cmap)
}

func TestUserVariable(t *testing.T) {
	uv := newTestUserVariable()

	{
		intVar, ok := uv.GetInt(IntKey)
		if !ok {
			t.Fatal("failed to get expected int params")
		}
		if got := len(intVar.Values); got != VarLen {
			t.Fatalf("different int var len, expected %v, got %v", got, VarLen)
		}
		if _, ok := intVar.GetByStr(DataKey); !ok {
			t.Errorf("cant get expected value with key %s", DataKey)
		}

		// not found case
		if _, ok := uv.GetInt(StrKey); ok {
			t.Errorf("got int param with invalid key %s", StrKey)
		}
	}

	{
		strVar, ok := uv.GetStr(StrKey)
		if !ok {
			t.Fatal("failed to get expected str params")
		}
		if got := len(strVar.Values); got != VarLen {
			t.Fatalf("different int var len, expected %v, got %v", got, VarLen)
		}
		if _, ok := strVar.GetByStr(DataKey); !ok {
			t.Errorf("cant get expected value with key %s", DataKey)
		}

		// not found case
		if _, ok := uv.GetStr(IntKey); ok {
			t.Errorf("got str param with invalid key %s", IntKey)
		}
	}
}

func TestUserVariableMarshal(t *testing.T) {
	uv := newTestUserVariable()
	vars, _ := uv.GetInt(IntKey)
	vars.SetByStr(DataKey, 100)

	dump, err := json.Marshal(&uv)
	if err != nil {
		t.Fatal(err)
	}

	newUV := UserVariables{}
	if err := json.Unmarshal(dump, &newUV); err != nil {
		t.Fatal(err)
	}
	// reuiring call refine() after unmarshal.
	newUV.refine(uv.constantMap, testDataIntVSpecs, testDataStrVSpecs)

	vars, _ = newUV.GetInt(IntKey)
	if v, ok := vars.GetByStr(DataKey); !ok {
		t.Errorf("unmarshaled value is not found with key %s", DataKey)
		t.Logf("%#v", newUV)
	} else if v != 100 {
		t.Errorf("unmarshaled value is different, expect %v, got %v", 100, v)
		t.Logf("%#v", newUV)
	}
}

func TestUserVariableForEach(t *testing.T) {
	gamestate := NewGameState(CSVDB, Repo)

	// Int
	{
		const expectLen = 4
		keys := make([]string, 0, expectLen)
		intparams := make([]IntParam, 0, expectLen)
		gamestate.SystemData.ForEachIntParam(func(k string, v IntParam) {
			keys = append(keys, k)
			intparams = append(intparams, v)
		})

		if len(keys) != expectLen {
			t.Errorf("different int param size, expect %v, got %v, keys %#v", expectLen, len(keys), keys)
		}
	}

	// String
	{
		const expectLen = 1
		keys := make([]string, 0, expectLen)
		strparams := make([]StrParam, 0, expectLen)
		gamestate.SystemData.ForEachStrParam(func(k string, v StrParam) {
			keys = append(keys, k)
			strparams = append(strparams, v)
		})

		if len(keys) != expectLen {
			t.Errorf("different str param size, expect %v, got %v, keys %#v", expectLen, len(keys), keys)
		}
	}
}
