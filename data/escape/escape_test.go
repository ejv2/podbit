package escape_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ethanv2/podbit/data/escape"
)

var EscapeTests = []struct {
	Orig, Expects string
}{
	{"", ""},
	{"    ", "    "},
	{"The quick brown fox jumps over the lazy dog", "The quick brown fox jumps over the lazy dog"},
	{"The quick brown fox’s lazy jumping dog", "The quick brown fox's lazy jumping dog"},
	{"ＡＢＣＤＥＦ", "ABCDEF"},
	{"0xＢＡＤＦ00Ｄ", "0xBADF00D"},
	{"ｅthａn_v2", "ethan_v2"},
}

// TestConvertibles runs data validation tests on escape.Convertibles.
func TestConvertibles(t *testing.T) {
	t.Run("SelfMapping", testSelfMapping)
	t.Run("ASCIIMapping", testASCIIMapping)
}

// testSelfMapping tests if any keys in the Convertibles set map to themselves
// (should never happen).
func testSelfMapping(t *testing.T) {
	for key, val := range escape.Convertibles {
		if key == val {
			t.Errorf("selfmap: %U maps to itself (value %U)", key, val)
		}
	}
}

// testASCIIMapping tests if any keys in the Convertibles set map to non-ASCII
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

func TestEscape(t *testing.T) {
	t.Run("EscapeString", testEscapeString)
	t.Run("EscapeBytes", testEscapeBytes)
}

// testEscapeString tests if each given input string can be mapped to the
// correct output string using only safe characters. The returned must be fully
// valid UTF-8.
func testEscapeString(t *testing.T) {
	for _, elem := range EscapeTests {
		e := escape.Escape(elem.Orig)
		if e != elem.Expects {
			t.Errorf("escape: bad escape: expected %q, got %q", elem.Expects, e)
		}
		if !utf8.ValidString(e) {
			t.Errorf("escape: bad escape: invalid UTF-8: %s", e)
		}
	}
}

// testEscapeBytes runs the same tests as testEscapeString, but converts each
// string to a slice of bytes first, expecting the exact same results, but in
// byte form. Returned bytes must be fully valid UTF-8.
func testEscapeBytes(t *testing.T) {
	for _, elem := range EscapeTests {
		by := []byte(elem.Orig)
		e := escape.EscapeBytes(by)

		if len(e) != len(elem.Expects) {
			t.Errorf("escape: bad escape: expected len %d, got %d", len(elem.Expects), len(e))
		}
		for i, entry := range e {
			if entry != elem.Expects[i] {
				t.Errorf("escape: bad escape: expected %q, got %q", elem.Expects, e)
			}
		}

		if !utf8.Valid(e) {
			t.Errorf("escape: bad escape: invalid UTF-8: %s", string(e))
		}
	}
}

func FuzzEscape(f *testing.F) {
	for _, elem := range EscapeTests {
		f.Add(elem.Orig)
	}

	f.Fuzz(func(t *testing.T, a string) {
		out := escape.Escape(a)
		if !utf8.ValidString(out) {
			t.Errorf("escape: bad escape fuzz: unvalid UTF-8")
		}

		for key := range escape.Convertibles {
			if strings.ContainsRune(out, key) {
				t.Errorf("escape: bad escape fuzz: still contains forbidden rune")
			}
		}
	})
}
