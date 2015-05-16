package robo

import (
	"bufio"
	"log"
	"os"
)

type Font map[rune]Glyph

type Glyph struct {
	S Set
	W Unit
}

type Set []Path

func PrintStdin(c *bufio.Writer) {
	var off Point

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		off = font.Print(c, scanner.Text(), off)
		off.X += 96
		off.Y = 0
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (f Font) Print(c *bufio.Writer, s string, off Point) Point {
	for _, ch := range s {
		gl, ok := f[ch]
		if ok {
			if off.Y+gl.W >= 4000 {
				off.X += 96
				off.Y = 0
			}
			off.Offset(c)
			for _, p := range gl.S {
				p.Line(c)
			}
			off.Y += gl.W
		} else if ch == '\t' {
			off.Y += f[' '].W * 8
		}
	}
	return off
}

var font = Font{
	' ': Glyph{
		S: Set{},
		W: Unit(43.49248),
	},
	'!': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(48.93), Unit(13.59)},
			},
			Path{
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(67.96), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(62.52), Unit(13.59)},
			},
		},
		W: Unit(27.1828),
	},
	'"': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(29.90), Unit(10.87)},
			},
			Path{
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(29.90), Unit(32.62)},
			},
		},
		W: Unit(43.49248),
	},
	'#': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(31.26)},
				Point{Unit(86.98), Unit(12.23)},
			},
			Path{
				Point{Unit(0.00), Unit(47.57)},
				Point{Unit(86.98), Unit(28.54)},
			},
			Path{
				Point{Unit(35.34), Unit(12.23)},
				Point{Unit(35.34), Unit(50.29)},
			},
			Path{
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(51.65), Unit(47.57)},
			},
		},
		W: Unit(57.08388),
	},
	'$': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(21.75)},
				Point{Unit(78.83), Unit(21.75)},
			},
			Path{
				Point{Unit(0.00), Unit(32.62)},
				Point{Unit(78.83), Unit(32.62)},
			},
			Path{
				Point{Unit(19.03), Unit(46.21)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(13.59), Unit(13.59)},
				Point{Unit(19.03), Unit(8.15)},
				Point{Unit(24.46), Unit(8.15)},
				Point{Unit(29.90), Unit(10.87)},
				Point{Unit(32.62), Unit(13.59)},
				Point{Unit(35.34), Unit(19.03)},
				Point{Unit(40.77), Unit(35.34)},
				Point{Unit(43.49), Unit(40.77)},
				Point{Unit(46.21), Unit(43.49)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(59.80), Unit(46.21)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(67.96), Unit(32.62)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(59.80), Unit(8.15)},
			},
		},
		W: Unit(54.3656),
	},
	'%': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(57.08)},
				Point{Unit(67.96), Unit(8.15)},
			},
			Path{
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(16.31), Unit(27.18)},
				Point{Unit(21.75), Unit(27.18)},
				Point{Unit(27.18), Unit(24.46)},
				Point{Unit(29.90), Unit(19.03)},
				Point{Unit(29.90), Unit(13.59)},
				Point{Unit(24.46), Unit(8.15)},
				Point{Unit(19.03), Unit(8.15)},
				Point{Unit(13.59), Unit(10.87)},
				Point{Unit(10.87), Unit(16.31)},
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(13.59), Unit(27.18)},
				Point{Unit(16.31), Unit(35.34)},
				Point{Unit(16.31), Unit(43.49)},
				Point{Unit(13.59), Unit(51.65)},
				Point{Unit(10.87), Unit(57.08)},
			},
			Path{
				Point{Unit(48.93), Unit(46.21)},
				Point{Unit(51.65), Unit(40.77)},
				Point{Unit(57.08), Unit(38.06)},
				Point{Unit(62.52), Unit(38.06)},
				Point{Unit(67.96), Unit(43.49)},
				Point{Unit(67.96), Unit(48.93)},
				Point{Unit(65.24), Unit(54.37)},
				Point{Unit(59.80), Unit(57.08)},
				Point{Unit(54.37), Unit(57.08)},
				Point{Unit(48.93), Unit(51.65)},
				Point{Unit(48.93), Unit(46.21)},
			},
		},
		W: Unit(65.23872),
	},
	'&': Glyph{
		S: Set{
			Path{
				Point{Unit(35.34), Unit(62.52)},
				Point{Unit(32.62), Unit(62.52)},
				Point{Unit(29.90), Unit(59.80)},
				Point{Unit(29.90), Unit(57.08)},
				Point{Unit(32.62), Unit(54.37)},
				Point{Unit(38.06), Unit(51.65)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(59.80), Unit(40.77)},
				Point{Unit(65.24), Unit(35.34)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(67.96), Unit(19.03)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(62.52), Unit(10.87)},
				Point{Unit(57.08), Unit(8.15)},
				Point{Unit(51.65), Unit(8.15)},
				Point{Unit(46.21), Unit(10.87)},
				Point{Unit(43.49), Unit(13.59)},
				Point{Unit(32.62), Unit(32.62)},
				Point{Unit(29.90), Unit(35.34)},
				Point{Unit(24.46), Unit(38.06)},
				Point{Unit(19.03), Unit(38.06)},
				Point{Unit(13.59), Unit(35.34)},
				Point{Unit(10.87), Unit(29.90)},
				Point{Unit(13.59), Unit(24.46)},
				Point{Unit(19.03), Unit(21.75)},
				Point{Unit(24.46), Unit(21.75)},
				Point{Unit(32.62), Unit(24.46)},
				Point{Unit(40.77), Unit(29.90)},
				Point{Unit(59.80), Unit(43.49)},
				Point{Unit(65.24), Unit(48.93)},
				Point{Unit(67.96), Unit(54.37)},
				Point{Unit(67.96), Unit(59.80)},
				Point{Unit(65.24), Unit(62.52)},
				Point{Unit(62.52), Unit(62.52)},
			},
		},
		W: Unit(70.67528),
	},
	'\'': Glyph{
		S: Set{
			Path{
				Point{Unit(16.31), Unit(13.59)},
				Point{Unit(13.59), Unit(10.87)},
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(19.03), Unit(16.31)},
				Point{Unit(24.46), Unit(13.59)},
				Point{Unit(27.18), Unit(10.87)},
			},
		},
		W: Unit(27.1828),
	},
	'(': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(29.90)},
				Point{Unit(5.44), Unit(24.46)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(24.46), Unit(13.59)},
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(48.93), Unit(10.87)},
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(73.39), Unit(19.03)},
				Point{Unit(81.55), Unit(24.46)},
				Point{Unit(86.98), Unit(29.90)},
			},
		},
		W: Unit(38.05592),
	},
	')': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(8.15)},
				Point{Unit(5.44), Unit(13.59)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(24.46), Unit(24.46)},
				Point{Unit(38.06), Unit(27.18)},
				Point{Unit(48.93), Unit(27.18)},
				Point{Unit(62.52), Unit(24.46)},
				Point{Unit(73.39), Unit(19.03)},
				Point{Unit(81.55), Unit(13.59)},
				Point{Unit(86.98), Unit(8.15)},
			},
		},
		W: Unit(38.05592),
	},
	'*': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(43.49), Unit(21.75)},
			},
			Path{
				Point{Unit(19.03), Unit(8.15)},
				Point{Unit(35.34), Unit(35.34)},
			},
			Path{
				Point{Unit(19.03), Unit(35.34)},
				Point{Unit(35.34), Unit(8.15)},
			},
		},
		W: Unit(43.49248),
	},
	'+': Glyph{
		S: Set{
			Path{
				Point{Unit(19.03), Unit(35.34)},
				Point{Unit(67.96), Unit(35.34)},
			},
			Path{
				Point{Unit(43.49), Unit(10.87)},
				Point{Unit(43.49), Unit(59.80)},
			},
		},
		W: Unit(70.67528),
	},
	',': Glyph{
		S: Set{
			Path{
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(13.59)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(70.68), Unit(16.31)},
				Point{Unit(76.11), Unit(13.59)},
				Point{Unit(78.83), Unit(10.87)},
			},
		},
		W: Unit(27.1828),
	},
	'-': Glyph{
		S: Set{
			Path{
				Point{Unit(43.49), Unit(10.87)},
				Point{Unit(43.49), Unit(59.80)},
			},
		},
		W: Unit(70.67528),
	},
	'.': Glyph{
		S: Set{
			Path{
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(67.96), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(62.52), Unit(13.59)},
			},
		},
		W: Unit(27.1828),
	},
	'/': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(54.37)},
				Point{Unit(86.98), Unit(5.44)},
			},
		},
		W: Unit(59.80216),
	},
	'0': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(21.75), Unit(10.87)},
				Point{Unit(35.34), Unit(8.15)},
				Point{Unit(43.49), Unit(8.15)},
				Point{Unit(57.08), Unit(10.87)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(24.46)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(65.24), Unit(38.06)},
				Point{Unit(57.08), Unit(43.49)},
				Point{Unit(43.49), Unit(46.21)},
				Point{Unit(35.34), Unit(46.21)},
				Point{Unit(21.75), Unit(43.49)},
				Point{Unit(13.59), Unit(38.06)},
				Point{Unit(10.87), Unit(29.90)},
				Point{Unit(10.87), Unit(24.46)},
			},
		},
		W: Unit(54.3656),
	},
	'1': Glyph{
		S: Set{
			Path{
				Point{Unit(21.75), Unit(16.31)},
				Point{Unit(19.03), Unit(21.75)},
				Point{Unit(10.87), Unit(29.90)},
				Point{Unit(67.96), Unit(29.90)},
			},
		},
		W: Unit(54.3656),
	},
	'2': Glyph{
		S: Set{
			Path{
				Point{Unit(24.46), Unit(10.87)},
				Point{Unit(21.75), Unit(10.87)},
				Point{Unit(16.31), Unit(13.59)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(13.59), Unit(38.06)},
				Point{Unit(16.31), Unit(40.77)},
				Point{Unit(21.75), Unit(43.49)},
				Point{Unit(27.18), Unit(43.49)},
				Point{Unit(32.62), Unit(40.77)},
				Point{Unit(40.77), Unit(35.34)},
				Point{Unit(67.96), Unit(8.15)},
				Point{Unit(67.96), Unit(46.21)},
			},
		},
		W: Unit(54.3656),
	},
	'3': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(10.87), Unit(43.49)},
				Point{Unit(32.62), Unit(27.18)},
				Point{Unit(32.62), Unit(35.34)},
				Point{Unit(35.34), Unit(40.77)},
				Point{Unit(38.06), Unit(43.49)},
				Point{Unit(46.21), Unit(46.21)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(59.80), Unit(43.49)},
				Point{Unit(65.24), Unit(38.06)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(62.52), Unit(10.87)},
				Point{Unit(57.08), Unit(8.15)},
			},
		},
		W: Unit(54.3656),
	},
	'4': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(35.34)},
				Point{Unit(48.93), Unit(8.15)},
				Point{Unit(48.93), Unit(48.93)},
			},
			Path{
				Point{Unit(10.87), Unit(35.34)},
				Point{Unit(67.96), Unit(35.34)},
			},
		},
		W: Unit(54.3656),
	},
	'5': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(40.77)},
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(35.34), Unit(10.87)},
				Point{Unit(32.62), Unit(13.59)},
				Point{Unit(29.90), Unit(21.75)},
				Point{Unit(29.90), Unit(29.90)},
				Point{Unit(32.62), Unit(38.06)},
				Point{Unit(38.06), Unit(43.49)},
				Point{Unit(46.21), Unit(46.21)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(59.80), Unit(43.49)},
				Point{Unit(65.24), Unit(38.06)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(62.52), Unit(10.87)},
				Point{Unit(57.08), Unit(8.15)},
			},
		},
		W: Unit(54.3656),
	},
	'6': Glyph{
		S: Set{
			Path{
				Point{Unit(19.03), Unit(43.49)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(10.87), Unit(27.18)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(21.75), Unit(13.59)},
				Point{Unit(35.34), Unit(10.87)},
				Point{Unit(48.93), Unit(10.87)},
				Point{Unit(59.80), Unit(13.59)},
				Point{Unit(65.24), Unit(19.03)},
				Point{Unit(67.96), Unit(27.18)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(65.24), Unit(38.06)},
				Point{Unit(59.80), Unit(43.49)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(48.93), Unit(46.21)},
				Point{Unit(40.77), Unit(43.49)},
				Point{Unit(35.34), Unit(38.06)},
				Point{Unit(32.62), Unit(29.90)},
				Point{Unit(32.62), Unit(27.18)},
				Point{Unit(35.34), Unit(19.03)},
				Point{Unit(40.77), Unit(13.59)},
				Point{Unit(48.93), Unit(10.87)},
			},
		},
		W: Unit(54.3656),
	},
	'7': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(46.21)},
				Point{Unit(67.96), Unit(19.03)},
			},
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(10.87), Unit(46.21)},
			},
		},
		W: Unit(54.3656),
	},
	'8': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(13.59), Unit(13.59)},
				Point{Unit(19.03), Unit(10.87)},
				Point{Unit(24.46), Unit(10.87)},
				Point{Unit(29.90), Unit(13.59)},
				Point{Unit(32.62), Unit(19.03)},
				Point{Unit(35.34), Unit(29.90)},
				Point{Unit(38.06), Unit(38.06)},
				Point{Unit(43.49), Unit(43.49)},
				Point{Unit(48.93), Unit(46.21)},
				Point{Unit(57.08), Unit(46.21)},
				Point{Unit(62.52), Unit(43.49)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(67.96), Unit(32.62)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(62.52), Unit(10.87)},
				Point{Unit(57.08), Unit(8.15)},
				Point{Unit(48.93), Unit(8.15)},
				Point{Unit(43.49), Unit(10.87)},
				Point{Unit(38.06), Unit(16.31)},
				Point{Unit(35.34), Unit(24.46)},
				Point{Unit(32.62), Unit(35.34)},
				Point{Unit(29.90), Unit(40.77)},
				Point{Unit(24.46), Unit(43.49)},
				Point{Unit(19.03), Unit(43.49)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(10.87), Unit(21.75)},
			},
		},
		W: Unit(54.3656),
	},
	'9': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(43.49)},
				Point{Unit(38.06), Unit(40.77)},
				Point{Unit(43.49), Unit(35.34)},
				Point{Unit(46.21), Unit(27.18)},
				Point{Unit(46.21), Unit(24.46)},
				Point{Unit(43.49), Unit(16.31)},
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(29.90), Unit(8.15)},
				Point{Unit(27.18), Unit(8.15)},
				Point{Unit(19.03), Unit(10.87)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(10.87), Unit(27.18)},
				Point{Unit(13.59), Unit(35.34)},
				Point{Unit(19.03), Unit(40.77)},
				Point{Unit(29.90), Unit(43.49)},
				Point{Unit(43.49), Unit(43.49)},
				Point{Unit(57.08), Unit(40.77)},
				Point{Unit(65.24), Unit(35.34)},
				Point{Unit(67.96), Unit(27.18)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(59.80), Unit(10.87)},
			},
		},
		W: Unit(54.3656),
	},
	':': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(13.59)},
				Point{Unit(32.62), Unit(10.87)},
				Point{Unit(35.34), Unit(13.59)},
				Point{Unit(32.62), Unit(16.31)},
				Point{Unit(29.90), Unit(13.59)},
			},
			Path{
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(67.96), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(62.52), Unit(13.59)},
			},
		},
		W: Unit(27.1828),
	},
	';': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(13.59)},
				Point{Unit(32.62), Unit(10.87)},
				Point{Unit(35.34), Unit(13.59)},
				Point{Unit(32.62), Unit(16.31)},
				Point{Unit(29.90), Unit(13.59)},
			},
			Path{
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(13.59)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(62.52), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(70.68), Unit(16.31)},
				Point{Unit(76.11), Unit(13.59)},
				Point{Unit(78.83), Unit(10.87)},
			},
		},
		W: Unit(27.1828),
	},
	'<': Glyph{
		S: Set{
			Path{
				Point{Unit(19.03), Unit(54.37)},
				Point{Unit(43.49), Unit(10.87)},
				Point{Unit(67.96), Unit(54.37)},
			},
		},
		W: Unit(65.23872),
	},
	'=': Glyph{
		S: Set{
			Path{
				Point{Unit(35.34), Unit(10.87)},
				Point{Unit(35.34), Unit(59.80)},
			},
			Path{
				Point{Unit(51.65), Unit(10.87)},
				Point{Unit(51.65), Unit(59.80)},
			},
		},
		W: Unit(70.67528),
	},
	'>': Glyph{
		S: Set{
			Path{
				Point{Unit(19.03), Unit(10.87)},
				Point{Unit(43.49), Unit(54.37)},
				Point{Unit(67.96), Unit(10.87)},
			},
		},
		W: Unit(65.23872),
	},
	'?': Glyph{
		S: Set{
			Path{
				Point{Unit(24.46), Unit(8.15)},
				Point{Unit(21.75), Unit(8.15)},
				Point{Unit(16.31), Unit(10.87)},
				Point{Unit(13.59), Unit(13.59)},
				Point{Unit(10.87), Unit(19.03)},
				Point{Unit(10.87), Unit(29.90)},
				Point{Unit(13.59), Unit(35.34)},
				Point{Unit(16.31), Unit(38.06)},
				Point{Unit(21.75), Unit(40.77)},
				Point{Unit(27.18), Unit(40.77)},
				Point{Unit(32.62), Unit(38.06)},
				Point{Unit(35.34), Unit(35.34)},
				Point{Unit(40.77), Unit(24.46)},
				Point{Unit(48.93), Unit(24.46)},
			},
			Path{
				Point{Unit(62.52), Unit(24.46)},
				Point{Unit(65.24), Unit(21.75)},
				Point{Unit(67.96), Unit(24.46)},
				Point{Unit(65.24), Unit(27.18)},
				Point{Unit(62.52), Unit(24.46)},
			},
		},
		W: Unit(48.92904),
	},
	'@': Glyph{
		S: Set{
			Path{
				Point{Unit(32.62), Unit(50.29)},
				Point{Unit(27.18), Unit(47.57)},
				Point{Unit(24.46), Unit(42.13)},
				Point{Unit(24.46), Unit(33.98)},
				Point{Unit(27.18), Unit(28.54)},
				Point{Unit(29.90), Unit(25.82)},
				Point{Unit(38.06), Unit(23.11)},
				Point{Unit(46.21), Unit(23.11)},
				Point{Unit(51.65), Unit(25.82)},
				Point{Unit(54.37), Unit(31.26)},
				Point{Unit(54.37), Unit(39.42)},
				Point{Unit(51.65), Unit(44.85)},
				Point{Unit(46.21), Unit(47.57)},
			},
			Path{
				Point{Unit(24.46), Unit(33.98)},
				Point{Unit(29.90), Unit(28.54)},
				Point{Unit(38.06), Unit(25.82)},
				Point{Unit(46.21), Unit(25.82)},
				Point{Unit(51.65), Unit(28.54)},
				Point{Unit(54.37), Unit(31.26)},
			},
			Path{
				Point{Unit(24.46), Unit(50.29)},
				Point{Unit(46.21), Unit(47.57)},
				Point{Unit(51.65), Unit(47.57)},
				Point{Unit(54.37), Unit(53.01)},
				Point{Unit(54.37), Unit(58.44)},
				Point{Unit(48.93), Unit(63.88)},
				Point{Unit(40.77), Unit(66.60)},
				Point{Unit(35.34), Unit(66.60)},
				Point{Unit(27.18), Unit(63.88)},
				Point{Unit(21.75), Unit(61.16)},
				Point{Unit(16.31), Unit(55.72)},
				Point{Unit(13.59), Unit(50.29)},
				Point{Unit(10.87), Unit(42.13)},
				Point{Unit(10.87), Unit(33.98)},
				Point{Unit(13.59), Unit(25.82)},
				Point{Unit(16.31), Unit(20.39)},
				Point{Unit(21.75), Unit(14.95)},
				Point{Unit(27.18), Unit(12.23)},
				Point{Unit(35.34), Unit(9.51)},
				Point{Unit(43.49), Unit(9.51)},
				Point{Unit(51.65), Unit(12.23)},
				Point{Unit(57.08), Unit(14.95)},
				Point{Unit(62.52), Unit(20.39)},
				Point{Unit(65.24), Unit(25.82)},
				Point{Unit(67.96), Unit(33.98)},
				Point{Unit(67.96), Unit(42.13)},
				Point{Unit(65.24), Unit(50.29)},
				Point{Unit(62.52), Unit(55.72)},
				Point{Unit(59.80), Unit(58.44)},
			},
			Path{
				Point{Unit(24.46), Unit(53.01)},
				Point{Unit(46.21), Unit(50.29)},
				Point{Unit(51.65), Unit(50.29)},
				Point{Unit(54.37), Unit(53.01)},
			},
		},
		W: Unit(73.39356000000001),
	},
	'A': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(67.96), Unit(2.72)},
			},
			Path{
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(67.96), Unit(46.21)},
			},
			Path{
				Point{Unit(48.93), Unit(10.87)},
				Point{Unit(48.93), Unit(38.06)},
			},
		},
		W: Unit(48.92904),
	},
	'B': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(10.87), Unit(33.98)},
				Point{Unit(13.59), Unit(42.13)},
				Point{Unit(16.31), Unit(44.85)},
				Point{Unit(21.75), Unit(47.57)},
				Point{Unit(27.18), Unit(47.57)},
				Point{Unit(32.62), Unit(44.85)},
				Point{Unit(35.34), Unit(42.13)},
				Point{Unit(38.06), Unit(33.98)},
			},
			Path{
				Point{Unit(38.06), Unit(9.51)},
				Point{Unit(38.06), Unit(33.98)},
				Point{Unit(40.77), Unit(42.13)},
				Point{Unit(43.49), Unit(44.85)},
				Point{Unit(48.93), Unit(47.57)},
				Point{Unit(57.08), Unit(47.57)},
				Point{Unit(62.52), Unit(44.85)},
				Point{Unit(65.24), Unit(42.13)},
				Point{Unit(67.96), Unit(33.98)},
				Point{Unit(67.96), Unit(9.51)},
			},
		},
		W: Unit(57.08388),
	},
	'C': Glyph{
		S: Set{
			Path{
				Point{Unit(24.46), Unit(50.29)},
				Point{Unit(19.03), Unit(47.57)},
				Point{Unit(13.59), Unit(42.13)},
				Point{Unit(10.87), Unit(36.70)},
				Point{Unit(10.87), Unit(25.82)},
				Point{Unit(13.59), Unit(20.39)},
				Point{Unit(19.03), Unit(14.95)},
				Point{Unit(24.46), Unit(12.23)},
				Point{Unit(32.62), Unit(9.51)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(54.37), Unit(12.23)},
				Point{Unit(59.80), Unit(14.95)},
				Point{Unit(65.24), Unit(20.39)},
				Point{Unit(67.96), Unit(25.82)},
				Point{Unit(67.96), Unit(36.70)},
				Point{Unit(65.24), Unit(42.13)},
				Point{Unit(59.80), Unit(47.57)},
				Point{Unit(54.37), Unit(50.29)},
			},
		},
		W: Unit(57.08388),
	},
	'D': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(10.87), Unit(28.54)},
				Point{Unit(13.59), Unit(36.70)},
				Point{Unit(19.03), Unit(42.13)},
				Point{Unit(24.46), Unit(44.85)},
				Point{Unit(32.62), Unit(47.57)},
				Point{Unit(46.21), Unit(47.57)},
				Point{Unit(54.37), Unit(44.85)},
				Point{Unit(59.80), Unit(42.13)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(67.96), Unit(28.54)},
				Point{Unit(67.96), Unit(9.51)},
			},
		},
		W: Unit(57.08388),
	},
	'E': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(10.87), Unit(44.85)},
			},
			Path{
				Point{Unit(38.06), Unit(9.51)},
				Point{Unit(38.06), Unit(31.26)},
			},
			Path{
				Point{Unit(67.96), Unit(9.51)},
				Point{Unit(67.96), Unit(44.85)},
			},
		},
		W: Unit(51.64732),
	},
	'F': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(67.96), Unit(8.15)},
			},
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(10.87), Unit(43.49)},
			},
			Path{
				Point{Unit(38.06), Unit(8.15)},
				Point{Unit(38.06), Unit(29.90)},
			},
		},
		W: Unit(48.92904),
	},
	'G': Glyph{
		S: Set{
			Path{
				Point{Unit(24.46), Unit(50.29)},
				Point{Unit(19.03), Unit(47.57)},
				Point{Unit(13.59), Unit(42.13)},
				Point{Unit(10.87), Unit(36.70)},
				Point{Unit(10.87), Unit(25.82)},
				Point{Unit(13.59), Unit(20.39)},
				Point{Unit(19.03), Unit(14.95)},
				Point{Unit(24.46), Unit(12.23)},
				Point{Unit(32.62), Unit(9.51)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(54.37), Unit(12.23)},
				Point{Unit(59.80), Unit(14.95)},
				Point{Unit(65.24), Unit(20.39)},
				Point{Unit(67.96), Unit(25.82)},
				Point{Unit(67.96), Unit(36.70)},
				Point{Unit(65.24), Unit(42.13)},
				Point{Unit(59.80), Unit(47.57)},
				Point{Unit(54.37), Unit(50.29)},
				Point{Unit(46.21), Unit(50.29)},
			},
			Path{
				Point{Unit(46.21), Unit(36.70)},
				Point{Unit(46.21), Unit(50.29)},
			},
		},
		W: Unit(57.08388),
	},
	'H': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
			Path{
				Point{Unit(10.87), Unit(48.93)},
				Point{Unit(67.96), Unit(48.93)},
			},
			Path{
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(38.06), Unit(48.93)},
			},
		},
		W: Unit(59.80216),
	},
	'I': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
		},
		W: Unit(21.74624),
	},
	'J': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(54.37), Unit(32.62)},
				Point{Unit(62.52), Unit(29.90)},
				Point{Unit(65.24), Unit(27.18)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(67.96), Unit(16.31)},
				Point{Unit(65.24), Unit(10.87)},
				Point{Unit(62.52), Unit(8.15)},
				Point{Unit(54.37), Unit(5.44)},
				Point{Unit(48.93), Unit(5.44)},
			},
		},
		W: Unit(43.49248),
	},
	'K': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(47.57)},
				Point{Unit(48.93), Unit(9.51)},
			},
			Path{
				Point{Unit(35.34), Unit(23.11)},
				Point{Unit(67.96), Unit(47.57)},
			},
		},
		W: Unit(57.08388),
	},
	'L': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(6.80)},
				Point{Unit(67.96), Unit(6.80)},
			},
			Path{
				Point{Unit(67.96), Unit(6.80)},
				Point{Unit(67.96), Unit(39.42)},
			},
		},
		W: Unit(46.21076),
	},
	'M': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(32.62)},
			},
			Path{
				Point{Unit(10.87), Unit(54.37)},
				Point{Unit(67.96), Unit(32.62)},
			},
			Path{
				Point{Unit(10.87), Unit(54.37)},
				Point{Unit(67.96), Unit(54.37)},
			},
		},
		W: Unit(65.23872),
	},
	'N': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(48.93)},
			},
			Path{
				Point{Unit(10.87), Unit(48.93)},
				Point{Unit(67.96), Unit(48.93)},
			},
		},
		W: Unit(59.80216),
	},
	'O': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(19.03), Unit(13.59)},
				Point{Unit(24.46), Unit(10.87)},
				Point{Unit(32.62), Unit(8.15)},
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(54.37), Unit(10.87)},
				Point{Unit(59.80), Unit(13.59)},
				Point{Unit(65.24), Unit(19.03)},
				Point{Unit(67.96), Unit(24.46)},
				Point{Unit(67.96), Unit(35.34)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(59.80), Unit(46.21)},
				Point{Unit(54.37), Unit(48.93)},
				Point{Unit(46.21), Unit(51.65)},
				Point{Unit(32.62), Unit(51.65)},
				Point{Unit(24.46), Unit(48.93)},
				Point{Unit(19.03), Unit(46.21)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(35.34)},
				Point{Unit(10.87), Unit(24.46)},
			},
		},
		W: Unit(59.80216),
	},
	'P': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(10.87), Unit(33.98)},
				Point{Unit(13.59), Unit(42.13)},
				Point{Unit(16.31), Unit(44.85)},
				Point{Unit(21.75), Unit(47.57)},
				Point{Unit(29.90), Unit(47.57)},
				Point{Unit(35.34), Unit(44.85)},
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(40.77), Unit(33.98)},
				Point{Unit(40.77), Unit(9.51)},
			},
		},
		W: Unit(57.08388),
	},
	'Q': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(19.03), Unit(13.59)},
				Point{Unit(24.46), Unit(10.87)},
				Point{Unit(32.62), Unit(8.15)},
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(54.37), Unit(10.87)},
				Point{Unit(59.80), Unit(13.59)},
				Point{Unit(65.24), Unit(19.03)},
				Point{Unit(67.96), Unit(24.46)},
				Point{Unit(67.96), Unit(35.34)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(59.80), Unit(46.21)},
				Point{Unit(54.37), Unit(48.93)},
				Point{Unit(46.21), Unit(51.65)},
				Point{Unit(32.62), Unit(51.65)},
				Point{Unit(24.46), Unit(48.93)},
				Point{Unit(19.03), Unit(46.21)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(35.34)},
				Point{Unit(10.87), Unit(24.46)},
			},
			Path{
				Point{Unit(57.08), Unit(32.62)},
				Point{Unit(73.39), Unit(48.93)},
			},
		},
		W: Unit(59.80216),
	},
	'R': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(10.87), Unit(33.98)},
				Point{Unit(13.59), Unit(42.13)},
				Point{Unit(16.31), Unit(44.85)},
				Point{Unit(21.75), Unit(47.57)},
				Point{Unit(27.18), Unit(47.57)},
				Point{Unit(32.62), Unit(44.85)},
				Point{Unit(35.34), Unit(42.13)},
				Point{Unit(38.06), Unit(33.98)},
				Point{Unit(38.06), Unit(9.51)},
			},
			Path{
				Point{Unit(38.06), Unit(28.54)},
				Point{Unit(67.96), Unit(47.57)},
			},
		},
		W: Unit(57.08388),
	},
	'S': Glyph{
		S: Set{
			Path{
				Point{Unit(19.03), Unit(46.21)},
				Point{Unit(13.59), Unit(40.77)},
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(13.59), Unit(13.59)},
				Point{Unit(19.03), Unit(8.15)},
				Point{Unit(24.46), Unit(8.15)},
				Point{Unit(29.90), Unit(10.87)},
				Point{Unit(32.62), Unit(13.59)},
				Point{Unit(35.34), Unit(19.03)},
				Point{Unit(40.77), Unit(35.34)},
				Point{Unit(43.49), Unit(40.77)},
				Point{Unit(46.21), Unit(43.49)},
				Point{Unit(51.65), Unit(46.21)},
				Point{Unit(59.80), Unit(46.21)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(67.96), Unit(32.62)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(65.24), Unit(13.59)},
				Point{Unit(59.80), Unit(8.15)},
			},
		},
		W: Unit(54.3656),
	},
	'T': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(67.96), Unit(21.75)},
			},
			Path{
				Point{Unit(10.87), Unit(2.72)},
				Point{Unit(10.87), Unit(40.77)},
			},
		},
		W: Unit(43.49248),
	},
	'U': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(51.65), Unit(10.87)},
				Point{Unit(59.80), Unit(13.59)},
				Point{Unit(65.24), Unit(19.03)},
				Point{Unit(67.96), Unit(27.18)},
				Point{Unit(67.96), Unit(32.62)},
				Point{Unit(65.24), Unit(40.77)},
				Point{Unit(59.80), Unit(46.21)},
				Point{Unit(51.65), Unit(48.93)},
				Point{Unit(10.87), Unit(48.93)},
			},
		},
		W: Unit(59.80216),
	},
	'V': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(2.72)},
				Point{Unit(67.96), Unit(24.46)},
			},
			Path{
				Point{Unit(10.87), Unit(46.21)},
				Point{Unit(67.96), Unit(24.46)},
			},
		},
		W: Unit(48.92904),
	},
	'W': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(5.44)},
				Point{Unit(67.96), Unit(19.03)},
			},
			Path{
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(67.96), Unit(19.03)},
			},
			Path{
				Point{Unit(10.87), Unit(32.62)},
				Point{Unit(67.96), Unit(46.21)},
			},
			Path{
				Point{Unit(10.87), Unit(59.80)},
				Point{Unit(67.96), Unit(46.21)},
			},
		},
		W: Unit(65.23872),
	},
	'X': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(67.96), Unit(46.21)},
			},
			Path{
				Point{Unit(10.87), Unit(46.21)},
				Point{Unit(67.96), Unit(8.15)},
			},
		},
		W: Unit(54.3656),
	},
	'Y': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(2.72)},
				Point{Unit(38.06), Unit(24.46)},
				Point{Unit(67.96), Unit(24.46)},
			},
			Path{
				Point{Unit(10.87), Unit(46.21)},
				Point{Unit(38.06), Unit(24.46)},
			},
		},
		W: Unit(48.92904),
	},
	'Z': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(46.21)},
				Point{Unit(67.96), Unit(8.15)},
			},
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(10.87), Unit(46.21)},
			},
			Path{
				Point{Unit(67.96), Unit(8.15)},
				Point{Unit(67.96), Unit(46.21)},
			},
		},
		W: Unit(54.3656),
	},
	'[': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(10.87)},
				Point{Unit(86.98), Unit(10.87)},
			},
			Path{
				Point{Unit(0.00), Unit(13.59)},
				Point{Unit(86.98), Unit(13.59)},
			},
			Path{
				Point{Unit(0.00), Unit(10.87)},
				Point{Unit(0.00), Unit(29.90)},
			},
			Path{
				Point{Unit(86.98), Unit(10.87)},
				Point{Unit(86.98), Unit(29.90)},
			},
		},
		W: Unit(38.05592),
	},
	'\\': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(0.00)},
				Point{Unit(76.11), Unit(38.06)},
			},
		},
		W: Unit(38.05592),
	},
	']': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(24.46)},
				Point{Unit(86.98), Unit(24.46)},
			},
			Path{
				Point{Unit(0.00), Unit(27.18)},
				Point{Unit(86.98), Unit(27.18)},
			},
			Path{
				Point{Unit(0.00), Unit(8.15)},
				Point{Unit(0.00), Unit(27.18)},
			},
			Path{
				Point{Unit(86.98), Unit(8.15)},
				Point{Unit(86.98), Unit(27.18)},
			},
		},
		W: Unit(38.05592),
	},
	'^': Glyph{
		S: Set{
			Path{
				Point{Unit(27.18), Unit(16.31)},
				Point{Unit(19.03), Unit(21.75)},
				Point{Unit(27.18), Unit(27.18)},
			},
			Path{
				Point{Unit(35.34), Unit(8.15)},
				Point{Unit(21.75), Unit(21.75)},
				Point{Unit(35.34), Unit(35.34)},
			},
			Path{
				Point{Unit(21.75), Unit(21.75)},
				Point{Unit(67.96), Unit(21.75)},
			},
		},
		W: Unit(43.49248),
	},
	'_': Glyph{
		S: Set{
			Path{
				Point{Unit(73.39), Unit(0.00)},
				Point{Unit(73.39), Unit(43.49)},
			},
		},
		W: Unit(43.49248),
	},
	'`': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(16.31)},
				Point{Unit(13.59), Unit(13.59)},
				Point{Unit(19.03), Unit(10.87)},
				Point{Unit(24.46), Unit(10.87)},
				Point{Unit(27.18), Unit(13.59)},
				Point{Unit(24.46), Unit(16.31)},
				Point{Unit(21.75), Unit(13.59)},
			},
		},
		W: Unit(27.1828),
	},
	'a': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(42.13)},
				Point{Unit(67.96), Unit(42.13)},
			},
			Path{
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(31.26)},
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(59.80), Unit(12.23)},
				Point{Unit(65.24), Unit(17.67)},
				Point{Unit(67.96), Unit(23.11)},
				Point{Unit(67.96), Unit(31.26)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(59.80), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'b': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(38.06), Unit(9.51)},
				Point{Unit(32.62), Unit(14.95)},
				Point{Unit(29.90), Unit(20.39)},
				Point{Unit(29.90), Unit(28.54)},
				Point{Unit(32.62), Unit(33.98)},
				Point{Unit(38.06), Unit(39.42)},
				Point{Unit(46.21), Unit(42.13)},
				Point{Unit(51.65), Unit(42.13)},
				Point{Unit(59.80), Unit(39.42)},
				Point{Unit(65.24), Unit(33.98)},
				Point{Unit(67.96), Unit(28.54)},
				Point{Unit(67.96), Unit(20.39)},
				Point{Unit(65.24), Unit(14.95)},
				Point{Unit(59.80), Unit(9.51)},
			},
		},
		W: Unit(51.64732),
	},
	'c': Glyph{
		S: Set{
			Path{
				Point{Unit(38.06), Unit(40.77)},
				Point{Unit(32.62), Unit(35.34)},
				Point{Unit(29.90), Unit(29.90)},
				Point{Unit(29.90), Unit(21.75)},
				Point{Unit(32.62), Unit(16.31)},
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(51.65), Unit(8.15)},
				Point{Unit(59.80), Unit(10.87)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(65.24), Unit(35.34)},
				Point{Unit(59.80), Unit(40.77)},
			},
		},
		W: Unit(48.92904),
	},
	'd': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(42.13)},
				Point{Unit(67.96), Unit(42.13)},
			},
			Path{
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(31.26)},
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(59.80), Unit(12.23)},
				Point{Unit(65.24), Unit(17.67)},
				Point{Unit(67.96), Unit(23.11)},
				Point{Unit(67.96), Unit(31.26)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(59.80), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'e': Glyph{
		S: Set{
			Path{
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(46.21), Unit(40.77)},
				Point{Unit(40.77), Unit(40.77)},
				Point{Unit(35.34), Unit(38.06)},
				Point{Unit(32.62), Unit(35.34)},
				Point{Unit(29.90), Unit(29.90)},
				Point{Unit(29.90), Unit(21.75)},
				Point{Unit(32.62), Unit(16.31)},
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(51.65), Unit(8.15)},
				Point{Unit(59.80), Unit(10.87)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(67.96), Unit(29.90)},
				Point{Unit(65.24), Unit(35.34)},
				Point{Unit(59.80), Unit(40.77)},
			},
		},
		W: Unit(48.92904),
	},
	'f': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(29.90)},
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(13.59), Unit(19.03)},
				Point{Unit(21.75), Unit(16.31)},
				Point{Unit(67.96), Unit(16.31)},
			},
			Path{
				Point{Unit(29.90), Unit(8.15)},
				Point{Unit(29.90), Unit(27.18)},
			},
		},
		W: Unit(32.61936),
	},
	'g': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(42.13)},
				Point{Unit(73.39), Unit(42.13)},
				Point{Unit(81.55), Unit(39.42)},
				Point{Unit(84.27), Unit(36.70)},
				Point{Unit(86.98), Unit(31.26)},
				Point{Unit(86.98), Unit(23.11)},
				Point{Unit(84.27), Unit(17.67)},
			},
			Path{
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(31.26)},
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(59.80), Unit(12.23)},
				Point{Unit(65.24), Unit(17.67)},
				Point{Unit(67.96), Unit(23.11)},
				Point{Unit(67.96), Unit(31.26)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(59.80), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'h': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(12.23)},
				Point{Unit(67.96), Unit(12.23)},
			},
			Path{
				Point{Unit(40.77), Unit(12.23)},
				Point{Unit(32.62), Unit(20.39)},
				Point{Unit(29.90), Unit(25.82)},
				Point{Unit(29.90), Unit(33.98)},
				Point{Unit(32.62), Unit(39.42)},
				Point{Unit(40.77), Unit(42.13)},
				Point{Unit(67.96), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'i': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(8.15)},
				Point{Unit(13.59), Unit(10.87)},
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(8.15), Unit(10.87)},
				Point{Unit(10.87), Unit(8.15)},
			},
			Path{
				Point{Unit(29.90), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
		},
		W: Unit(21.74624),
	},
	'j': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(10.87), Unit(19.03)},
				Point{Unit(8.15), Unit(16.31)},
				Point{Unit(10.87), Unit(13.59)},
			},
			Path{
				Point{Unit(29.90), Unit(16.31)},
				Point{Unit(76.11), Unit(16.31)},
				Point{Unit(84.27), Unit(13.59)},
				Point{Unit(86.98), Unit(8.15)},
				Point{Unit(86.98), Unit(2.72)},
			},
		},
		W: Unit(27.1828),
	},
	'k': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(29.90), Unit(36.70)},
				Point{Unit(57.08), Unit(9.51)},
			},
			Path{
				Point{Unit(46.21), Unit(20.39)},
				Point{Unit(67.96), Unit(39.42)},
			},
		},
		W: Unit(46.21076),
	},
	'l': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
		},
		W: Unit(21.74624),
	},
	'm': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(10.87)},
				Point{Unit(67.96), Unit(10.87)},
			},
			Path{
				Point{Unit(40.77), Unit(10.87)},
				Point{Unit(32.62), Unit(19.03)},
				Point{Unit(29.90), Unit(24.46)},
				Point{Unit(29.90), Unit(32.62)},
				Point{Unit(32.62), Unit(38.06)},
				Point{Unit(40.77), Unit(40.77)},
				Point{Unit(67.96), Unit(40.77)},
			},
			Path{
				Point{Unit(40.77), Unit(40.77)},
				Point{Unit(32.62), Unit(48.93)},
				Point{Unit(29.90), Unit(54.37)},
				Point{Unit(29.90), Unit(62.52)},
				Point{Unit(32.62), Unit(67.96)},
				Point{Unit(40.77), Unit(70.68)},
				Point{Unit(67.96), Unit(70.68)},
			},
		},
		W: Unit(81.5484),
	},
	'n': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(12.23)},
				Point{Unit(67.96), Unit(12.23)},
			},
			Path{
				Point{Unit(40.77), Unit(12.23)},
				Point{Unit(32.62), Unit(20.39)},
				Point{Unit(29.90), Unit(25.82)},
				Point{Unit(29.90), Unit(33.98)},
				Point{Unit(32.62), Unit(39.42)},
				Point{Unit(40.77), Unit(42.13)},
				Point{Unit(67.96), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'o': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(59.80), Unit(12.23)},
				Point{Unit(65.24), Unit(17.67)},
				Point{Unit(67.96), Unit(23.11)},
				Point{Unit(67.96), Unit(31.26)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(59.80), Unit(42.13)},
				Point{Unit(51.65), Unit(44.85)},
				Point{Unit(46.21), Unit(44.85)},
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(31.26)},
				Point{Unit(29.90), Unit(23.11)},
			},
		},
		W: Unit(51.64732),
	},
	'p': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(9.51)},
				Point{Unit(86.98), Unit(9.51)},
			},
			Path{
				Point{Unit(38.06), Unit(9.51)},
				Point{Unit(32.62), Unit(14.95)},
				Point{Unit(29.90), Unit(20.39)},
				Point{Unit(29.90), Unit(28.54)},
				Point{Unit(32.62), Unit(33.98)},
				Point{Unit(38.06), Unit(39.42)},
				Point{Unit(46.21), Unit(42.13)},
				Point{Unit(51.65), Unit(42.13)},
				Point{Unit(59.80), Unit(39.42)},
				Point{Unit(65.24), Unit(33.98)},
				Point{Unit(67.96), Unit(28.54)},
				Point{Unit(67.96), Unit(20.39)},
				Point{Unit(65.24), Unit(14.95)},
				Point{Unit(59.80), Unit(9.51)},
			},
		},
		W: Unit(51.64732),
	},
	'q': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(42.13)},
				Point{Unit(86.98), Unit(42.13)},
			},
			Path{
				Point{Unit(38.06), Unit(42.13)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(31.26)},
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(51.65), Unit(9.51)},
				Point{Unit(59.80), Unit(12.23)},
				Point{Unit(65.24), Unit(17.67)},
				Point{Unit(67.96), Unit(23.11)},
				Point{Unit(67.96), Unit(31.26)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(59.80), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'r': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(9.51)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(46.21), Unit(9.51)},
				Point{Unit(38.06), Unit(12.23)},
				Point{Unit(32.62), Unit(17.67)},
				Point{Unit(29.90), Unit(23.11)},
				Point{Unit(29.90), Unit(31.26)},
			},
		},
		W: Unit(35.33764),
	},
	's': Glyph{
		S: Set{
			Path{
				Point{Unit(38.06), Unit(39.42)},
				Point{Unit(32.62), Unit(36.70)},
				Point{Unit(29.90), Unit(28.54)},
				Point{Unit(29.90), Unit(20.39)},
				Point{Unit(32.62), Unit(12.23)},
				Point{Unit(38.06), Unit(9.51)},
				Point{Unit(43.49), Unit(12.23)},
				Point{Unit(46.21), Unit(17.67)},
				Point{Unit(48.93), Unit(31.26)},
				Point{Unit(51.65), Unit(36.70)},
				Point{Unit(57.08), Unit(39.42)},
				Point{Unit(59.80), Unit(39.42)},
				Point{Unit(65.24), Unit(36.70)},
				Point{Unit(67.96), Unit(28.54)},
				Point{Unit(67.96), Unit(20.39)},
				Point{Unit(65.24), Unit(12.23)},
				Point{Unit(59.80), Unit(9.51)},
			},
		},
		W: Unit(46.21076),
	},
	't': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(16.31)},
				Point{Unit(57.08), Unit(16.31)},
				Point{Unit(65.24), Unit(19.03)},
				Point{Unit(67.96), Unit(24.46)},
				Point{Unit(67.96), Unit(29.90)},
			},
			Path{
				Point{Unit(29.90), Unit(8.15)},
				Point{Unit(29.90), Unit(27.18)},
			},
		},
		W: Unit(32.61936),
	},
	'u': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(12.23)},
				Point{Unit(57.08), Unit(12.23)},
				Point{Unit(65.24), Unit(14.95)},
				Point{Unit(67.96), Unit(20.39)},
				Point{Unit(67.96), Unit(28.54)},
				Point{Unit(65.24), Unit(33.98)},
				Point{Unit(57.08), Unit(42.13)},
			},
			Path{
				Point{Unit(29.90), Unit(42.13)},
				Point{Unit(67.96), Unit(42.13)},
			},
		},
		W: Unit(51.64732),
	},
	'v': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(5.44)},
				Point{Unit(67.96), Unit(21.75)},
			},
			Path{
				Point{Unit(29.90), Unit(38.06)},
				Point{Unit(67.96), Unit(21.75)},
			},
		},
		W: Unit(43.49248),
	},
	'w': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(8.15)},
				Point{Unit(67.96), Unit(19.03)},
			},
			Path{
				Point{Unit(29.90), Unit(29.90)},
				Point{Unit(67.96), Unit(19.03)},
			},
			Path{
				Point{Unit(29.90), Unit(29.90)},
				Point{Unit(67.96), Unit(40.77)},
			},
			Path{
				Point{Unit(29.90), Unit(51.65)},
				Point{Unit(67.96), Unit(40.77)},
			},
		},
		W: Unit(59.80216),
	},
	'x': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(9.51)},
				Point{Unit(67.96), Unit(39.42)},
			},
			Path{
				Point{Unit(29.90), Unit(39.42)},
				Point{Unit(67.96), Unit(9.51)},
			},
		},
		W: Unit(46.21076),
	},
	'y': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(5.44)},
				Point{Unit(67.96), Unit(21.75)},
			},
			Path{
				Point{Unit(29.90), Unit(38.06)},
				Point{Unit(67.96), Unit(21.75)},
				Point{Unit(78.83), Unit(16.31)},
				Point{Unit(84.27), Unit(10.87)},
				Point{Unit(86.98), Unit(5.44)},
				Point{Unit(86.98), Unit(2.72)},
			},
		},
		W: Unit(43.49248),
	},
	'z': Glyph{
		S: Set{
			Path{
				Point{Unit(29.90), Unit(39.42)},
				Point{Unit(67.96), Unit(9.51)},
			},
			Path{
				Point{Unit(29.90), Unit(9.51)},
				Point{Unit(29.90), Unit(39.42)},
			},
			Path{
				Point{Unit(67.96), Unit(9.51)},
				Point{Unit(67.96), Unit(39.42)},
			},
		},
		W: Unit(46.21076),
	},
	'{': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(24.46)},
				Point{Unit(2.72), Unit(19.03)},
				Point{Unit(5.44), Unit(16.31)},
				Point{Unit(10.87), Unit(13.59)},
				Point{Unit(16.31), Unit(13.59)},
				Point{Unit(21.75), Unit(16.31)},
				Point{Unit(24.46), Unit(19.03)},
				Point{Unit(29.90), Unit(21.75)},
				Point{Unit(35.34), Unit(21.75)},
				Point{Unit(40.77), Unit(16.31)},
			},
			Path{
				Point{Unit(2.72), Unit(19.03)},
				Point{Unit(8.15), Unit(16.31)},
				Point{Unit(13.59), Unit(16.31)},
				Point{Unit(19.03), Unit(19.03)},
				Point{Unit(21.75), Unit(21.75)},
				Point{Unit(27.18), Unit(24.46)},
				Point{Unit(32.62), Unit(24.46)},
				Point{Unit(38.06), Unit(21.75)},
				Point{Unit(43.49), Unit(10.87)},
				Point{Unit(48.93), Unit(21.75)},
				Point{Unit(54.37), Unit(24.46)},
				Point{Unit(59.80), Unit(24.46)},
				Point{Unit(65.24), Unit(21.75)},
				Point{Unit(67.96), Unit(19.03)},
				Point{Unit(73.39), Unit(16.31)},
				Point{Unit(78.83), Unit(16.31)},
				Point{Unit(84.27), Unit(19.03)},
			},
			Path{
				Point{Unit(46.21), Unit(16.31)},
				Point{Unit(51.65), Unit(21.75)},
				Point{Unit(57.08), Unit(21.75)},
				Point{Unit(62.52), Unit(19.03)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(70.68), Unit(13.59)},
				Point{Unit(76.11), Unit(13.59)},
				Point{Unit(81.55), Unit(16.31)},
				Point{Unit(84.27), Unit(19.03)},
				Point{Unit(86.98), Unit(24.46)},
			},
		},
		W: Unit(38.05592),
	},
	'|': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(10.87)},
				Point{Unit(86.98), Unit(10.87)},
			},
		},
		W: Unit(21.74624),
	},
	'}': Glyph{
		S: Set{
			Path{
				Point{Unit(0.00), Unit(13.59)},
				Point{Unit(2.72), Unit(19.03)},
				Point{Unit(5.44), Unit(21.75)},
				Point{Unit(10.87), Unit(24.46)},
				Point{Unit(16.31), Unit(24.46)},
				Point{Unit(21.75), Unit(21.75)},
				Point{Unit(24.46), Unit(19.03)},
				Point{Unit(29.90), Unit(16.31)},
				Point{Unit(35.34), Unit(16.31)},
				Point{Unit(40.77), Unit(21.75)},
			},
			Path{
				Point{Unit(2.72), Unit(19.03)},
				Point{Unit(8.15), Unit(21.75)},
				Point{Unit(13.59), Unit(21.75)},
				Point{Unit(19.03), Unit(19.03)},
				Point{Unit(21.75), Unit(16.31)},
				Point{Unit(27.18), Unit(13.59)},
				Point{Unit(32.62), Unit(13.59)},
				Point{Unit(38.06), Unit(16.31)},
				Point{Unit(43.49), Unit(27.18)},
				Point{Unit(48.93), Unit(16.31)},
				Point{Unit(54.37), Unit(13.59)},
				Point{Unit(59.80), Unit(13.59)},
				Point{Unit(65.24), Unit(16.31)},
				Point{Unit(67.96), Unit(19.03)},
				Point{Unit(73.39), Unit(21.75)},
				Point{Unit(78.83), Unit(21.75)},
				Point{Unit(84.27), Unit(19.03)},
			},
			Path{
				Point{Unit(46.21), Unit(21.75)},
				Point{Unit(51.65), Unit(16.31)},
				Point{Unit(57.08), Unit(16.31)},
				Point{Unit(62.52), Unit(19.03)},
				Point{Unit(65.24), Unit(21.75)},
				Point{Unit(70.68), Unit(24.46)},
				Point{Unit(76.11), Unit(24.46)},
				Point{Unit(81.55), Unit(21.75)},
				Point{Unit(84.27), Unit(19.03)},
				Point{Unit(86.98), Unit(13.59)},
			},
		},
		W: Unit(38.05592),
	},
	'~': Glyph{
		S: Set{
			Path{
				Point{Unit(51.65), Unit(8.15)},
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(38.06), Unit(10.87)},
				Point{Unit(35.34), Unit(16.31)},
				Point{Unit(35.34), Unit(21.75)},
				Point{Unit(38.06), Unit(27.18)},
				Point{Unit(46.21), Unit(38.06)},
				Point{Unit(48.93), Unit(43.49)},
				Point{Unit(48.93), Unit(48.93)},
				Point{Unit(46.21), Unit(54.37)},
				Point{Unit(40.77), Unit(57.08)},
			},
			Path{
				Point{Unit(46.21), Unit(8.15)},
				Point{Unit(40.77), Unit(10.87)},
				Point{Unit(38.06), Unit(16.31)},
				Point{Unit(38.06), Unit(21.75)},
				Point{Unit(40.77), Unit(27.18)},
				Point{Unit(48.93), Unit(38.06)},
				Point{Unit(51.65), Unit(43.49)},
				Point{Unit(51.65), Unit(48.93)},
				Point{Unit(48.93), Unit(54.37)},
				Point{Unit(40.77), Unit(57.08)},
				Point{Unit(35.34), Unit(57.08)},
			},
		},
		W: Unit(65.23872),
	},
	'\u007f': Glyph{
		S: Set{
			Path{
				Point{Unit(10.87), Unit(16.31)},
				Point{Unit(13.59), Unit(10.87)},
				Point{Unit(19.03), Unit(8.15)},
				Point{Unit(24.46), Unit(8.15)},
				Point{Unit(29.90), Unit(10.87)},
				Point{Unit(32.62), Unit(16.31)},
				Point{Unit(32.62), Unit(21.75)},
				Point{Unit(29.90), Unit(27.18)},
				Point{Unit(24.46), Unit(29.90)},
				Point{Unit(19.03), Unit(29.90)},
				Point{Unit(13.59), Unit(27.18)},
				Point{Unit(10.87), Unit(21.75)},
				Point{Unit(10.87), Unit(16.31)},
			},
		},
		W: Unit(38.05592),
	},
}
