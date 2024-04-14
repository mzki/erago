package state

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mzki/erago/state/csv"
	"github.com/mzki/erago/util/log"
)

// Character has some parameters including csv character's parameters.
type Character struct {
	ID     int64  // ID: means reference for csv.Character. -1 means no reference.
	UID    uint64 // Uniq ID. Indicating its identity.
	IsAssi int64  // experience of assistant

	Name       string // formal name
	CallName   string // usual name
	NickName   string // use for special case
	MasterName string // call for you

	UserVariables
}

// return new Empty Character initialized only UID.
// The ID is setted -1 and its means no reference for csv.Character.
func newEmptyCharacter(uid uint64, cm *csv.CsvManager) *Character {
	c := &Character{
		ID:            -1,
		UID:           uid,
		IsAssi:        0,
		UserVariables: newUserVariablesChara(cm),
	}
	return c
}

// new Character initialized by using parameters of given csv.Character.
func newInitializedCharacter(uid uint64, chara *csv.Character, cm *csv.CsvManager) *Character {
	c := newEmptyCharacter(uid, cm)
	c.CsvInitialize(chara)
	return c
}

// Initialize by csv parameters.
func (c *Character) CsvInitialize(csv_chara *csv.Character) {
	c.ID = int64(csv_chara.ID)
	c.Name = csv_chara.Name
	c.CallName = csv_chara.CallName
	c.MasterName = csv_chara.MasterName
	c.NickName = csv_chara.NickName

	for k, csv_vars := range csv_chara.GetStrMap() {
		copy(c.StrMap[k].Values, csv_vars)
	}
	for k, csv_vars := range csv_chara.GetIntMap() {
		copy(c.IntMap[k].Values, csv_vars)
	}
}

// To marshall object, the field needs to be exported.
// But, ordinary, the field should be hidden from its users.
//
// As compromise of these, use unexported type and but the field is exported.
type characters []*Character

// Character pool. Its size are fit automatically
// to appropriate size.
//
// Adding new character is done by Characters.AddID().
type Characters struct {
	List          characters
	CountNewChara uint64 // count of call newCharacters()
	csv           *csv.CsvManager
}

// Characters's list has capacity at least minListCapacity.
const minListCapacity = 16

func newCharacters(csv *csv.CsvManager) *Characters {
	return &Characters{
		List:          make([]*Character, 0, minListCapacity),
		csv:           csv,
		CountNewChara: 0,
	}
}

func (cs *Characters) refine(csvM *csv.CsvManager) {
	cs.csv = csvM

	intVSpecs := intVariableSpecs(csvM.IntVariableSpecs(csv.ScopeChara))
	strVSpecs := strVariableSpecs(csvM.StrVariableSpecs(csv.ScopeChara))
	constants := csvM.Constants()

	for index, c := range cs.List {
		csvC, ok := csvM.CharaMap[c.ID]
		if ok {
			c.UserVariables.refineByCsvChara(constants, intVSpecs, strVSpecs, csvC)
		} else {
			log.Infof("Chara index %v: unknown character ID (%v) exist", index, c.ID)
			c.UserVariables.refine(constants, intVSpecs, strVSpecs)
		}
	}
}

// like array access a_chara = charas[i], if idx out of range return nil
func (cs Characters) Get(i int) *Character {
	if cs.inRange(i) {
		return cs.List[i]
	}
	return nil
}

// like array access charas[i] = a_chara
func (cs Characters) Set(i int, c *Character) {
	if cs.inRange(i) {
		cs.List[i] = c
	}
	// TODO: error message required?
}

func (cs Characters) inRange(idx int) bool {
	return 0 <= idx && idx < len(cs.List)
}

// a current number of character
func (cs Characters) Len() int {
	return len(cs.List)
}

// implement sort.Interface. swap positions between
// i-th character and j-th character.
func (cs Characters) Swap(i, j int) {
	cs.List[i], cs.List[j] = cs.List[j], cs.List[i]
}

// implement sort.Interface. default compare Character's ID.
func (cs Characters) Less(i, j int) bool {
	return cs.List[i].ID < cs.List[j].ID
}

// add empty Character that has not initialized parameters, and return added new Character.
func (cs *Characters) AddEmptyCharacter() *Character {
	newc := cs.newEmptyCharacter()
	cs.append(newc)
	return newc
}

// new character with increase new-count.
func (cs *Characters) newEmptyCharacter() *Character {
	cs.CountNewChara += 1
	return newEmptyCharacter(cs.CountNewChara, cs.csv)
}

