package state

import (
	"testing"
)

func Test_newCharaReferences(t *testing.T) {
	charas := newCharacters(CSVDB)
	type args struct {
		n_chara int
		src     *Characters
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal", args{10, charas}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCharaReferences(tt.args.n_chara, tt.args.src); got == nil {
				t.Errorf("newCharaReferences() = %v", got)
			}
		})
	}
}

func TestCharaReferences_GetChara(t *testing.T) {
	type args struct {
		idx int
	}
	tests := []struct {
		name                string
		args                args
		preconditionAndWant func(cref *CharaReferences, charas *Characters) *Character
	}{
		{"ref 0, chara 0", args{0}, func(cref *CharaReferences, charas *Characters) *Character {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			cref.Set(0, c)
			return c
		}},
		{"ref 0, chara 1", args{0}, func(cref *CharaReferences, charas *Characters) *Character {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			cref.Set(0, c)
			return c
		}},
		{"ref 1, chara 0", args{1}, func(cref *CharaReferences, charas *Characters) *Character {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			cref.Set(1, c)
			return c
		}},
		{"ref 1, chara 1", args{1}, func(cref *CharaReferences, charas *Characters) *Character {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			cref.Set(1, c)
			return c
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(10, charas)
			want := tt.preconditionAndWant(cref, charas)
			if got := cref.GetChara(tt.args.idx); got != want {
				t.Errorf("CharaReferences.GetChara() = %p, want %p", got, want)
			}
		})
	}
}

func TestCharaReferences_GetIndex(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name                string
		args                args
		preconditionAndWant func(cref *CharaReferences, charas *Characters) int
	}{
		{"ref 0, chara 0", args{0}, func(cref *CharaReferences, charas *Characters) int {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			cref.Set(0, c)
			return 0
		}},
		{"ref 0, chara 1", args{0}, func(cref *CharaReferences, charas *Characters) int {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			cref.Set(0, c)
			return 1
		}},
		{"ref 1, chara 0", args{1}, func(cref *CharaReferences, charas *Characters) int {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			cref.Set(1, c)
			return 0
		}},
		{"ref 1, chara 1", args{1}, func(cref *CharaReferences, charas *Characters) int {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			cref.Set(1, c)
			return 1
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(10, charas)
			want := tt.preconditionAndWant(cref, charas)
			if got := cref.GetIndex(tt.args.i); got != want {
				t.Errorf("CharaReferences.GetIndex() = %v, want %v", got, want)
			}
		})
	}
}

func TestCharaReferences_Set(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name                string
		args                args
		preconditionAndWant func(cref *CharaReferences, charas *Characters) (int, *Character)
		wantErr             bool
	}{
		{"ref 0, chara 0", args{0}, func(cref *CharaReferences, charas *Characters) (int, *Character) {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			return 0, c
		}, false},
		{"ref 0, chara 1", args{0}, func(cref *CharaReferences, charas *Characters) (int, *Character) {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			return 1, c
		}, false},
		{"ref 1, chara 0", args{1}, func(cref *CharaReferences, charas *Characters) (int, *Character) {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			return 0, c
		}, false},
		{"ref 1, chara 1", args{1}, func(cref *CharaReferences, charas *Characters) (int, *Character) {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(1)
			return 1, c
		}, false},
		{"ref any, chara Not found error", args{0}, func(cref *CharaReferences, charas *Characters) (int, *Character) {
			c := charas.AddEmptyCharacter()
			charas.Remove(0)
			return 0, c
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(10, charas)
			want, argC := tt.preconditionAndWant(cref, charas)
			if err := cref.Set(tt.args.i, argC); (err != nil) != tt.wantErr {
				t.Errorf("CharaReferences.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := cref.Indexes[tt.args.i]; got != want {
				t.Errorf("CharaReferences.Set() internal idx = %v, want %v", got, want)
			}
		})
	}
}

func TestCharaReferences_First(t *testing.T) {
	tests := []struct {
		name                string
		preconditionAndWant func(cref *CharaReferences, charas *Characters) *Character
	}{
		{"normal", func(cref *CharaReferences, charas *Characters) *Character {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			c := charas.Get(0)
			return c
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(10, charas)
			want := tt.preconditionAndWant(cref, charas)
			if got := cref.First(); got != want {
				t.Errorf("CharaReferences.First() = %p, want %p", got, want)
			}
		})
	}
}

func TestCharaReferences_Len(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{10}, 10},
		{"zero", args{0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(tt.args.len, charas)
			if got := cref.Len(); got != tt.want {
				t.Errorf("CharaReferences.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCharaReferences_Clear(t *testing.T) {
	tests := []struct {
		name                string
		preconditionAndWant func(cref *CharaReferences, charas *Characters)
	}{
		{"ref 2", func(cref *CharaReferences, charas *Characters) {
			charas.AddEmptyCharacter()
			charas.AddEmptyCharacter()
			cref.Set(0, charas.Get(0))
			cref.Set(0, charas.Get(1))
		}},
		{"ref 10", func(cref *CharaReferences, charas *Characters) {
			for i := 0; i < 10; i++ {
				c := charas.AddEmptyCharacter()
				cref.Set(i, c)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charas := newCharacters(CSVDB)
			cref := newCharaReferences(10, charas)
			tt.preconditionAndWant(cref, charas)
			cref.Clear()
			for i, v := range cref.Indexes {
				if v != 0 {
					t.Errorf("CharaReferences.Clear(); Not cleared values at %v, value = %v, expect = 0", i, v)
				}
			}
		})
	}
}
