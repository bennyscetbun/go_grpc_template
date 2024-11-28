package apihelpers

import "testing"

func TestIsValidPassword(t *testing.T) {
	tests := map[string]bool{
		"123456789Ab@": true,
		"aB&123456789": true,
		"123456789Ab@3456789012345678901234567890123456789": true,

		"123456789Ab@34567890123456789012345678901234567890": false,
		"123456789Ab2": false,
		"123456789AB@": false,
		"123456789ab@": false,
		"12345678Ab@":  false,
	}

	for str, b := range tests {
		t.Run("TestIsValidPassword_"+str, func(t *testing.T) {
			if IsValidPassword(str) != b {
				t.Error("Expected", b)
			}
		})
	}
}
