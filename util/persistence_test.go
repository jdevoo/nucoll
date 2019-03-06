package util

import (
	"reflect"
	"testing"
)

func TestQueryReader(t *testing.T) {
	var tests = []struct {
		handle          string
		firstHandleOnly bool
		expected        []string
	}{
		{"query1", true, []string{"caterpillarinc", "nike", "total"}},
		{"query2", false, []string{"caterpillarinc", "nike", "total", "disney"}},
	}

	for _, test := range tests {
		actual, err := QueryReader(test.handle, test.firstHandleOnly)
		if err != nil {
			t.Log(err)
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Fatalf("TwtReader(%s, %t): expected %v, actual %v", test.handle, test.firstHandleOnly, test.expected, actual)
		} else {
			t.Logf("TwtReader(%s, %t): %v", test.handle, test.firstHandleOnly, test.expected)
		}
	}
}

func TestCSVReader(t *testing.T) {
	type record struct {
		FieldA string
		FieldB int
		FieldC uint64
	}

	var tests = []struct {
		filename string
		expected []record
	}{
		{"file1", []record{
			{"first", 12, 18446744073709551615},
			{"second", 22, 18446744073709551614},
			{"third", 32, 18446744073709551613},
		}},
		{"file2", []record{
			{"first", 12, 18446744073709551615},
			{"third", 32, 18446744073709551613},
		}},
		{"file3", []record{
			{"", 0, 0},
			{"", 0, 0},
			{"", 0, 0},
			{"", 0, 0},
		}},
	}

	for _, test := range tests {
		data := []record{}
		err := CSVReader(test.filename, DatExt, &data)
		if err != nil {
			t.Log(err)
		}
		if !reflect.DeepEqual(data, test.expected) {
			t.Fatalf("CSVReader %s: expected %v, actual %v", test.filename, test.expected, data)
		} else {
			t.Logf("CSVReader %s: %v", test.filename, test.expected)
		}
	}
}
