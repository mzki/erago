package state

// name indexer is mapping string key to index.
type NameIndexer interface {
	// get index by querying string name
	// if not found return -1
	GetIndex(string) int
}

// it always returns index -1 for any key.
// to use it, simply none_indexer := NoneNameIndexer{}.
type NoneNameIndexer struct{}

func (n NoneNameIndexer) GetIndex(key string) int {
	return -1
}

// IntParam can be treated as []int64.
// And can use string key.
type IntParam struct {
	Values      []int64     // it must be exported to marshall object.
	nameIndexer NameIndexer // it must not be exported to marshall.
}

func NewIntParam(vars []int64, indexer NameIndexer) IntParam {
	return IntParam{
		Values:      vars,
		nameIndexer: indexer,
	}
}

// return size of its values.
func (ip IntParam) Len() int {
	return len(ip.Values)
}

// it is same as ip.Values[i].
func (ip IntParam) Get(i int) int64 {
	return ip.Values[i]
}

// it is same as ip.Values[i] = val.
func (ip IntParam) Set(i int, val int64) {
	ip.Values[i] = val
}

// get index by string key.
func (ip IntParam) GetIndex(key string) int {
	return ip.nameIndexer.GetIndex(key)
}

// same as io.Values[i] but i is obtained by
// using string key,
func (ip IntParam) GetByStr(key string) (int64, bool) {
	i := ip.GetIndex(key)
	if i == -1 {
		return -1, false
	}
	return ip.Values[i], true
}

// same as io.Values[i] = val but i is obtained by
// using string key,
func (ip IntParam) SetByStr(key string, val int64) bool {
	i := ip.GetIndex(key)
	if i == -1 {
		return false
	}
	ip.Values[i] = val
	return true
}

// same as []int[from:to], but taking over nameIndexer.
func (ip IntParam) Slice(from, to int) IntParam {
	return NewIntParam(ip.Values[from:to], ip.nameIndexer)
}

// It fills by given value to all values contained in IntParam.
func (ip IntParam) Fill(value int64) {
	for i := 0; i < len(ip.Values); i++ {
		ip.Values[i] = value
	}
}

// StrParam can be treated as []string.
// And can use string key.
type StrParam struct {
	Values      []string
	nameIndexer NameIndexer
}

func NewStrParam(vars []string, indexer NameIndexer) StrParam {
	return StrParam{
		Values:      vars,
		nameIndexer: indexer,
	}
}

// return size of its values.
func (ip StrParam) Len() int {
	return len(ip.Values)
}

// it is same as ip.Values[i].
func (ip StrParam) Get(i int) string {
	return ip.Values[i]
}

// it is same as ip.Values[i] = val.
func (ip StrParam) Set(i int, val string) {
	ip.Values[i] = val
}

// get index by string key.
func (ip StrParam) GetIndex(key string) int {
	return ip.nameIndexer.GetIndex(key)
}

// same as io.Values[i] but i is obtained by
// using string key,
func (ip StrParam) GetByStr(key string) (string, bool) {
	i := ip.GetIndex(key)
	if i == -1 {
		return "", false
	}
	return ip.Values[i], true
}

// same as io.Values[i] = val but i is obtained by
// using string key,
func (ip StrParam) SetByStr(key string, val string) bool {
	i := ip.GetIndex(key)
	if i == -1 {
		return false
	}
	ip.Values[i] = val
	return true
}

// same as []string[from:to], but taking over nameIndexer.
func (ip StrParam) Slice(from, to int) StrParam {
	return NewStrParam(ip.Values[from:to], ip.nameIndexer)
}

// It fills by given value to all values contained in IntParam.
func (ip StrParam) Fill(value string) {
	for i := 0; i < len(ip.Values); i++ {
		ip.Values[i] = value
	}
}
