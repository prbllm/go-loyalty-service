package luhn

import "testing"

func TestIsValidOrderNumber(t *testing.T) {
	cases := []struct {
		number string
		valid  bool
	}{
		{"79927398713", true},
		{"4532015112830366", true},
		{" 4532015112830366 ", true},
		{"79927398710", false},
		{"abc123", false},
		{"", false},
		{"   ", false},
	}

	for _, tc := range cases {
		if got := IsValidOrderNumber(tc.number); got != tc.valid {
			t.Fatalf("number %q expected %v, got %v", tc.number, tc.valid, got)
		}
	}
}
