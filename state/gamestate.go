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
	return state.repo.LoadSystemData(context.Background(), no, state)
}

// save shared data to "share.sav"
func (state *GameState) SaveShare() error {
	return state.repo.SaveShareData(context.Background(), state)
}

// load shared data from "share.sav"
func (state *GameState) LoadShare() error {
	return state.repo.LoadShareData(context.Background(), state)
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
type intParamMap map[string]IntParam
type strParamMap map[string]StrParam

type UserVariables struct {
	IntMap intParamMap
	StrMap strParamMap
}

func newUserVariables() UserVariables {
	return UserVariables{
		IntMap: make(intParamMap),
		StrMap: make(strParamMap),
	}
}

// NOTE: slice of given imap and smap are taken over UserVariables.
// be sure to not pass shared imap and smap.
func newUserVariablesByMap(imap map[string][]int64, smap map[string][]string, cmap map[string]csv.Constant) UserVariables {
	uv := newUserVariables()
	for name, ivars := range imap {
		var nidx NameIndexer
		if c, ok := cmap[name]; ok {
			nidx = c.NameIndex
		} else {
			nidx = NoneNameIndexer{}
		}
		uv.IntMap[name] = NewIntParam(ivars, nidx)
	}

	for name, svars := range smap {
		var nidx NameIndexer
		if c, ok := cmap[name]; ok {
			nidx = c.NameIndex
		} else {
			nidx = NoneNameIndexer{}
		}
		uv.StrMap[name] = NewStrParam(svars, nidx)
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
		return vars, true
	}
	return IntParam{}, false
}

// Get []string variable queried by varname.
// return StrParam, found.
func (usr_vars UserVariables) GetStr(varname string) (StrParam, bool) {
	if vars, ok := usr_vars.StrMap[varname]; ok {
		return vars, true
	}
	return StrParam{}, false
}

// iteration of each int parameters.
func (usr_vars UserVariables) ForEachIntParam(f func(string, IntParam)) {
	for key, param := range usr_vars.IntMap {
		f(key, param)
	}
}

// iteration of each str parameters.
func (usr_vars UserVariables) ForEachStrParam(f func(string, StrParam)) {
	for key, param := range usr_vars.StrMap {
		f(key, param)
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

// SaveInfo has information isolated save and load.
type SaveInfo struct {
	LastLoadVer     int32
	LastLoadComment string
	SaveComment     string
}

func newSaveInfo() *SaveInfo {
	return &SaveInfo{}
}
