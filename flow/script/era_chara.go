package script

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	"local/erago/state"
)

const (
	luaCharaListName       = "chara"
	luaCharaRefsMasterName = "master"
	luaCharaRefsTargetName = "target"
	luaCharaRefsPlayerName = "player"
	luaCharaRefsAssiName   = "assi"

	luaCharaListMetaName = "charalist"
	luaCharaRefsMetaName = "chararefs"
	luaCharacterMetaName = "character"
)

// register Chara data, method, metatables.
func registerCharaParams(L *lua.LState, gamestate *state.GameState) {
	era_module := mustGetEraModule(L)
	if lv := era_module.RawGetString(luaCharaListName); lua.LVAsBool(lv) {
		return // already exists
	}

	registerCharaMeta(L) // must be first

	{ // register chara list
		chara_list_meta := newMetatable(L, luaCharaListMetaName, map[string]lua.LValue{
			"__index":     L.NewFunction(getCharaListFields),
			"__len":       L.NewFunction(lenScalable),
			"__metatable": metaProtectObj,
		})
		chara_list := newLuaCharaList(L, gamestate.SystemData.Chara)
		ud := newUserDataWithMt(L, chara_list, chara_list_meta)

		era_module.RawSetString(luaCharaListName, ud)
	}

	{ // register chara refereces
		get_set_func := L.NewFunction(getSetCharaReferences)
		chara_refs_meta := newMetatable(L, luaCharaRefsMetaName, map[string]lua.LValue{
			"__index":     get_set_func,
			"__newindex":  get_set_func,
			"__len":       L.NewFunction(lenScalable),
			"__metatable": metaProtectObj,
		})

		for key, lv := range map[string]lua.LValue{
			luaCharaRefsAssiName:   newUserDataWithMt(L, gamestate.SystemData.Assi, chara_refs_meta),
			luaCharaRefsMasterName: newUserDataWithMt(L, gamestate.SystemData.Master, chara_refs_meta),
			luaCharaRefsTargetName: newUserDataWithMt(L, gamestate.SystemData.Target, chara_refs_meta),
			luaCharaRefsPlayerName: newUserDataWithMt(L, gamestate.SystemData.Player, chara_refs_meta),
		} {
			era_module.RawSetString(key, lv)
		}
	}
}

// //  cahracter list or state.Characters

type luaCharaList struct {
	*state.Characters
	methods map[string]*lua.LFunction
	// chara meta
}

func newLuaCharaList(L *lua.LState, cs *state.Characters) luaCharaList {
	return luaCharaList{
		Characters: cs,
		methods: map[string]*lua.LFunction{
			"len":      L.NewFunction(lenScalable),
			"add":      L.NewFunction(charaListAdd),
			"addEmpty": L.NewFunction(charaListAddEmpty),
			"remove":   L.NewFunction(charaListRemove),
			"clear":    L.NewFunction(charaListClear),
			// TODO: implement charas_methods_table["range"] = L.NewFunction(
			// TODO: implement charas_methods_table["find"] = L.NewFunction(
			// TODO: implement charas_methods_table["sort"] = L.NewFunction(
		},
	}
}

func checkLuaCharaList(L *lua.LState, pos int) luaCharaList {
	ud := L.CheckUserData(pos)
	if charas, ok := ud.Value.(luaCharaList); ok {
		return charas
	}
	L.ArgError(pos, "require Characters")
	return luaCharaList{}
}

// +gendoc "Characters"
// * added_charas = chara:add(id1, id2, ...)
func charaListAdd(L *lua.LState) int {
	nargs := L.GetTop()
	if nargs < 2 {
		L.ArgError(2, "require some Character ID")
	}

	charas := checkLuaCharaList(L, 1)
	chara_meta := L.GetTypeMetatable(luaCharacterMetaName)

	if n_ids := nargs - 1; n_ids == 1 {
		c, err := charas.AddID(L.CheckInt64(2))
		if err != nil {
			L.ArgError(2, err.Error())
		}
		L.Push(newUserDataWithMt(L, c, chara_meta))
		return 1
	} else {
		// case: n_ids > 1
		added_charas := L.NewTable()
		for i := 2; i <= nargs; i++ {
			id := L.CheckInt64(i)
			c, err := charas.AddID(id)
			if err != nil {
				L.ArgError(i, err.Error())
			}

			ud := newUserDataWithMt(L, c, chara_meta)
			added_charas.Insert(i-1, ud)
		}

		L.Push(added_charas)
		return 1
	}
}

// +gendoc "Characters"
// * empty_chara = chara:addEmpty()
func charaListAddEmpty(L *lua.LState) int {
	charas := checkLuaCharaList(L, 1)
	empty_chara := charas.AddEmptyCharacter()
	ud := newUserDataWithMt(L, empty_chara, L.GetTypeMetatable(luaCharacterMetaName))
	L.Push(ud)
	return 1
}

// +gendoc "Characters"
// * chara:remove(index)
func charaListRemove(L *lua.LState) int {
	charas := checkLuaCharaList(L, 1)
	idx := L.CheckInt(2)
	if ok := charas.Remove(idx); !ok {
		L.ArgError(2, "given index is empty in character list")
	}
	return 0
}

// +gendoc "Characters"
// * chara:clear()
func charaListClear(L *lua.LState) int {
	charas := checkLuaCharaList(L, 1)
	charas.Clear()
	return 0
}

