package escape_test

import (
	"testing"

	"github.com/ethanv2/podbit/data/escape"
)

// TestConvertibles runs data validation tests on escape.Convertibles.
func TestConvertibles(t *testing.T) {
	t.Run("SelfMapping", testSelfMapping)
	t.Run("ASCIIMapping", testASCIIMapping)
}

// TestSelfMapping tests if any keys in the Convertibles set map to themselves
// (should never happen).
func testSelfMapping(t *testing.T) {
	for key, val := range escape.Convertibles {
		if key == val {
			t.Errorf("selfmap: %U maps to itself (value %U)", key, val)
		}
	}
}

// TestASCIIMapping tests if any keys in the Convertibles set map to non-ASCII
// characters (should never happen). Additionally, tests if any ASCII
// characters have unwittingly been entered to be converted (again, should
// never happen).
func testASCIIMapping(t *testing.T) {
	for key, val := range escape.Convertibles {
		if int32(key) <= 127 {
			t.Errorf("asciimap: ascii value found to be converted: %U", key)
		}
		if int32(val) > 127 {
			t.Errorf("asciimap: non-ascii value found as safe conversion: %U", val)
		}
	}
}
