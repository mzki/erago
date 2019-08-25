package csv

import (
	"strings"
	"testing"
)

func TestReadNames(t *testing.T) {
	_, err := readNames("not-exists.file", make([]int, 0), make([]string, 0))
	if err == nil {
		t.Fatal("Fatal Error: processing non exist file but non-error")
	}
}

func TestReadConstant(t *testing.T) {
	var constText = `id, name, int_value1, str_value2
0, fisrt, 10, 20
1, second,20, 40
10,third, 30, 60`
	reader := strings.NewReader(constText)
	ibuf := make([]int, 100)
	sbuf := make([]string, 100)
	constant, err := readConstant(reader, ibuf, sbuf)
	if err != nil {
		t.Fatal(err)
	}
	if got, expect := constant.Names[10], "third"; got != expect {
		t.Errorf("differenct name, expect %v, got %v", expect, got)
	}
	if got, expect := constant.NameIndex.GetIndex("second"), 1; got != expect {
		t.Errorf("differenct nameindex, expect %v, got %v", expect, got)
	}
	if got, expect := constant.CustomFields.MustInts("value1").Get(1), int64(20); got != expect {
		t.Errorf("differenct customFieldValue, expect %v, got %v", expect, got)
	}
	if got, expect := constant.CustomFields.MustStrings("value2").Get(10), "60"; got != expect {
		t.Errorf("differenct customFieldValue, expect %v, got %v", expect, got)
	}
}

func TestReadConstantInvalidHeader(t *testing.T) {
	var constText = `id, num_value1, str_value2
0, fisrt, 10, 20
1, second,20, 40
10,third, 30, 60`
	reader := strings.NewReader(constText)
	ibuf := make([]int, 100)
	sbuf := make([]string, 100)
	constant, err := readConstant(reader, ibuf, sbuf)
	if err == nil {
		t.Error("header does not have id, name fields but no error")
		t.Logf("return value is: %v", constant)
	}
}

func BenchmarkNameIndexOf(b *testing.B) {
	cm, err := newCsvManagerInited()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := cm.NameIndexOf("Base", "体力")
		_ = idx
	}
}

func BenchmarkDirectNameIndex(b *testing.B) {
	cm, err := newCsvManagerInited()
	if err != nil {
		b.Fatal(err)
	}
	base_idxs := cm.constants["Base"].NameIndex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := base_idxs["体力"]
		_ = idx
	}
}