// Add Characters detected by Character's ID.
// return list of added Characters and
// error that Chara is not found.
func (cs *Characters) AddIDs(IDs ...int64) ([]*Character, error) {
	if len(IDs) == 0 {
		return nil, errors.New("require at least one ID.")
	}
	added := make([]*Character, len(IDs))
	for i, id := range IDs {
		c, err := cs.AddID(id)
		if err != nil {
			return nil, err
		}
		added[i] = c
	}
	return added, nil
}

// Add one character detected by character's ID and return added new Character.
// if character of given id is not found retrun that error.
func (cs *Characters) AddID(ID int64) (*Character, error) {
	if csv_c, ok := cs.csv.CharaMap[ID]; ok {
		newc := cs.AddEmptyCharacter()
		newc.CsvInitialize(csv_c)
		return newc, nil
	}
	return nil, fmt.Errorf("chara ID %d is not found in CSV", ID)
}

// IsAddableID returns true when Character ID is addable into Characters, otherwise returns false.
func (cs *Characters) IsAddableID(ID int64) bool {
	_, ok := cs.csv.CharaMap[ID]
	return ok
}

// the size of characters increases this size.
const characterAppendSize = 16

func (cs *Characters) append(c *Character) {
	if len(cs.List) == cap(cs.List) {
		new_charas := make([]*Character, len(cs.List), len(cs.List)+characterAppendSize)
		copy(new_charas, cs.List)
		cs.List = new_charas
	}
	cs.List = append(cs.List, c)
}

// Remove Character at index and return IsRemoved.
// NOTE: after removing, the characters after given idx
// are shifted to fill empty index.
func (cs *Characters) Remove(idx int) bool {
	if !cs.inRange(idx) {
		return false
	}
	copy(cs.List[idx:], cs.List[idx+1:])
	last := cs.Len() - 1
	cs.List[last] = nil
	cs.List = cs.List[:last]
	cs.compaction()
	return true
}

func (cs *Characters) compaction() {
	clist := cs.List
	if cap(clist) == minListCapacity {
		// case: minimum capacity, do nothing
		return
	} else if len(clist) >= cap(clist)/2 {
		// case: len exceeds half of capacity, do nothing
		return
	}

	// case: len < capacity/2
	new_cap := cap(clist) / 2
	if new_cap < minListCapacity {
		new_cap = minListCapacity
	}
	new_clist := make(characters, len(clist), new_cap)
	copy(new_clist, clist)
	cs.List = new_clist
}

// clear all chara
func (cs *Characters) Clear() {
	cs.List = make([]*Character, 0, minListCapacity)
}

// sort by "less" function, which returns true
// if first character is less than second character.
// NOTE: if "less" function returns inverted booleen,
// sort order is reversed.
func (cs Characters) SortBy(by func(*Character, *Character) bool) {
	lessFunc(by).Sort(cs)
}

// reverse ordered sort by "less" func.
func (cs Characters) RevSortBy(by func(*Character, *Character) bool) {
	lessFunc(by).ReverseSort(cs)
}

// find index and Character by bool func.
// if not found, return (-1,nil)
func (cs Characters) findBy(f func(*Character) bool) (int, *Character) {
	for i, c := range cs.List {
		if f(c) {
			return i, c
		}
	}
	return -1, nil
}

// find Character by bool func.
// if not found, return nil
func (cs Characters) FindBy(f func(*Character) bool) *Character {
	_, c := cs.findBy(f)
	return c
}

// find true index by bool func.
// if not found, return -1
func (cs Characters) FindIndexBy(f func(*Character) bool) int {
	i, _ := cs.findBy(f)
	return i
}

// find index by compairing UID.
func (cs Characters) FindIndexByUID(uid uint64) int {
	return cs.FindIndexBy(func(c *Character) bool {
		return c.UID == uid
	})
}

// lessFunc converts its interface to sort.Interface.
type lessFunc func(*Character, *Character) bool

// Stable Sort for Characters.
func (by lessFunc) Sort(charas Characters) {
	data := by.Sortable(charas)
	sort.Stable(data)
}

// Reverse ordered Character
func (by lessFunc) ReverseSort(charas Characters) {
	data := by.Sortable(charas)
	sort.Stable(sort.Reverse(data))
}

// convert sortable object.
func (by lessFunc) Sortable(charas Characters) sort.Interface {
	return charaSorter{
		charas,
		by,
	}
}

// it implements sort.Interface
type charaSorter struct {
	Characters
	lessFunc
}

// implement sort.Interface
func (c charaSorter) Less(i, j int) bool {
	return c.lessFunc(c.List[i], c.List[j])
}
