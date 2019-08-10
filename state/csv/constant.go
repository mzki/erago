package csv

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// It converts field specified by column i to int number and returns the number.
// It occurs panic if the passed index is in invalid range or the specified field doesn't represent the number.
func getAsInt(record []string, i int) int {
	field := record[i]
	if len(field) == 0 { // empty field is treated as 0 value.
		return 0
	}

	index, err := strconv.Atoi(field)
	if err != nil {
		panic(err)
	}
	return index
}

// Constant is a constant data set defined by a csv file.
// it must contains constant key names Names and hash map for index-key pair NameIndex.
// may contains some custom fields.
type Constant struct {
	Names
	NameIndex
	CustomFields
}

const (
	csvHeaderFieldID   = "id"
	csvHeaderFieldName = "name"

	headerFieldPrefixNum = "num_"
	headerFieldPrefixStr = "str_"
)

// read CSV file that defines names and custom fields for each variable,
// return constant or error when read failed.
func readConstantFile(file string, intBuffer []int, strBuffer []string) (*Constant, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	constant, err := readConstant(fp, intBuffer, strBuffer)
	if err != nil {
		return nil, fmt.Errorf("file(%s): %v", file, err)
	}
	return constant, nil
}

// read CSV constants from io.Reader that defines names and custom fields for each variable.
// return constant or error when read failed.
func readConstant(ioreader io.Reader, intBuffer []int, strBuffer []string) (*Constant, error) {
	reader := NewReader(ioreader)

	// parsing header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("can not parse header: %v", err)
	}

	if headers[0] != csvHeaderFieldID || headers[1] != csvHeaderFieldName {
		return nil, fmt.Errorf("header fields should starts with %s, %s, but %s, %s",
			csvHeaderFieldID, csvHeaderFieldName, headers[0], headers[1])
	}

	type parsedKey struct {
		name string
		typ  CustomFieldType
	}
	headerTypes := make([]parsedKey, 0, len(headers))

	for _, h := range headers[2:] {
		var dtype CustomFieldType
		var name string
		if prefix := headerFieldPrefixNum; strings.HasPrefix(h, prefix) {
			dtype = CFIntType
			name = strings.TrimPrefix(h, prefix)
		} else if prefix := headerFieldPrefixStr; strings.HasPrefix(h, prefix) {
			dtype = CFStrType
			name = strings.TrimPrefix(h, prefix)
		} else {
			return nil, fmt.Errorf("custom header name should starts with either of `%s` or `%s`, but got `%s`",
				headerFieldPrefixNum, headerFieldPrefixStr, h)
		}

		if len(name) == 0 {
			return nil, fmt.Errorf("custom header field should have some name. but nothing, %s", h)
		}

		headerTypes = append(headerTypes, parsedKey{
			name: name,
			typ:  dtype,
		})
	}

	customBuffers := make([]interface{}, 0, len(headerTypes))
	for _, h := range headerTypes {
		// TODO Get buffer from sync.Pool
		if h.typ == CFIntType {
			customBuffers = append(customBuffers, make([]int64, 0, len(intBuffer)))
		} else if h.typ == CFStrType {
			customBuffers = append(customBuffers, make([]string, 0, len(strBuffer)))
		} else {
			panic("should not occur")
		}
		// else case should not be occurd since such case is terminated by constructing headerTypes.
	}

	// reset buffer size
	intBuffer = intBuffer[:0]
	strBuffer = strBuffer[:0]

	var max_index = 0

	// parse content
	for record, err := reader.Read(); err != io.EOF; record, err = reader.Read() {
		// parse error occured
		if err != nil {
			return nil, err
		}
		if len(record) < len(headerTypes) {
			return nil, fmt.Errorf("missing custom field values, expect %v fields but got %v fields, %v", len(headerTypes), len(record), record)
		}

		index := getAsInt(record, 0)
		key := record[1]

		intBuffer = append(intBuffer, index)
		strBuffer = append(strBuffer, key)

		// parsing CustomFields
		const customOffset = 2
		for i, field := range record[customOffset:] {
			if h := headerTypes[i]; h.typ == CFIntType {
				cbuf := customBuffers[i].([]int64)
				customBuffers[i] = append(cbuf, int64(getAsInt(record, i+customOffset)))
			} else if h.typ == CFStrType {
				cbuf := customBuffers[i].([]string)
				customBuffers[i] = append(cbuf, field)
			} else {
				panic("should not occur")
			}
		}

		if max_index < index {
			max_index = index
		}
	}

	// store result,
	names := newNames(max_index + 1)
	for i, index := range intBuffer {
		name := strBuffer[i]
		if len(names[index]) > 0 {
			return nil, fmt.Errorf(">\"%d,%s\": csv index(%d) is already used.", index, name, index)
		}
		names[index] = name
	}

	customFields := make(map[string]slicedType, len(headerTypes))
	for i, h := range headerTypes {
		if h.typ == CFIntType {
			nums := make([]int64, len(names))
			numBuf := customBuffers[i].([]int64)
			for i, index := range intBuffer {
				nums[index] = numBuf[i]
			}
			customFields[h.name] = &Numbers{nums}
		} else if h.typ == CFStrType {
			strs := newNames(len(names))
			strBuf := customBuffers[i].([]string)
			for i, index := range intBuffer {
				strs[index] = strBuf[i]
			}
			customFields[h.name] = &Strings{strs}
		}
	}

	return &Constant{
		Names:        names,
		NameIndex:    newNameIndex(names),
		CustomFields: CustomFields{customFields},
	}, nil
}

