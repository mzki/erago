package state

import (
	"errors"
	"fmt"
)

// It has references for *Character.
// Its capacity of array is a number of csv.Character.
//
// The reference is actually index of []*Character.
// So the reference is not always equal to *Character,
// which case is occurred if []*Character is sorted,
// and relation between index and *Character are changed.
type CharaReferences struct {
	Indexes []int
	src     *Characters
}

// n_chara is expected to a number of csv.Character's.
func newCharaReferences(n_chara int, src *Characters) *CharaReferences {
	return &CharaReferences{
		Indexes: make([]int, n_chara),
		src:     src,
	}
}

// get character using idx, as like chara = reference[i],
// exception that if index out of range return nil.
func (cref CharaReferences) GetChara(idx int) *Character {
	if !cref.checkRange(idx) {
		return nil
	}
	srcI := cref.GetIndex(idx)
	return cref.src.Get(srcI)
}

// get index of Character at i. if i is out of range return -1.
func (cref CharaReferences) GetIndex(i int) int {
	if !cref.checkRange(i) {
		return -1
	}
	return cref.Indexes[i]
}

func (cref CharaReferences) checkRange(idx int) bool {
	return 0 <= idx && idx < len(cref.Indexes)
}

// set character with index. same as reference[i] = chara.
// if given character is not found in original chara list,
// return that error.
func (cref *CharaReferences) Set(i int, c *Character) error {
	if c == nil {
		return errors.New("CharaReferences.Set: nil character is not accepted")
	} else if !cref.checkRange(i) {
		return errors.New("CharaReferences.Set: index out of range")
	}

	chara_i := cref.src.FindIndexByUID(c.UID)
	if chara_i < 0 {
		return fmt.Errorf("CharaReferences.Set: %v is not found in current Chara list", c.Name)
	}
	cref.Indexes[i] = chara_i
	return nil
}

// get first character
func (cref CharaReferences) First() *Character {
	return cref.GetChara(0)
}

// return its array size.
func (cref CharaReferences) Len() int {
	return len(cref.Indexes)
}

// all Indexes are cleared by zero.
func (cref *CharaReferences) Clear() {
	for i, _ := range cref.Indexes {
		cref.Indexes[i] = 0
	}
}
