package state

import (
	"context"
	"github.com/mzki/erago/state/csv"
)

// game State holds all game parameters.
type GameState struct {
	CSV *csv.CsvManager

	ShareData  *UserVariables
	SystemData *SystemData

	*SaveInfo

	repo Repository
}

// consturct gameState with CSV Manager and config.
func NewGameState(csvdb *csv.CsvManager, repo Repository) *GameState {
	shareData := newUserVariablesShare(csvdb)
	state := &GameState{
		CSV:        csvdb,
		SystemData: newSystemData(csvdb),
		ShareData:  &shareData,
		SaveInfo:   newSaveInfo(),
		repo:       repo,
	}
	return state
}

// clear all data, including system and share, using 0 and empty string.
func (state *GameState) Clear() {
	state.SystemData.Clear()
	state.ShareData.Clear()
}

// save game system state to save[No.].
func (state *GameState) SaveSystem(no int) error {
	return state.repo.SaveSystemData(context.Background(), no, state)
}

// save game system state to save[No.] with comment.
func (state *GameState) SaveSystemWithComment(no int, comment string) error {
	state.SaveComment = comment
	return state.SaveSystem(no)
}

// load game system state from save[No.].
func (state *GameState) LoadSystem(no int) error {
	err := state.repo.LoadSystemData(context.Background(), no, state)
	if err != nil {
		return err
	}
	// require to recover unexported fields
	state.SystemData.refine(state.CSV)
	return nil
}

// save shared data to "share.sav"
func (state *GameState) SaveShare() error {
	return state.repo.SaveShareData(context.Background(), state)
}

// load shared data from "share.sav"
func (state *GameState) LoadShare() error {
	err := state.repo.LoadShareData(context.Background(), state)
	if err != nil {
		return err
	}
	// require to recover unexported fields
	state.ShareData.refine(state.CSV.Constants())
	return nil
}

// load only header from save[No.].
func (state *GameState) LoadHeader(no int) (*MetaData, error) {
	metaList, err := state.repo.LoadMetaList(context.Background(), no)
	if err != nil {
		return nil, err
	}
	return metaList[0], err
}

// check whether does save[No.] exists?
func (state *GameState) FileExists(no int) bool {
	return state.repo.Exist(context.Background(), no)
}

// to prevent user modify but export field
// use unexported type.

type intData struct {
	Values []int64
}
type strData struct {
	Values []string
}
type intParamMap map[string]intData
type strParamMap map[string]strData

type UserVariables struct {
	// exported to marshall/unmarshall object. user should not
	// access this field directory
	IntMap intParamMap
	StrMap strParamMap

	// unexported to not marshall/unmarshall object.
	constantMap map[string]csv.Constant
}

// NOTE: slice of given imap and smap are taken over UserVariables.
// be sure to not pass shared imap and smap.
func newUserVariablesByMap(imap map[string][]int64, smap map[string][]string, cmap map[string]csv.Constant) UserVariables {
	intMap := make(intParamMap, len(imap))
	for k, v := range imap {
		intMap[k] = intData{v}
	}

	strMap := make(strParamMap, len(smap))
	for k, v := range smap {
		strMap[k] = strData{v}
	}

	uv := UserVariables{
		IntMap:      intMap,
		StrMap:      strMap,
		constantMap: cmap,
	}
	return uv
}

func newUserVariablesSystem(cm *csv.CsvManager) UserVariables {
	return newUserVariablesByMap(
		cm.BuildIntUserVars(csv.ScopeSystem),
		cm.BuildStrUserVars(csv.ScopeSystem),
		cm.Constants(),
	)
}

func newUserVariablesShare(cm *csv.CsvManager) UserVariables {
	return newUserVariablesByMap(
		cm.BuildIntUserVars(csv.ScopeShare),
		cm.BuildStrUserVars(csv.ScopeShare),
		cm.Constants(),
	)
}

// clear contents of UserVariables.
// Int vars are cleared by 0, and Str vars are empty string.
func (uvars UserVariables) Clear() {
	for _, v := range uvars.IntMap {
		ZeroClear(v.Values)
	}
	for _, v := range uvars.StrMap {
		StrClear(v.Values)
	}
}

// Get IntParam queried by varname.
// return IntParam, found.
func (usr_vars UserVariables) GetInt(varname string) (IntParam, bool) {
	if vars, ok := usr_vars.IntMap[varname]; ok {
		indexer, _ := usr_vars.nameIndexer(varname)
		return NewIntParam(vars.Values, indexer), true
	}
	return IntParam{}, false
}

// Get []string variable queried by varname.
// return StrParam, found.
func (usr_vars UserVariables) GetStr(varname string) (StrParam, bool) {
	if vars, ok := usr_vars.StrMap[varname]; ok {
		indexer, _ := usr_vars.nameIndexer(varname)
		return NewStrParam(vars.Values, indexer), true
	}
	return StrParam{}, false
}

func (usr_vars *UserVariables) nameIndexer(varname string) (NameIndexer, bool) {
	if c, ok := usr_vars.constantMap[varname]; ok {
		return c.NameIndex, ok
	} else {
		return NoneNameIndexer{}, false
	}
}

// This methods is used for technical reason:
// UserVariables after unmarshaling has no constantMap since it is unexported,
// therefore, requiring re-set csv relationship.
func (usr_vars *UserVariables) refine(cmap map[string]csv.Constant) {
	usr_vars.constantMap = cmap
}

// iteration of each int parameters.
func (usr_vars UserVariables) ForEachIntParam(f func(string, IntParam)) {
	for key, vars := range usr_vars.IntMap {
		indexer, _ := usr_vars.nameIndexer(key)
		f(key, NewIntParam(vars.Values, indexer))
	}
}

// iteration of each str parameters.
func (usr_vars UserVariables) ForEachStrParam(f func(string, StrParam)) {
	for key, vars := range usr_vars.StrMap {
		indexer, _ := usr_vars.nameIndexer(key)
		f(key, NewStrParam(vars.Values, indexer))
	}
}

// System data has data using for the game system.
// it is remains after end game.
type SystemData struct {
	Chara *Characters

	// references of Chara
	Target *CharaReferences
	Master *CharaReferences
	Player *CharaReferences
	Assi   *CharaReferences

	UserVariables
}

func newSystemData(csvM *csv.CsvManager) *SystemData {
	charas := newCharacters(csvM)
	n_csv_chara := len(csvM.CharaMap)
	sysdata := &SystemData{
		Chara:         charas,
		Target:        newCharaReferences(n_csv_chara, charas),
		Master:        newCharaReferences(n_csv_chara, charas),
		Player:        newCharaReferences(n_csv_chara, charas),
		Assi:          newCharaReferences(n_csv_chara, charas),
		UserVariables: newUserVariablesSystem(csvM),
	}
	return sysdata
}

// clear all data using 0, empty string and nil.
func (sysdata *SystemData) Clear() {
	sysdata.Chara.Clear()

	sysdata.Target.Clear()
	sysdata.Master.Clear()
	sysdata.Player.Clear()
	sysdata.Assi.Clear()

	sysdata.UserVariables.Clear()
}

// refine csv relationship for internally, requiring after unmarshal.
func (sysdata *SystemData) refine(csvM *csv.CsvManager) {
	sysdata.Chara.refine(csvM)

	constants := csvM.Constants()
	sysdata.UserVariables.refine(constants)
}

// SaveInfo has information isolated save and load.
type SaveInfo struct {
	LastLoadVer     int32
	LastLoadComment string
	SaveComment     string
}

func newSaveInfo() *SaveInfo {
	return &SaveInfo{}
}
