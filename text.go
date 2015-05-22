package robo

import (
	"bufio"
	"io"
	"log"
)

type Font map[rune]Glyph

type Glyph struct {
	S Set
	W Unit
}

type Set []Path

func Print(c *bufio.Writer, in io.Reader, scale Unit) {
	var off Point

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		font.putchar(c, scanner.Text(), scale, &off)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (f Font) putchar(c *bufio.Writer, s string, scale Unit, off *Point) {
	for _, ch := range s {
		gl, ok := f[ch]
		if ok {
			if off.Y+gl.W*scale >= 4000 {
				off.X += height * scale
				off.Y = 0
			}
			off.Offset(c)
			for _, p := range gl.S {
				if scale > 10 {
					p.Scale(scale).Curve(c, 0)
				} else {
					p.Scale(scale).Line(c)
				}
			}
			off.Y += gl.W * scale
		} else if ch == '\t' {
			tab := 8 * f['W'].W * scale // widest char
			pos := int(off.Y / tab)
			off.Y = Unit(pos+1) * tab
		}
	}
	off.X += height * scale
	off.Y = 0
}

var height = Unit(72)
var font = Font{
	' ': Glyph{
		S: Set{},
		W: 32,
	},
	'!': Glyph{
		S: Set{
			Path{
				Point{8, 10},
				Point{36, 10},
			},
			Path{
				Point{46, 10},
				Point{48, 8},
				Point{50, 10},
				Point{48, 12},
				Point{46, 10},
			},
		},
		W: 20,
	},
	'"': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{22, 8},
			},
			Path{
				Point{8, 24},
				Point{22, 24},
			},
		},
		W: 32,
	},
	'#': Glyph{
		S: Set{
			Path{
				Point{0, 23},
				Point{64, 9},
			},
			Path{
				Point{0, 35},
				Point{64, 21},
			},
			Path{
				Point{26, 9},
				Point{26, 37},
			},
			Path{
				Point{38, 7},
				Point{38, 35},
			},
		},
		W: 42,
	},
	'$': Glyph{
		S: Set{
			Path{
				Point{0, 16},
				Point{58, 16},
			},
			Path{
				Point{0, 24},
				Point{58, 24},
			},
			Path{
				Point{14, 34},
				Point{10, 30},
				Point{8, 24},
				Point{8, 16},
				Point{10, 10},
				Point{14, 6},
				Point{18, 6},
				Point{22, 8},
				Point{24, 10},
				Point{26, 14},
				Point{30, 26},
				Point{32, 30},
				Point{34, 32},
				Point{38, 34},
				Point{44, 34},
				Point{48, 30},
				Point{50, 24},
				Point{50, 16},
				Point{48, 10},
				Point{44, 6},
			},
		},
		W: 40,
	},
	'%': Glyph{
		S: Set{
			Path{
				Point{8, 42},
				Point{50, 6},
			},
			Path{
				Point{8, 16},
				Point{12, 20},
				Point{16, 20},
				Point{20, 18},
				Point{22, 14},
				Point{22, 10},
				Point{18, 6},
				Point{14, 6},
				Point{10, 8},
				Point{8, 12},
				Point{8, 16},
				Point{10, 20},
				Point{12, 26},
				Point{12, 32},
				Point{10, 38},
				Point{8, 42},
			},
			Path{
				Point{36, 34},
				Point{38, 30},
				Point{42, 28},
				Point{46, 28},
				Point{50, 32},
				Point{50, 36},
				Point{48, 40},
				Point{44, 42},
				Point{40, 42},
				Point{36, 38},
				Point{36, 34},
			},
		},
		W: 48,
	},
	'&': Glyph{
		S: Set{
			Path{
				Point{26, 46},
				Point{24, 46},
				Point{22, 44},
				Point{22, 42},
				Point{24, 40},
				Point{28, 38},
				Point{38, 34},
				Point{44, 30},
				Point{48, 26},
				Point{50, 22},
				Point{50, 14},
				Point{48, 10},
				Point{46, 8},
				Point{42, 6},
				Point{38, 6},
				Point{34, 8},
				Point{32, 10},
				Point{24, 24},
				Point{22, 26},
				Point{18, 28},
				Point{14, 28},
				Point{10, 26},
				Point{8, 22},
				Point{10, 18},
				Point{14, 16},
				Point{18, 16},
				Point{24, 18},
				Point{30, 22},
				Point{44, 32},
				Point{48, 36},
				Point{50, 40},
				Point{50, 44},
				Point{48, 46},
				Point{46, 46},
			},
		},
		W: 52,
	},
	'\'': Glyph{
		S: Set{
			Path{
				Point{12, 10},
				Point{10, 8},
				Point{8, 10},
				Point{10, 12},
				Point{14, 12},
				Point{18, 10},
				Point{20, 8},
			},
		},
		W: 20,
	},
	'(': Glyph{
		S: Set{
			Path{
				Point{0, 22},
				Point{4, 18},
				Point{10, 14},
				Point{18, 10},
				Point{28, 8},
				Point{36, 8},
				Point{46, 10},
				Point{54, 14},
				Point{60, 18},
				Point{64, 22},
			},
		},
		W: 28,
	},
	')': Glyph{
		S: Set{
			Path{
				Point{0, 6},
				Point{4, 10},
				Point{10, 14},
				Point{18, 18},
				Point{28, 20},
				Point{36, 20},
				Point{46, 18},
				Point{54, 14},
				Point{60, 10},
				Point{64, 6},
			},
		},
		W: 28,
	},
	'*': Glyph{
		S: Set{
			Path{
				Point{8, 16},
				Point{32, 16},
			},
			Path{
				Point{14, 6},
				Point{26, 26},
			},
			Path{
				Point{14, 26},
				Point{26, 6},
			},
		},
		W: 32,
	},
	'+': Glyph{
		S: Set{
			Path{
				Point{14, 26},
				Point{50, 26},
			},
			Path{
				Point{32, 8},
				Point{32, 44},
			},
		},
		W: 52,
	},
	',': Glyph{
		S: Set{
			Path{
				Point{48, 12},
				Point{50, 10},
				Point{48, 8},
				Point{46, 10},
				Point{48, 12},
				Point{52, 12},
				Point{56, 10},
				Point{58, 8},
			},
		},
		W: 20,
	},
	'-': Glyph{
		S: Set{
			Path{
				Point{32, 8},
				Point{32, 44},
			},
		},
		W: 52,
	},
	'.': Glyph{
		S: Set{
			Path{
				Point{46, 10},
				Point{48, 8},
				Point{50, 10},
				Point{48, 12},
				Point{46, 10},
			},
		},
		W: 20,
	},
	'/': Glyph{
		S: Set{
			Path{
				Point{0, 40},
				Point{64, 4},
			},
		},
		W: 44,
	},
	'0': Glyph{
		S: Set{
			Path{
				Point{8, 18},
				Point{10, 12},
				Point{16, 8},
				Point{26, 6},
				Point{32, 6},
				Point{42, 8},
				Point{48, 12},
				Point{50, 18},
				Point{50, 22},
				Point{48, 28},
				Point{42, 32},
				Point{32, 34},
				Point{26, 34},
				Point{16, 32},
				Point{10, 28},
				Point{8, 22},
				Point{8, 18},
			},
		},
		W: 40,
	},
	'1': Glyph{
		S: Set{
			Path{
				Point{16, 12},
				Point{14, 16},
				Point{8, 22},
				Point{50, 22},
			},
		},
		W: 40,
	},
	'2': Glyph{
		S: Set{
			Path{
				Point{18, 8},
				Point{16, 8},
				Point{12, 10},
				Point{10, 12},
				Point{8, 16},
				Point{8, 24},
				Point{10, 28},
				Point{12, 30},
				Point{16, 32},
				Point{20, 32},
				Point{24, 30},
				Point{30, 26},
				Point{50, 6},
				Point{50, 34},
			},
		},
		W: 40,
	},
	'3': Glyph{
		S: Set{
			Path{
				Point{8, 10},
				Point{8, 32},
				Point{24, 20},
				Point{24, 26},
				Point{26, 30},
				Point{28, 32},
				Point{34, 34},
				Point{38, 34},
				Point{44, 32},
				Point{48, 28},
				Point{50, 22},
				Point{50, 16},
				Point{48, 10},
				Point{46, 8},
				Point{42, 6},
			},
		},
		W: 40,
	},
	'4': Glyph{
		S: Set{
			Path{
				Point{8, 26},
				Point{36, 6},
				Point{36, 36},
			},
			Path{
				Point{8, 26},
				Point{50, 26},
			},
		},
		W: 40,
	},
	'5': Glyph{
		S: Set{
			Path{
				Point{8, 30},
				Point{8, 10},
				Point{26, 8},
				Point{24, 10},
				Point{22, 16},
				Point{22, 22},
				Point{24, 28},
				Point{28, 32},
				Point{34, 34},
				Point{38, 34},
				Point{44, 32},
				Point{48, 28},
				Point{50, 22},
				Point{50, 16},
				Point{48, 10},
				Point{46, 8},
				Point{42, 6},
			},
		},
		W: 40,
	},
	'6': Glyph{
		S: Set{
			Path{
				Point{14, 32},
				Point{10, 30},
				Point{8, 24},
				Point{8, 20},
				Point{10, 14},
				Point{16, 10},
				Point{26, 8},
				Point{36, 8},
				Point{44, 10},
				Point{48, 14},
				Point{50, 20},
				Point{50, 22},
				Point{48, 28},
				Point{44, 32},
				Point{38, 34},
				Point{36, 34},
				Point{30, 32},
				Point{26, 28},
				Point{24, 22},
				Point{24, 20},
				Point{26, 14},
				Point{30, 10},
				Point{36, 8},
			},
		},
		W: 40,
	},
	'7': Glyph{
		S: Set{
			Path{
				Point{8, 34},
				Point{50, 14},
			},
			Path{
				Point{8, 6},
				Point{8, 34},
			},
		},
		W: 40,
	},
	'8': Glyph{
		S: Set{
			Path{
				Point{8, 16},
				Point{10, 10},
				Point{14, 8},
				Point{18, 8},
				Point{22, 10},
				Point{24, 14},
				Point{26, 22},
				Point{28, 28},
				Point{32, 32},
				Point{36, 34},
				Point{42, 34},
				Point{46, 32},
				Point{48, 30},
				Point{50, 24},
				Point{50, 16},
				Point{48, 10},
				Point{46, 8},
				Point{42, 6},
				Point{36, 6},
				Point{32, 8},
				Point{28, 12},
				Point{26, 18},
				Point{24, 26},
				Point{22, 30},
				Point{18, 32},
				Point{14, 32},
				Point{10, 30},
				Point{8, 24},
				Point{8, 16},
			},
		},
		W: 40,
	},
	'9': Glyph{
		S: Set{
			Path{
				Point{22, 32},
				Point{28, 30},
				Point{32, 26},
				Point{34, 20},
				Point{34, 18},
				Point{32, 12},
				Point{28, 8},
				Point{22, 6},
				Point{20, 6},
				Point{14, 8},
				Point{10, 12},
				Point{8, 18},
				Point{8, 20},
				Point{10, 26},
				Point{14, 30},
				Point{22, 32},
				Point{32, 32},
				Point{42, 30},
				Point{48, 26},
				Point{50, 20},
				Point{50, 16},
				Point{48, 10},
				Point{44, 8},
			},
		},
		W: 40,
	},
	':': Glyph{
		S: Set{
			Path{
				Point{22, 10},
				Point{24, 8},
				Point{26, 10},
				Point{24, 12},
				Point{22, 10},
			},
			Path{
				Point{46, 10},
				Point{48, 8},
				Point{50, 10},
				Point{48, 12},
				Point{46, 10},
			},
		},
		W: 20,
	},
	';': Glyph{
		S: Set{
			Path{
				Point{22, 10},
				Point{24, 8},
				Point{26, 10},
				Point{24, 12},
				Point{22, 10},
			},
			Path{
				Point{48, 12},
				Point{50, 10},
				Point{48, 8},
				Point{46, 10},
				Point{48, 12},
				Point{52, 12},
				Point{56, 10},
				Point{58, 8},
			},
		},
		W: 20,
	},
	'<': Glyph{
		S: Set{
			Path{
				Point{14, 40},
				Point{32, 8},
				Point{50, 40},
			},
		},
		W: 48,
	},
	'=': Glyph{
		S: Set{
			Path{
				Point{26, 8},
				Point{26, 44},
			},
			Path{
				Point{38, 8},
				Point{38, 44},
			},
		},
		W: 52,
	},
	'>': Glyph{
		S: Set{
			Path{
				Point{14, 8},
				Point{32, 40},
				Point{50, 8},
			},
		},
		W: 48,
	},
	'?': Glyph{
		S: Set{
			Path{
				Point{18, 6},
				Point{16, 6},
				Point{12, 8},
				Point{10, 10},
				Point{8, 14},
				Point{8, 22},
				Point{10, 26},
				Point{12, 28},
				Point{16, 30},
				Point{20, 30},
				Point{24, 28},
				Point{26, 26},
				Point{30, 18},
				Point{36, 18},
			},
			Path{
				Point{46, 18},
				Point{48, 16},
				Point{50, 18},
				Point{48, 20},
				Point{46, 18},
			},
		},
		W: 36,
	},
	'@': Glyph{
		S: Set{
			Path{
				Point{24, 37},
				Point{20, 35},
				Point{18, 31},
				Point{18, 25},
				Point{20, 21},
				Point{22, 19},
				Point{28, 17},
				Point{34, 17},
				Point{38, 19},
				Point{40, 23},
				Point{40, 29},
				Point{38, 33},
				Point{34, 35},
			},
			Path{
				Point{18, 25},
				Point{22, 21},
				Point{28, 19},
				Point{34, 19},
				Point{38, 21},
				Point{40, 23},
			},
			Path{
				Point{18, 37},
				Point{34, 35},
				Point{38, 35},
				Point{40, 39},
				Point{40, 43},
				Point{36, 47},
				Point{30, 49},
				Point{26, 49},
				Point{20, 47},
				Point{16, 45},
				Point{12, 41},
				Point{10, 37},
				Point{8, 31},
				Point{8, 25},
				Point{10, 19},
				Point{12, 15},
				Point{16, 11},
				Point{20, 9},
				Point{26, 7},
				Point{32, 7},
				Point{38, 9},
				Point{42, 11},
				Point{46, 15},
				Point{48, 19},
				Point{50, 25},
				Point{50, 31},
				Point{48, 37},
				Point{46, 41},
				Point{44, 43},
			},
			Path{
				Point{18, 39},
				Point{34, 37},
				Point{38, 37},
				Point{40, 39},
			},
		},
		W: 54,
	},
	'A': Glyph{
		S: Set{
			Path{
				Point{8, 18},
				Point{50, 2},
			},
			Path{
				Point{8, 18},
				Point{50, 34},
			},
			Path{
				Point{36, 8},
				Point{36, 28},
			},
		},
		W: 36,
	},
	'B': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 7},
				Point{8, 25},
				Point{10, 31},
				Point{12, 33},
				Point{16, 35},
				Point{20, 35},
				Point{24, 33},
				Point{26, 31},
				Point{28, 25},
			},
			Path{
				Point{28, 7},
				Point{28, 25},
				Point{30, 31},
				Point{32, 33},
				Point{36, 35},
				Point{42, 35},
				Point{46, 33},
				Point{48, 31},
				Point{50, 25},
				Point{50, 7},
			},
		},
		W: 42,
	},
	'C': Glyph{
		S: Set{
			Path{
				Point{18, 37},
				Point{14, 35},
				Point{10, 31},
				Point{8, 27},
				Point{8, 19},
				Point{10, 15},
				Point{14, 11},
				Point{18, 9},
				Point{24, 7},
				Point{34, 7},
				Point{40, 9},
				Point{44, 11},
				Point{48, 15},
				Point{50, 19},
				Point{50, 27},
				Point{48, 31},
				Point{44, 35},
				Point{40, 37},
			},
		},
		W: 42,
	},
	'D': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 7},
				Point{8, 21},
				Point{10, 27},
				Point{14, 31},
				Point{18, 33},
				Point{24, 35},
				Point{34, 35},
				Point{40, 33},
				Point{44, 31},
				Point{48, 27},
				Point{50, 21},
				Point{50, 7},
			},
		},
		W: 42,
	},
	'E': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 7},
				Point{8, 33},
			},
			Path{
				Point{28, 7},
				Point{28, 23},
			},
			Path{
				Point{50, 7},
				Point{50, 33},
			},
		},
		W: 38,
	},
	'F': Glyph{
		S: Set{
			Path{
				Point{8, 6},
				Point{50, 6},
			},
			Path{
				Point{8, 6},
				Point{8, 32},
			},
			Path{
				Point{28, 6},
				Point{28, 22},
			},
		},
		W: 36,
	},
	'G': Glyph{
		S: Set{
			Path{
				Point{18, 37},
				Point{14, 35},
				Point{10, 31},
				Point{8, 27},
				Point{8, 19},
				Point{10, 15},
				Point{14, 11},
				Point{18, 9},
				Point{24, 7},
				Point{34, 7},
				Point{40, 9},
				Point{44, 11},
				Point{48, 15},
				Point{50, 19},
				Point{50, 27},
				Point{48, 31},
				Point{44, 35},
				Point{40, 37},
				Point{34, 37},
			},
			Path{
				Point{34, 27},
				Point{34, 37},
			},
		},
		W: 42,
	},
	'H': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{50, 8},
			},
			Path{
				Point{8, 36},
				Point{50, 36},
			},
			Path{
				Point{28, 8},
				Point{28, 36},
			},
		},
		W: 44,
	},
	'I': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{50, 8},
			},
		},
		W: 16,
	},
	'J': Glyph{
		S: Set{
			Path{
				Point{8, 24},
				Point{40, 24},
				Point{46, 22},
				Point{48, 20},
				Point{50, 16},
				Point{50, 12},
				Point{48, 8},
				Point{46, 6},
				Point{40, 4},
				Point{36, 4},
			},
		},
		W: 32,
	},
	'K': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 35},
				Point{36, 7},
			},
			Path{
				Point{26, 17},
				Point{50, 35},
			},
		},
		W: 42,
	},
	'L': Glyph{
		S: Set{
			Path{
				Point{8, 5},
				Point{50, 5},
			},
			Path{
				Point{50, 5},
				Point{50, 29},
			},
		},
		W: 34,
	},
	'M': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{50, 8},
			},
			Path{
				Point{8, 8},
				Point{50, 24},
			},
			Path{
				Point{8, 40},
				Point{50, 24},
			},
			Path{
				Point{8, 40},
				Point{50, 40},
			},
		},
		W: 48,
	},
	'N': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{50, 8},
			},
			Path{
				Point{8, 8},
				Point{50, 36},
			},
			Path{
				Point{8, 36},
				Point{50, 36},
			},
		},
		W: 44,
	},
	'O': Glyph{
		S: Set{
			Path{
				Point{8, 18},
				Point{10, 14},
				Point{14, 10},
				Point{18, 8},
				Point{24, 6},
				Point{34, 6},
				Point{40, 8},
				Point{44, 10},
				Point{48, 14},
				Point{50, 18},
				Point{50, 26},
				Point{48, 30},
				Point{44, 34},
				Point{40, 36},
				Point{34, 38},
				Point{24, 38},
				Point{18, 36},
				Point{14, 34},
				Point{10, 30},
				Point{8, 26},
				Point{8, 18},
			},
		},
		W: 44,
	},
	'P': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 7},
				Point{8, 25},
				Point{10, 31},
				Point{12, 33},
				Point{16, 35},
				Point{22, 35},
				Point{26, 33},
				Point{28, 31},
				Point{30, 25},
				Point{30, 7},
			},
		},
		W: 42,
	},
	'Q': Glyph{
		S: Set{
			Path{
				Point{8, 18},
				Point{10, 14},
				Point{14, 10},
				Point{18, 8},
				Point{24, 6},
				Point{34, 6},
				Point{40, 8},
				Point{44, 10},
				Point{48, 14},
				Point{50, 18},
				Point{50, 26},
				Point{48, 30},
				Point{44, 34},
				Point{40, 36},
				Point{34, 38},
				Point{24, 38},
				Point{18, 36},
				Point{14, 34},
				Point{10, 30},
				Point{8, 26},
				Point{8, 18},
			},
			Path{
				Point{42, 24},
				Point{54, 36},
			},
		},
		W: 44,
	},
	'R': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{8, 7},
				Point{8, 25},
				Point{10, 31},
				Point{12, 33},
				Point{16, 35},
				Point{20, 35},
				Point{24, 33},
				Point{26, 31},
				Point{28, 25},
				Point{28, 7},
			},
			Path{
				Point{28, 21},
				Point{50, 35},
			},
		},
		W: 42,
	},
	'S': Glyph{
		S: Set{
			Path{
				Point{14, 34},
				Point{10, 30},
				Point{8, 24},
				Point{8, 16},
				Point{10, 10},
				Point{14, 6},
				Point{18, 6},
				Point{22, 8},
				Point{24, 10},
				Point{26, 14},
				Point{30, 26},
				Point{32, 30},
				Point{34, 32},
				Point{38, 34},
				Point{44, 34},
				Point{48, 30},
				Point{50, 24},
				Point{50, 16},
				Point{48, 10},
				Point{44, 6},
			},
		},
		W: 40,
	},
	'T': Glyph{
		S: Set{
			Path{
				Point{8, 16},
				Point{50, 16},
			},
			Path{
				Point{8, 2},
				Point{8, 30},
			},
		},
		W: 32,
	},
	'U': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{38, 8},
				Point{44, 10},
				Point{48, 14},
				Point{50, 20},
				Point{50, 24},
				Point{48, 30},
				Point{44, 34},
				Point{38, 36},
				Point{8, 36},
			},
		},
		W: 44,
	},
	'V': Glyph{
		S: Set{
			Path{
				Point{8, 2},
				Point{50, 18},
			},
			Path{
				Point{8, 34},
				Point{50, 18},
			},
		},
		W: 36,
	},
	'W': Glyph{
		S: Set{
			Path{
				Point{8, 4},
				Point{50, 14},
			},
			Path{
				Point{8, 24},
				Point{50, 14},
			},
			Path{
				Point{8, 24},
				Point{50, 34},
			},
			Path{
				Point{8, 44},
				Point{50, 34},
			},
		},
		W: 48,
	},
	'X': Glyph{
		S: Set{
			Path{
				Point{8, 6},
				Point{50, 34},
			},
			Path{
				Point{8, 34},
				Point{50, 6},
			},
		},
		W: 40,
	},
	'Y': Glyph{
		S: Set{
			Path{
				Point{8, 2},
				Point{28, 18},
				Point{50, 18},
			},
			Path{
				Point{8, 34},
				Point{28, 18},
			},
		},
		W: 36,
	},
	'Z': Glyph{
		S: Set{
			Path{
				Point{8, 34},
				Point{50, 6},
			},
			Path{
				Point{8, 6},
				Point{8, 34},
			},
			Path{
				Point{50, 6},
				Point{50, 34},
			},
		},
		W: 40,
	},
	'[': Glyph{
		S: Set{
			Path{
				Point{0, 8},
				Point{64, 8},
			},
			Path{
				Point{0, 10},
				Point{64, 10},
			},
			Path{
				Point{0, 8},
				Point{0, 22},
			},
			Path{
				Point{64, 8},
				Point{64, 22},
			},
		},
		W: 28,
	},
	'\\': Glyph{
		S: Set{
			Path{
				Point{8, 0},
				Point{56, 28},
			},
		},
		W: 28,
	},
	']': Glyph{
		S: Set{
			Path{
				Point{0, 18},
				Point{64, 18},
			},
			Path{
				Point{0, 20},
				Point{64, 20},
			},
			Path{
				Point{0, 6},
				Point{0, 20},
			},
			Path{
				Point{64, 6},
				Point{64, 20},
			},
		},
		W: 28,
	},
	'^': Glyph{
		S: Set{
			Path{
				Point{20, 12},
				Point{14, 16},
				Point{20, 20},
			},
			Path{
				Point{26, 6},
				Point{16, 16},
				Point{26, 26},
			},
			Path{
				Point{16, 16},
				Point{50, 16},
			},
		},
		W: 32,
	},
	'_': Glyph{
		S: Set{
			Path{
				Point{54, 0},
				Point{54, 32},
			},
		},
		W: 32,
	},
	'`': Glyph{
		S: Set{
			Path{
				Point{8, 12},
				Point{10, 10},
				Point{14, 8},
				Point{18, 8},
				Point{20, 10},
				Point{18, 12},
				Point{16, 10},
			},
		},
		W: 20,
	},
	'a': Glyph{
		S: Set{
			Path{
				Point{22, 31},
				Point{50, 31},
			},
			Path{
				Point{28, 31},
				Point{24, 27},
				Point{22, 23},
				Point{22, 17},
				Point{24, 13},
				Point{28, 9},
				Point{34, 7},
				Point{38, 7},
				Point{44, 9},
				Point{48, 13},
				Point{50, 17},
				Point{50, 23},
				Point{48, 27},
				Point{44, 31},
			},
		},
		W: 38,
	},
	'b': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{28, 7},
				Point{24, 11},
				Point{22, 15},
				Point{22, 21},
				Point{24, 25},
				Point{28, 29},
				Point{34, 31},
				Point{38, 31},
				Point{44, 29},
				Point{48, 25},
				Point{50, 21},
				Point{50, 15},
				Point{48, 11},
				Point{44, 7},
			},
		},
		W: 38,
	},
	'c': Glyph{
		S: Set{
			Path{
				Point{28, 30},
				Point{24, 26},
				Point{22, 22},
				Point{22, 16},
				Point{24, 12},
				Point{28, 8},
				Point{34, 6},
				Point{38, 6},
				Point{44, 8},
				Point{48, 12},
				Point{50, 16},
				Point{50, 22},
				Point{48, 26},
				Point{44, 30},
			},
		},
		W: 36,
	},
	'd': Glyph{
		S: Set{
			Path{
				Point{8, 31},
				Point{50, 31},
			},
			Path{
				Point{28, 31},
				Point{24, 27},
				Point{22, 23},
				Point{22, 17},
				Point{24, 13},
				Point{28, 9},
				Point{34, 7},
				Point{38, 7},
				Point{44, 9},
				Point{48, 13},
				Point{50, 17},
				Point{50, 23},
				Point{48, 27},
				Point{44, 31},
			},
		},
		W: 38,
	},
	'e': Glyph{
		S: Set{
			Path{
				Point{34, 6},
				Point{34, 30},
				Point{30, 30},
				Point{26, 28},
				Point{24, 26},
				Point{22, 22},
				Point{22, 16},
				Point{24, 12},
				Point{28, 8},
				Point{34, 6},
				Point{38, 6},
				Point{44, 8},
				Point{48, 12},
				Point{50, 16},
				Point{50, 22},
				Point{48, 26},
				Point{44, 30},
			},
		},
		W: 36,
	},
	'f': Glyph{
		S: Set{
			Path{
				Point{8, 22},
				Point{8, 18},
				Point{10, 14},
				Point{16, 12},
				Point{50, 12},
			},
			Path{
				Point{22, 6},
				Point{22, 20},
			},
		},
		W: 24,
	},
	'g': Glyph{
		S: Set{
			Path{
				Point{22, 31},
				Point{54, 31},
				Point{60, 29},
				Point{62, 27},
				Point{64, 23},
				Point{64, 17},
				Point{62, 13},
			},
			Path{
				Point{28, 31},
				Point{24, 27},
				Point{22, 23},
				Point{22, 17},
				Point{24, 13},
				Point{28, 9},
				Point{34, 7},
				Point{38, 7},
				Point{44, 9},
				Point{48, 13},
				Point{50, 17},
				Point{50, 23},
				Point{48, 27},
				Point{44, 31},
			},
		},
		W: 38,
	},
	'h': Glyph{
		S: Set{
			Path{
				Point{8, 9},
				Point{50, 9},
			},
			Path{
				Point{30, 9},
				Point{24, 15},
				Point{22, 19},
				Point{22, 25},
				Point{24, 29},
				Point{30, 31},
				Point{50, 31},
			},
		},
		W: 38,
	},
	'i': Glyph{
		S: Set{
			Path{
				Point{8, 6},
				Point{10, 8},
				Point{8, 10},
				Point{6, 8},
				Point{8, 6},
			},
			Path{
				Point{22, 8},
				Point{50, 8},
			},
		},
		W: 16,
	},
	'j': Glyph{
		S: Set{
			Path{
				Point{8, 10},
				Point{10, 12},
				Point{8, 14},
				Point{6, 12},
				Point{8, 10},
			},
			Path{
				Point{22, 12},
				Point{56, 12},
				Point{62, 10},
				Point{64, 6},
				Point{64, 2},
			},
		},
		W: 20,
	},
	'k': Glyph{
		S: Set{
			Path{
				Point{8, 7},
				Point{50, 7},
			},
			Path{
				Point{22, 27},
				Point{42, 7},
			},
			Path{
				Point{34, 15},
				Point{50, 29},
			},
		},
		W: 34,
	},
	'l': Glyph{
		S: Set{
			Path{
				Point{8, 8},
				Point{50, 8},
			},
		},
		W: 16,
	},
	'm': Glyph{
		S: Set{
			Path{
				Point{22, 8},
				Point{50, 8},
			},
			Path{
				Point{30, 8},
				Point{24, 14},
				Point{22, 18},
				Point{22, 24},
				Point{24, 28},
				Point{30, 30},
				Point{50, 30},
			},
			Path{
				Point{30, 30},
				Point{24, 36},
				Point{22, 40},
				Point{22, 46},
				Point{24, 50},
				Point{30, 52},
				Point{50, 52},
			},
		},
		W: 60,
	},
	'n': Glyph{
		S: Set{
			Path{
				Point{22, 9},
				Point{50, 9},
			},
			Path{
				Point{30, 9},
				Point{24, 15},
				Point{22, 19},
				Point{22, 25},
				Point{24, 29},
				Point{30, 31},
				Point{50, 31},
			},
		},
		W: 38,
	},
	'o': Glyph{
		S: Set{
			Path{
				Point{22, 17},
				Point{24, 13},
				Point{28, 9},
				Point{34, 7},
				Point{38, 7},
				Point{44, 9},
				Point{48, 13},
				Point{50, 17},
				Point{50, 23},
				Point{48, 27},
				Point{44, 31},
				Point{38, 33},
				Point{34, 33},
				Point{28, 31},
				Point{24, 27},
				Point{22, 23},
				Point{22, 17},
			},
		},
		W: 38,
	},
	'p': Glyph{
		S: Set{
			Path{
				Point{22, 7},
				Point{64, 7},
			},
			Path{
				Point{28, 7},
				Point{24, 11},
				Point{22, 15},
				Point{22, 21},
				Point{24, 25},
				Point{28, 29},
				Point{34, 31},
				Point{38, 31},
				Point{44, 29},
				Point{48, 25},
				Point{50, 21},
				Point{50, 15},
				Point{48, 11},
				Point{44, 7},
			},
		},
		W: 38,
	},
	'q': Glyph{
		S: Set{
			Path{
				Point{22, 31},
				Point{64, 31},
			},
			Path{
				Point{28, 31},
				Point{24, 27},
				Point{22, 23},
				Point{22, 17},
				Point{24, 13},
				Point{28, 9},
				Point{34, 7},
				Point{38, 7},
				Point{44, 9},
				Point{48, 13},
				Point{50, 17},
				Point{50, 23},
				Point{48, 27},
				Point{44, 31},
			},
		},
		W: 38,
	},
	'r': Glyph{
		S: Set{
			Path{
				Point{22, 7},
				Point{50, 7},
			},
			Path{
				Point{34, 7},
				Point{28, 9},
				Point{24, 13},
				Point{22, 17},
				Point{22, 23},
			},
		},
		W: 26,
	},
	's': Glyph{
		S: Set{
			Path{
				Point{28, 29},
				Point{24, 27},
				Point{22, 21},
				Point{22, 15},
				Point{24, 9},
				Point{28, 7},
				Point{32, 9},
				Point{34, 13},
				Point{36, 23},
				Point{38, 27},
				Point{42, 29},
				Point{44, 29},
				Point{48, 27},
				Point{50, 21},
				Point{50, 15},
				Point{48, 9},
				Point{44, 7},
			},
		},
		W: 34,
	},
	't': Glyph{
		S: Set{
			Path{
				Point{8, 12},
				Point{42, 12},
				Point{48, 14},
				Point{50, 18},
				Point{50, 22},
			},
			Path{
				Point{22, 6},
				Point{22, 20},
			},
		},
		W: 24,
	},
	'u': Glyph{
		S: Set{
			Path{
				Point{22, 9},
				Point{42, 9},
				Point{48, 11},
				Point{50, 15},
				Point{50, 21},
				Point{48, 25},
				Point{42, 31},
			},
			Path{
				Point{22, 31},
				Point{50, 31},
			},
		},
		W: 38,
	},
	'v': Glyph{
		S: Set{
			Path{
				Point{22, 4},
				Point{50, 16},
			},
			Path{
				Point{22, 28},
				Point{50, 16},
			},
		},
		W: 32,
	},
	'w': Glyph{
		S: Set{
			Path{
				Point{22, 6},
				Point{50, 14},
			},
			Path{
				Point{22, 22},
				Point{50, 14},
			},
			Path{
				Point{22, 22},
				Point{50, 30},
			},
			Path{
				Point{22, 38},
				Point{50, 30},
			},
		},
		W: 44,
	},
	'x': Glyph{
		S: Set{
			Path{
				Point{22, 7},
				Point{50, 29},
			},
			Path{
				Point{22, 29},
				Point{50, 7},
			},
		},
		W: 34,
	},
	'y': Glyph{
		S: Set{
			Path{
				Point{22, 4},
				Point{50, 16},
			},
			Path{
				Point{22, 28},
				Point{50, 16},
				Point{58, 12},
				Point{62, 8},
				Point{64, 4},
				Point{64, 2},
			},
		},
		W: 32,
	},
	'z': Glyph{
		S: Set{
			Path{
				Point{22, 29},
				Point{50, 7},
			},
			Path{
				Point{22, 7},
				Point{22, 29},
			},
			Path{
				Point{50, 7},
				Point{50, 29},
			},
		},
		W: 34,
	},
	'{': Glyph{
		S: Set{
			Path{
				Point{0, 18},
				Point{2, 14},
				Point{4, 12},
				Point{8, 10},
				Point{12, 10},
				Point{16, 12},
				Point{18, 14},
				Point{22, 16},
				Point{26, 16},
				Point{30, 12},
			},
			Path{
				Point{2, 14},
				Point{6, 12},
				Point{10, 12},
				Point{14, 14},
				Point{16, 16},
				Point{20, 18},
				Point{24, 18},
				Point{28, 16},
				Point{32, 8},
				Point{36, 16},
				Point{40, 18},
				Point{44, 18},
				Point{48, 16},
				Point{50, 14},
				Point{54, 12},
				Point{58, 12},
				Point{62, 14},
			},
			Path{
				Point{34, 12},
				Point{38, 16},
				Point{42, 16},
				Point{46, 14},
				Point{48, 12},
				Point{52, 10},
				Point{56, 10},
				Point{60, 12},
				Point{62, 14},
				Point{64, 18},
			},
		},
		W: 28,
	},
	'|': Glyph{
		S: Set{
			Path{
				Point{0, 8},
				Point{64, 8},
			},
		},
		W: 16,
	},
	'}': Glyph{
		S: Set{
			Path{
				Point{0, 10},
				Point{2, 14},
				Point{4, 16},
				Point{8, 18},
				Point{12, 18},
				Point{16, 16},
				Point{18, 14},
				Point{22, 12},
				Point{26, 12},
				Point{30, 16},
			},
			Path{
				Point{2, 14},
				Point{6, 16},
				Point{10, 16},
				Point{14, 14},
				Point{16, 12},
				Point{20, 10},
				Point{24, 10},
				Point{28, 12},
				Point{32, 20},
				Point{36, 12},
				Point{40, 10},
				Point{44, 10},
				Point{48, 12},
				Point{50, 14},
				Point{54, 16},
				Point{58, 16},
				Point{62, 14},
			},
			Path{
				Point{34, 16},
				Point{38, 12},
				Point{42, 12},
				Point{46, 14},
				Point{48, 16},
				Point{52, 18},
				Point{56, 18},
				Point{60, 16},
				Point{62, 14},
				Point{64, 10},
			},
		},
		W: 28,
	},
	'~': Glyph{
		S: Set{
			Path{
				Point{38, 6},
				Point{34, 6},
				Point{28, 8},
				Point{26, 12},
				Point{26, 16},
				Point{28, 20},
				Point{34, 28},
				Point{36, 32},
				Point{36, 36},
				Point{34, 40},
				Point{30, 42},
			},
			Path{
				Point{34, 6},
				Point{30, 8},
				Point{28, 12},
				Point{28, 16},
				Point{30, 20},
				Point{36, 28},
				Point{38, 32},
				Point{38, 36},
				Point{36, 40},
				Point{30, 42},
				Point{26, 42},
			},
		},
		W: 48,
	},
	'\u007f': Glyph{
		S: Set{
			Path{
				Point{8, 12},
				Point{10, 8},
				Point{14, 6},
				Point{18, 6},
				Point{22, 8},
				Point{24, 12},
				Point{24, 16},
				Point{22, 20},
				Point{18, 22},
				Point{14, 22},
				Point{10, 20},
				Point{8, 16},
				Point{8, 12},
			},
		},
		W: 28,
	},
}
