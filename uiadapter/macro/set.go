package macro

// Set is set of some Macros.
type Set []*Macro

// get deep copy of macro from macro Set.
func (s Set) Get(no int) (*Macro, bool) {
	if no < 0 || len(s) <= no {
		return nil, false
	}
	m := s[no]
	new_toks := make([]Token, len(m.Tokens))
	for i, tok := range m.Tokens {
		new_toks[i] = tok
	}
	return &Macro{new_toks}, true
}

// Default macro set
var DefaultSet Set = make([]*Macro, 0)

func LoadDefault(file string) error {
	// TODO implement load macro.tml
	return nil
}
