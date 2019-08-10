package csv

import (
	"fmt"
	"runtime"
	"testing"
)

/*
# benchmark result

## CPU Time

**EmptyInterface is 10x faster than UnionInterface**

	BenchmarkUnionInterfaceInt64	100000000	        20.4 ns/op
	BenchmarkEmptyInterfaceInt64	1000000000	         1.94 ns/op
	BenchmarkUnionInterfaceStr		100000000	        20.7 ns/op
	BenchmarkEmptyInterfaceStr		1000000000	         1.95 ns/op


## Memory Usage

**EmptyInterface is 1.15x larger memory usage than UnionInterface**

      flat  flat%   sum%        cum   cum%
  204.10MB 27.51% 27.51%   266.60MB 35.94%  github.com/mzki/erago/state/csv.BenchmarkEmptyInterfaceStr
  175.83MB 23.70% 51.21%   225.33MB 30.37%  github.com/mzki/erago/state/csv.BenchmarkUnionInterfaceStr
  134.59MB 18.14% 69.36%   134.59MB 18.14%  github.com/mzki/erago/state/csv.BenchmarkEmptyInterfaceInt64
  115.33MB 15.55% 84.90%   115.33MB 15.55%  github.com/mzki/erago/state/csv.BenchmarkUnionInterfaceInt64

*/

type UnionStrInt interface {
	Int64() int64
	Str() string
}

type I64 int64

func (i64 I64) Int64() int64 {
	return int64(i64)
}

func (i64 I64) Str() string {
	return ""
}

type Str string

func (str Str) Int64() int64 {
	return 0
}

func (str Str) Str() string {
	return string(str)
}

const SliceSize = 1000 * 1000

func BenchmarkUnionInterfaceInt64(b *testing.B) {
	unions := make([]UnionStrInt, 0, SliceSize)
	for i := 0; i < SliceSize; i++ {
		unions = append(unions, I64(i))
	}
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unions[i%SliceSize].Int64()
	}
}

func BenchmarkEmptyInterfaceInt64(b *testing.B) {
	unions := make([]interface{}, 0, SliceSize)
	for i := 0; i < SliceSize; i++ {
		unions = append(unions, int64(i))
	}
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unions[i%SliceSize].(int64)
	}
}

func BenchmarkEmptyInterfaceCallInt64(b *testing.B) {
	unions := make([]interface{}, 0, SliceSize)
	for i := 0; i < SliceSize; i++ {
		unions = append(unions, int64(i))
	}

	getInt64 := func(ary []interface{}, i int) int64 {
		return ary[i].(int64)
	}

	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getInt64(unions, i%SliceSize)
	}
}

func BenchmarkUnionInterfaceStr(b *testing.B) {
	unions := make([]UnionStrInt, 0, SliceSize)
	for i := 0; i < SliceSize; i++ {
		unions = append(unions, Str(fmt.Sprint(i)))
	}
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unions[i%SliceSize].Str()
	}
}

func BenchmarkEmptyInterfaceStr(b *testing.B) {
	unions := make([]interface{}, 0, SliceSize)
	for i := 0; i < SliceSize; i++ {
		unions = append(unions, fmt.Sprint(i))
	}
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unions[i%SliceSize].(string)
	}
}
