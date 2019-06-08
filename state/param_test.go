package state

import (
	"testing"
)

type NameIndex map[string]int

func (n NameIndex) GetIndex(k string) int {
	i, ok := n[k]
	if !ok {
		return IndexNotFound
	}
	return i
}

var defaultNameIndexer = NameIndex{
	"a": 0,
	"b": 1,
	"c": 2,
	"d": 3,
	"e": 4,
}

func TestIntParamSlice(t *testing.T) {
	vars := []int64{0, 1, 2, 3, 4}
	intparam := NewIntParam(vars, defaultNameIndexer)
	size := intparam.Len()

	testcases := []struct {
		from         int
		to           int
		okKey        string
		okKeyValue   int64
		okIndex      int
		okIndexValue int64
		ngKey        string
	}{
		{
			from:         0,
			to:           1,
			okKey:        "a",
			okKeyValue:   0,
			okIndex:      0,
			okIndexValue: 0,
			ngKey:        "b",
		},
		{
			from:         size - 1,
			to:           size,
			okKey:        "e",
			okKeyValue:   4,
			okIndex:      0,
			okIndexValue: 4,
			ngKey:        "c",
		},
	}

	for _, test := range testcases {
		sliced := intparam.Slice(test.from, test.to)
		val, ok := sliced.GetByStr(test.okKey)
		if !ok {
			t.Errorf("sliced intparam has not shared key, values %#v\n%#v", intparam, test)
		}
		if val != test.okKeyValue {
			t.Errorf("sliced.GetByStr() returns invalid value, got %v expect %v\n%#v", val, test.okKeyValue, test)
		}
		val = sliced.Get(test.okIndex)
		if val != test.okIndexValue {
			t.Errorf("sliced.Get() returns invalid value, got %v expect %v\n%#v", val, test.okIndexValue, test)
		}

		if _, ok = sliced.GetByStr(test.ngKey); ok {
			t.Errorf("sliced.GetByStr(invalid key) returns no error\n%#v", test)
		}
	}
}

func TestIntParamSliceNest(t *testing.T) {
	vars := []int64{0, 1, 2, 3, 4}
	intparam := NewIntParam(vars, defaultNameIndexer)
	sliced := intparam.Slice(1, intparam.Len())
	sliced_sliced := sliced.Slice(0, sliced.Len()-2)

	if val := sliced_sliced.Get(0); val != vars[1] {
		t.Errorf("sliced_sliced.Get() returns invalid value, got %v expect %v", val, vars[1])
	}

	val, ok := sliced_sliced.GetByStr("b")
	if !ok {
		t.Errorf("sliced_sliced has no shared key %v", "b")
	}
	if val != vars[1] {
		t.Errorf("sliced_sliced.GetByStr() returns invalid value, got %v expect %v", val, vars[1])
	}

	val, ok = sliced_sliced.GetByStr("c")
	if !ok {
		t.Errorf("sliced_sliced has no shared key %v", "c")
	}
	if val != vars[2] {
		t.Errorf("sliced_sliced.GetByStr() returns invalid value, got %v expect %v", val, vars[2])
	}

	// StrParamSliceNest is tested too.
}
func BenchmarkIntParamSliceNestN(b *testing.B) {
	vars := []int64{0, 1, 2, 3, 4}
	intparam := NewIntParam(vars, defaultNameIndexer)
	sliced := intparam.Slice(1, intparam.Len())
	sliced2 := sliced.Slice(0, sliced.Len()-1)
	sliced3 := sliced2.Slice(1, sliced2.Len()-1)

	buildBenchFunc := func(target IntParam, key string) func(*testing.B) {
		return func(b *testing.B) {
			if _, ok := target.GetByStr(key); !ok {
				b.Fatalf("can not get by key %v", key)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = target.GetByStr(key)
			}
		}
	}

	b.Run("N=0", buildBenchFunc(intparam, "c"))
	b.Run("N=1", buildBenchFunc(sliced, "c"))
	b.Run("N=2", buildBenchFunc(sliced2, "c"))
	b.Run("N=3", buildBenchFunc(sliced3, "c"))
}

func TestStrParamSlice(t *testing.T) {
	vars := []string{"0", "1", "2", "3", "4"}
	strparam := NewStrParam(vars, defaultNameIndexer)
	size := strparam.Len()

	testcases := []struct {
		from         int
		to           int
		okKey        string
		okKeyValue   string
		okIndex      int
		okIndexValue string
		ngKey        string
	}{
		{
			from:         0,
			to:           1,
			okKey:        "a",
			okKeyValue:   "0",
			okIndex:      0,
			okIndexValue: "0",
			ngKey:        "b",
		},
		{
			from:         size - 1,
			to:           size,
			okKey:        "e",
			okKeyValue:   "4",
			okIndex:      0,
			okIndexValue: "4",
			ngKey:        "c",
		},
	}

	for _, test := range testcases {
		sliced := strparam.Slice(test.from, test.to)
		val, ok := sliced.GetByStr(test.okKey)
		if !ok {
			t.Errorf("sliced strparam has not shared key, values %#v\n%#v", strparam, test)
		}
		if val != test.okKeyValue {
			t.Errorf("sliced.GetByStr() returns invalid value, got %v expect %v\n%#v", val, test.okKeyValue, test)
		}
		val = sliced.Get(test.okIndex)
		if val != test.okIndexValue {
			t.Errorf("sliced.Get() returns invalid value, got %v expect %v\n%#v", val, test.okIndexValue, test)
		}

		if _, ok = sliced.GetByStr(test.ngKey); ok {
			t.Errorf("sliced.GetByStr(invalid key) returns no error\n%#v", test)
		}
	}
}