// == Constant members declaration

// read CSV file that defines names for each variable,
// return read names and occured error.
func readNames(file string, intBuffer []int, strBuffer []string) (Names, error) {
	intBuffer = intBuffer[:0]
	strBuffer = strBuffer[:0]

	var max_index int
	err := ReadFileFunc(file, func(record []string) error {
		if len(record) < 2 { // ignore
			return nil
		}

		index := getAsInt(record, 0)
		key := record[1]

		intBuffer = append(intBuffer, index)
		strBuffer = append(strBuffer, key)

		if max_index < index {
			max_index = index
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	names := newNames(max_index + 1)
	for i, index := range intBuffer {
		name := strBuffer[i]
		if len(names[index]) > 0 {
			return nil, fmt.Errorf("file(%s), >\"%d,%s\": csv index(%d) is already used.", file, index, name, index)
		}
		names[index] = name
	}
	return names, nil
}

// Names has CSV defined names.
type Names []string

func newNames(s int) Names {
	if s < 0 {
		s = 0
	}
	return make(Names, s)
}

// as like name := names[i]
func (ns Names) Get(i int) string {
	return ns[i]
}

// return size of array.
func (ns Names) Len() int {
	return len(ns)
}

// check whether index is valid range?
func (ns Names) InRange(index int) bool {
	return 0 <= index && index < len(ns)
}

// NameIndex holds indexes corresponding to each Name defined in CSV.
type NameIndex map[string]int

// IndexNotFound is a value implying the index is not found.
const IndexNotFound int = -1

func newNameIndex(names Names) NameIndex {
	name_idx := make(NameIndex, len(names))
	for i, name := range names {
		if len(name) > 0 {
			name_idx.set(name, i)
		}
	}
	return name_idx
}

// return name is exist?
func (ni NameIndex) Has(name string) bool {
	_, ok := ni[name]
	return ok
}

// return index of name. if not found return IndexNotFound
func (ni NameIndex) GetIndex(name string) int {
	if i, ok := ni[name]; ok {
		return i
	}
	return IndexNotFound
}

func (ni NameIndex) set(name string, idx int) {
	ni[name] = idx
}

// CustomFields is a read only hash map for key and csv constant pair.
type CustomFields struct {
	slices map[string]slicedType
}

type slicedType interface{ customFieldType() CustomFieldType }

// CustomFieldType indicates the data type for a custom field in a csv constant.
type CustomFieldType int8

const (
	CFNoneType CustomFieldType = iota
	CFIntType
	CFStrType
)

// TypeOf returns custom field type for the csv constant specified by the key.
// It returns CFNoneType if the field specified by the key is not found.
func (cf *CustomFields) TypeOf(key string) CustomFieldType {
	if v, ok := cf.slices[key]; ok {
		return v.customFieldType()
	} else {
		return CFNoneType
	}
}

// Has returns whether CustomFields have data specified by key.
func (cf *CustomFields) Has(key string) bool {
	return cf.TypeOf(key) != CFNoneType
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

func (*Numbers) customFieldType() CustomFieldType { return CFIntType }

func (v *Numbers) Get(i int) int64 { return v.data[i] }

func (v *Numbers) Len() int { return len(v.data) }

// Strings is the read-only string slice treated as a string type for csv constant
type Strings struct {
	Names
}

func (*Strings) customFieldType() CustomFieldType { return CFStrType }
