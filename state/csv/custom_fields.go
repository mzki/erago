package csv

// CustomFields is a read only hash map for key and csv constant pair.
type CustomFields struct {
	slices map[string]slicedType
}

type slicedType interface{ customFieldType() CustomFieldType }

// CustomFieldType indicates the data type for a custom field in a csv constant.
type CustomFieldType int8

const (
	NoneType CustomFieldType = iota
	IntType
	StrType
)

// TypeOf returns custom field type for the csv constant specified by the key.
func (cf *CustomFields) TypeOf(key string) CustomFieldType {
	if v, ok := cf.slices[key]; ok {
		return v.customFieldType()
	} else {
		return NoneType
	}
}

func keyPanic(key string) {
	panic("csv: accessing with missing key(" + key + ")")
}

// MustNumbers returns csv constant with number type specified by the key.
// If the constant is not a number type, it will panic.
func (cf *CustomFields) MustNumbers(key string) *Numbers {
	v, ok := cf.Numbers(key)
	if !ok {
		keyPanic(key)
	}
	return v
}

// Numbers returns csv constant with number type specified by the key.
// If the constant is not a number type, it will returns false as 2nd return value.
func (cf *CustomFields) Numbers(key string) (*Numbers, bool) {
	v, ok := cf.slices[key]
	if !ok {
		return nil, false
	}

	if ns, vok := v.(*Numbers); vok {
		return ns, true
	} else {
		return nil, false
	}
}

// MustStrings returns csv constant with string type specified by the key.
// If the constant is not a string type, it will panic.
func (cf *CustomFields) MustStrings(key string) *Strings {
	v, ok := cf.Strings(key)
	if !ok {
		keyPanic(key)
	}
	return v
}

// MustStrings returns csv constant with string type specified by the key.
// If the constant is not a string type, it will returns false as 2nd return value.
func (cf *CustomFields) Strings(key string) (*Strings, bool) {
	v, ok := cf.slices[key]
	if !ok {
		return nil, false
	}

	if ns, vok := v.(*Strings); vok {
		return ns, true
	} else {
		return nil, false
	}
}

// Numbers is the read-only int64 slice treated as a number type for csv constant
type Numbers struct {
	data []int64
}

func (*Numbers) customFieldType() CustomFieldType { return IntType }

func (v *Numbers) Get(i int) int64 { return v.data[i] }

func (v *Numbers) Len() int { return len(v.data) }

// Strings is the read-only string slice treated as a string type for csv constant
type Strings struct {
	Names
}

func (*Strings) customFieldType() CustomFieldType { return StrType }
