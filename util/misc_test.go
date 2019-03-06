package util

import "testing"

func TestExists(t *testing.T) {
	var tests = []struct {
		input    string
		handles  []string
		expected bool
	}{
		{"@jdevoo", []string{"@jdevoo", "@oazanon"}, true},
		{"@jdevoo", []string{"@oazanon"}, false},
		{"@jdevoo", []string{}, false},
	}

	for _, test := range tests {
		actual := Exists(test.input, test.handles)
		if actual != test.expected {
			t.Fatalf("Exists(%s, %v): expected %t, actual %t", test.input, test.handles, test.expected, actual)
		} else {
			t.Logf("Exists(%s, %v): %t", test.input, test.handles, test.expected)
		}
	}
}

func TestDigitsOnly(t *testing.T) {
	var tests = []struct {
		input    string
		expected bool
	}{
		{"@jdevoo", false},
		{"123", true},
	}

	for _, test := range tests {
		actual := DigitsOnly(test.input)
		if actual != test.expected {
			t.Fatalf("DigitsOnly(%s): expected %t, actual %t", test.input, actual, test.expected)
		} else {
			t.Logf("DigitsOnly(%s): %t", test.input, test.expected)
		}
	}
}
