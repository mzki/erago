package eragoj

func Error(err string) error {
	return error(nil)
}

type AAA struct {
	Value int32
}

func ShowAAA(a *AAA) *AAA {
	return &AAA{a.Value + 10}
}
