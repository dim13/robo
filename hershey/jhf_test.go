package hershey

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	// 'A' from Hershey's roman simplex, wrapped after the third pair
	f, err := parse(strings.NewReader("   32  1JZ\n   65  9I[RFJ[ \nRRFZ[ RMTWT\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := f[' '].W, 32; got != want {
		t.Errorf("space width: got %d, want %d", got, want)
	}
	want := Glyph{
		S: Set{
			{{X: -24 + top, Y: 18}, {X: 18 + top, Y: 2}},
			{{X: -24 + top, Y: 18}, {X: 18 + top, Y: 34}},
			{{X: 4 + top, Y: 8}, {X: 4 + top, Y: 28}},
		},
		W: 36,
	}
	if got := f['A']; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEmbedded(t *testing.T) {
	if got, want := len(font), 329; got != want {
		t.Errorf("got %d glyphs, want %d", got, want)
	}
}
