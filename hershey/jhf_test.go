package hershey

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	// 'A' from Hershey's roman simplex, wrapped after the third pair
	f, err := parse(strings.NewReader("   32  1JZ\n  501  9I[RFJ[ \nRRFZ[ RMTWT\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := f[' '].W, 16; got != want {
		t.Errorf("space width: got %d, want %d", got, want)
	}
	want := Glyph{
		S: Set{
			{{X: -12 + top, Y: 9}, {X: 9 + top, Y: 1}},
			{{X: -12 + top, Y: 9}, {X: 9 + top, Y: 17}},
			{{X: 2 + top, Y: 4}, {X: 2 + top, Y: 14}},
		},
		W: 18,
	}
	if got := f['!']; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEmbedded(t *testing.T) {
	if got, want := len(font), 96; got != want {
		t.Errorf("got %d glyphs, want %d", got, want)
	}
}