// get character in the list or method for chara list.
func getCharaListFields(L *lua.LState) int {
	charas := checkLuaCharaList(L, 1)
	switch key := L.CheckAny(2).(type) {

	case lua.LString:
		// extract methods
		fn, ok := charas.methods[key.String()]
		if !ok {
			L.ArgError(2, fmt.Sprintf("method %s is not found", key))
		}
		L.Push(fn)
		return 1

	case lua.LNumber:
		// extract character
		chara := charas.Get(int(key))
		if chara == nil {
			L.ArgError(2, indexOutMessage)
		}
		ud := newUserDataWithMt(L, chara, L.GetTypeMetatable(luaCharacterMetaName))
		L.Push(ud)
		return 1

	default:
		// unknown key
		L.ArgError(2, fmt.Sprintf("invalid key %s", key))
	}
	return 0
}

// // character references: Target, Assi, Master, etc

func checkCharaRefereces(L *lua.LState, pos int) *state.CharaReferences {
	ud := L.CheckUserData(pos)
	if data, ok := ud.Value.(*state.CharaReferences); ok {
		return data
	}
	L.ArgError(pos, "require character references: target, master, assi ...")
	return nil
}

func getSetCharaReferences(L *lua.LState) int {
	refs := checkCharaRefereces(L, 1)
	index := L.CheckInt(2)

	if L.GetTop() == 3 {
		c := checkCharacter(L, 3)
		if err := refs.Set(index, c); err != nil {
			L.ArgError(3, "can not set references of character. "+err.Error())
		}
		return 0
	}

	c := refs.GetChara(index)
	if c == nil {
		L.ArgError(2, indexOutMessage)
	}
	ud := newUserDataWithMt(L, c, L.GetTypeMetatable(luaCharacterMetaName))
	L.Push(ud)
	return 1
}

// // lua character

func registerCharaMeta(L *lua.LState) {
	if lv := L.GetTypeMetatable(luaCharacterMetaName); lua.LVAsBool(lv) {
		return // already exists
	}

	_ = newMetatable(L, luaCharacterMetaName, map[string]lua.LValue{
		"__index":     L.NewFunction(getCharaFields),
		"__newindex":  L.NewFunction(setCharaFields),
		"__metatable": metaProtectObj,
	})
}

// check position of L is character?
func checkCharacter(L *lua.LState, pos int) *state.Character {
	ud := L.CheckUserData(pos)
	if chara, ok := ud.Value.(*state.Character); ok {
		return chara
	}
	L.ArgError(pos, "require character")
	return nil
}

// +gendoc "Lua Character"
// * chara.id
//
// キャラクターのIDを示す数値。CSVファイル上では番号と表され、キャラクターの種類を示す。
// readonlyな変数である。
//
// * chara.uid
//
// キャラクターのUID。キャラクター自身を区別する一意の数値。
// 例えば、IDが同じ(同じキャラクターの種類)キャラクターが複数いる場合、
// UIDによって個々を区別することができる。
// readonlyな変数である。
//
// * chara.is_assi
//
// 数値型の変数。
// キャラクターが調教の助手が可能であるかを示すことを目的としている。
// readonlyな変数である。
//
// * chara.name
//
// 文字列型の変数。
// キャラクターの正式名を保持する。読み書き可能な変数である。
//
// * chara.master_name
//
// TODO deprecated. move to CStr?
// 文字列型の変数。
// 主人の呼び名を保持する。読み書き可能な変数である。
//
// * chara.nick_name
//
// TODO deprecated. move to CStr?
// 文字列型の変数。
// キャラクターのあだ名を保持する。読み書き可能な変数である。
//
// * chara.call_name
//
// TODO deprecated. move to CStr?
// 文字列型の変数。
// 呼び名を保持する。読み書き可能な変数である。
//

const (
	// read only
	characterFieldIDName     = "id"
	characterFieldUIDName    = "uid"
	characterFieldIsAssiName = "is_assi"

	// read and write
	characterFieldNameName       = "name"
	characterFieldMasterNameName = "master_name"
	characterFieldNickNameName   = "nick_name"
	characterFieldCallNameName   = "call_name"
)

// case __index
// TODO: replace map[string]func
func getCharaFields(L *lua.LState) int {
	c := checkCharacter(L, 1)
	key := L.CheckString(2)

	switch key {
	// int fields
	case characterFieldIDName:
		L.Push(lua.LNumber(c.ID))
		return 1

	case characterFieldUIDName:
		L.Push(lua.LNumber(c.UID))
		return 1

	case characterFieldIsAssiName:
		L.Push(lua.LNumber(c.IsAssi))
		return 1

	// string fields
	case characterFieldNameName:
		L.Push(lua.LString(c.Name))
		return 1

	case characterFieldNickNameName:
		L.Push(lua.LString(c.NickName))
		return 1

	case characterFieldMasterNameName:
		L.Push(lua.LString(c.MasterName))
		return 1

	case characterFieldCallNameName:
		L.Push(lua.LString(c.CallName))
		return 1

	// user defined values
	default:
		if iparam, ok := c.GetInt(key); ok {
			L.Push(newLIntParam(L, iparam))
			return 1
		} else if sparam, ok := c.GetStr(key); ok {
			L.Push(newLStrParam(L, sparam))
			return 1
		}
		L.ArgError(2, "unknown character field: "+key)
	}
	return 0
}

// case __newindex
// TODO: replace map[string]func
func setCharaFields(L *lua.LState) int {
	c := checkCharacter(L, 1)
	key := L.CheckString(2)
	val := L.CheckString(3) // assignment is allowed only string type.
	switch key {
	case characterFieldNameName:
		c.Name = val

	case characterFieldCallNameName:
		c.CallName = val

	case characterFieldMasterNameName:
		c.MasterName = val

	case characterFieldNickNameName:
		c.NickName = val

	default:
		L.ArgError(2, "character does not have "+key)
	}
	return 0
}
