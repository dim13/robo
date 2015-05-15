package main

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
		off.X += 100
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
				off.X += 100
				off.Y = 0
			}
			off.Offset(c)
			for _, p := range gl.S {
				p[0].Move(c)
				p[1:].Draw(c)
			}
			off.Y += gl.W
		}
	}
	return off
}

var font = Font{
	' ': Glyph{
		S: Set{},
		W: Unit(48),
	},
	'!': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(15)},
				Point{Unit(54), Unit(15)},
			},
			Path{
				Point{Unit(69), Unit(15)},
				Point{Unit(72), Unit(12)},
				Point{Unit(75), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(69), Unit(15)},
			},
		},
		W: Unit(30),
	},
	'"': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(33), Unit(12)},
			},
			Path{
				Point{Unit(12), Unit(36)},
				Point{Unit(33), Unit(36)},
			},
		},
		W: Unit(48),
	},
	'#': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(34.5)},
				Point{Unit(96), Unit(13.5)},
			},
			Path{
				Point{Unit(0), Unit(52.5)},
				Point{Unit(96), Unit(31.5)},
			},
			Path{
				Point{Unit(39), Unit(13.5)},
				Point{Unit(39), Unit(55.5)},
			},
			Path{
				Point{Unit(57), Unit(10.5)},
				Point{Unit(57), Unit(52.5)},
			},
		},
		W: Unit(63),
	},
	'$': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(24)},
				Point{Unit(87), Unit(24)},
			},
			Path{
				Point{Unit(0), Unit(36)},
				Point{Unit(87), Unit(36)},
			},
			Path{
				Point{Unit(21), Unit(51)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(36)},
				Point{Unit(12), Unit(24)},
				Point{Unit(15), Unit(15)},
				Point{Unit(21), Unit(9)},
				Point{Unit(27), Unit(9)},
				Point{Unit(33), Unit(12)},
				Point{Unit(36), Unit(15)},
				Point{Unit(39), Unit(21)},
				Point{Unit(45), Unit(39)},
				Point{Unit(48), Unit(45)},
				Point{Unit(51), Unit(48)},
				Point{Unit(57), Unit(51)},
				Point{Unit(66), Unit(51)},
				Point{Unit(72), Unit(45)},
				Point{Unit(75), Unit(36)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(66), Unit(9)},
			},
		},
		W: Unit(60),
	},
	'%': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(63)},
				Point{Unit(75), Unit(9)},
			},
			Path{
				Point{Unit(12), Unit(24)},
				Point{Unit(18), Unit(30)},
				Point{Unit(24), Unit(30)},
				Point{Unit(30), Unit(27)},
				Point{Unit(33), Unit(21)},
				Point{Unit(33), Unit(15)},
				Point{Unit(27), Unit(9)},
				Point{Unit(21), Unit(9)},
				Point{Unit(15), Unit(12)},
				Point{Unit(12), Unit(18)},
				Point{Unit(12), Unit(24)},
				Point{Unit(15), Unit(30)},
				Point{Unit(18), Unit(39)},
				Point{Unit(18), Unit(48)},
				Point{Unit(15), Unit(57)},
				Point{Unit(12), Unit(63)},
			},
			Path{
				Point{Unit(54), Unit(51)},
				Point{Unit(57), Unit(45)},
				Point{Unit(63), Unit(42)},
				Point{Unit(69), Unit(42)},
				Point{Unit(75), Unit(48)},
				Point{Unit(75), Unit(54)},
				Point{Unit(72), Unit(60)},
				Point{Unit(66), Unit(63)},
				Point{Unit(60), Unit(63)},
				Point{Unit(54), Unit(57)},
				Point{Unit(54), Unit(51)},
			},
		},
		W: Unit(72),
	},
	'&': Glyph{
		S: Set{
			Path{
				Point{Unit(39), Unit(69)},
				Point{Unit(36), Unit(69)},
				Point{Unit(33), Unit(66)},
				Point{Unit(33), Unit(63)},
				Point{Unit(36), Unit(60)},
				Point{Unit(42), Unit(57)},
				Point{Unit(57), Unit(51)},
				Point{Unit(66), Unit(45)},
				Point{Unit(72), Unit(39)},
				Point{Unit(75), Unit(33)},
				Point{Unit(75), Unit(21)},
				Point{Unit(72), Unit(15)},
				Point{Unit(69), Unit(12)},
				Point{Unit(63), Unit(9)},
				Point{Unit(57), Unit(9)},
				Point{Unit(51), Unit(12)},
				Point{Unit(48), Unit(15)},
				Point{Unit(36), Unit(36)},
				Point{Unit(33), Unit(39)},
				Point{Unit(27), Unit(42)},
				Point{Unit(21), Unit(42)},
				Point{Unit(15), Unit(39)},
				Point{Unit(12), Unit(33)},
				Point{Unit(15), Unit(27)},
				Point{Unit(21), Unit(24)},
				Point{Unit(27), Unit(24)},
				Point{Unit(36), Unit(27)},
				Point{Unit(45), Unit(33)},
				Point{Unit(66), Unit(48)},
				Point{Unit(72), Unit(54)},
				Point{Unit(75), Unit(60)},
				Point{Unit(75), Unit(66)},
				Point{Unit(72), Unit(69)},
				Point{Unit(69), Unit(69)},
			},
		},
		W: Unit(78),
	},
	'\'': Glyph{
		S: Set{
			Path{
				Point{Unit(18), Unit(15)},
				Point{Unit(15), Unit(12)},
				Point{Unit(12), Unit(15)},
				Point{Unit(15), Unit(18)},
				Point{Unit(21), Unit(18)},
				Point{Unit(27), Unit(15)},
				Point{Unit(30), Unit(12)},
			},
		},
		W: Unit(30),
	},
	'(': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(33)},
				Point{Unit(6), Unit(27)},
				Point{Unit(15), Unit(21)},
				Point{Unit(27), Unit(15)},
				Point{Unit(42), Unit(12)},
				Point{Unit(54), Unit(12)},
				Point{Unit(69), Unit(15)},
				Point{Unit(81), Unit(21)},
				Point{Unit(90), Unit(27)},
				Point{Unit(96), Unit(33)},
			},
		},
		W: Unit(42),
	},
	')': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(9)},
				Point{Unit(6), Unit(15)},
				Point{Unit(15), Unit(21)},
				Point{Unit(27), Unit(27)},
				Point{Unit(42), Unit(30)},
				Point{Unit(54), Unit(30)},
				Point{Unit(69), Unit(27)},
				Point{Unit(81), Unit(21)},
				Point{Unit(90), Unit(15)},
				Point{Unit(96), Unit(9)},
			},
		},
		W: Unit(42),
	},
	'*': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(24)},
				Point{Unit(48), Unit(24)},
			},
			Path{
				Point{Unit(21), Unit(9)},
				Point{Unit(39), Unit(39)},
			},
			Path{
				Point{Unit(21), Unit(39)},
				Point{Unit(39), Unit(9)},
			},
		},
		W: Unit(48),
	},
	'+': Glyph{
		S: Set{
			Path{
				Point{Unit(21), Unit(39)},
				Point{Unit(75), Unit(39)},
			},
			Path{
				Point{Unit(48), Unit(12)},
				Point{Unit(48), Unit(66)},
			},
		},
		W: Unit(78),
	},
	',': Glyph{
		S: Set{
			Path{
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(15)},
				Point{Unit(72), Unit(12)},
				Point{Unit(69), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(78), Unit(18)},
				Point{Unit(84), Unit(15)},
				Point{Unit(87), Unit(12)},
			},
		},
		W: Unit(30),
	},
	'-': Glyph{
		S: Set{
			Path{
				Point{Unit(48), Unit(12)},
				Point{Unit(48), Unit(66)},
			},
		},
		W: Unit(78),
	},
	'.': Glyph{
		S: Set{
			Path{
				Point{Unit(69), Unit(15)},
				Point{Unit(72), Unit(12)},
				Point{Unit(75), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(69), Unit(15)},
			},
		},
		W: Unit(30),
	},
	'/': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(60)},
				Point{Unit(96), Unit(6)},
			},
		},
		W: Unit(66),
	},
	'0': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(27)},
				Point{Unit(15), Unit(18)},
				Point{Unit(24), Unit(12)},
				Point{Unit(39), Unit(9)},
				Point{Unit(48), Unit(9)},
				Point{Unit(63), Unit(12)},
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(27)},
				Point{Unit(75), Unit(33)},
				Point{Unit(72), Unit(42)},
				Point{Unit(63), Unit(48)},
				Point{Unit(48), Unit(51)},
				Point{Unit(39), Unit(51)},
				Point{Unit(24), Unit(48)},
				Point{Unit(15), Unit(42)},
				Point{Unit(12), Unit(33)},
				Point{Unit(12), Unit(27)},
			},
		},
		W: Unit(60),
	},
	'1': Glyph{
		S: Set{
			Path{
				Point{Unit(24), Unit(18)},
				Point{Unit(21), Unit(24)},
				Point{Unit(12), Unit(33)},
				Point{Unit(75), Unit(33)},
			},
		},
		W: Unit(60),
	},
	'2': Glyph{
		S: Set{
			Path{
				Point{Unit(27), Unit(12)},
				Point{Unit(24), Unit(12)},
				Point{Unit(18), Unit(15)},
				Point{Unit(15), Unit(18)},
				Point{Unit(12), Unit(24)},
				Point{Unit(12), Unit(36)},
				Point{Unit(15), Unit(42)},
				Point{Unit(18), Unit(45)},
				Point{Unit(24), Unit(48)},
				Point{Unit(30), Unit(48)},
				Point{Unit(36), Unit(45)},
				Point{Unit(45), Unit(39)},
				Point{Unit(75), Unit(9)},
				Point{Unit(75), Unit(51)},
			},
		},
		W: Unit(60),
	},
	'3': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(15)},
				Point{Unit(12), Unit(48)},
				Point{Unit(36), Unit(30)},
				Point{Unit(36), Unit(39)},
				Point{Unit(39), Unit(45)},
				Point{Unit(42), Unit(48)},
				Point{Unit(51), Unit(51)},
				Point{Unit(57), Unit(51)},
				Point{Unit(66), Unit(48)},
				Point{Unit(72), Unit(42)},
				Point{Unit(75), Unit(33)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(69), Unit(12)},
				Point{Unit(63), Unit(9)},
			},
		},
		W: Unit(60),
	},
	'4': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(39)},
				Point{Unit(54), Unit(9)},
				Point{Unit(54), Unit(54)},
			},
			Path{
				Point{Unit(12), Unit(39)},
				Point{Unit(75), Unit(39)},
			},
		},
		W: Unit(60),
	},
	'5': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(45)},
				Point{Unit(12), Unit(15)},
				Point{Unit(39), Unit(12)},
				Point{Unit(36), Unit(15)},
				Point{Unit(33), Unit(24)},
				Point{Unit(33), Unit(33)},
				Point{Unit(36), Unit(42)},
				Point{Unit(42), Unit(48)},
				Point{Unit(51), Unit(51)},
				Point{Unit(57), Unit(51)},
				Point{Unit(66), Unit(48)},
				Point{Unit(72), Unit(42)},
				Point{Unit(75), Unit(33)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(69), Unit(12)},
				Point{Unit(63), Unit(9)},
			},
		},
		W: Unit(60),
	},
	'6': Glyph{
		S: Set{
			Path{
				Point{Unit(21), Unit(48)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(36)},
				Point{Unit(12), Unit(30)},
				Point{Unit(15), Unit(21)},
				Point{Unit(24), Unit(15)},
				Point{Unit(39), Unit(12)},
				Point{Unit(54), Unit(12)},
				Point{Unit(66), Unit(15)},
				Point{Unit(72), Unit(21)},
				Point{Unit(75), Unit(30)},
				Point{Unit(75), Unit(33)},
				Point{Unit(72), Unit(42)},
				Point{Unit(66), Unit(48)},
				Point{Unit(57), Unit(51)},
				Point{Unit(54), Unit(51)},
				Point{Unit(45), Unit(48)},
				Point{Unit(39), Unit(42)},
				Point{Unit(36), Unit(33)},
				Point{Unit(36), Unit(30)},
				Point{Unit(39), Unit(21)},
				Point{Unit(45), Unit(15)},
				Point{Unit(54), Unit(12)},
			},
		},
		W: Unit(60),
	},
	'7': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(51)},
				Point{Unit(75), Unit(21)},
			},
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(12), Unit(51)},
			},
		},
		W: Unit(60),
	},
	'8': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(24)},
				Point{Unit(15), Unit(15)},
				Point{Unit(21), Unit(12)},
				Point{Unit(27), Unit(12)},
				Point{Unit(33), Unit(15)},
				Point{Unit(36), Unit(21)},
				Point{Unit(39), Unit(33)},
				Point{Unit(42), Unit(42)},
				Point{Unit(48), Unit(48)},
				Point{Unit(54), Unit(51)},
				Point{Unit(63), Unit(51)},
				Point{Unit(69), Unit(48)},
				Point{Unit(72), Unit(45)},
				Point{Unit(75), Unit(36)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(69), Unit(12)},
				Point{Unit(63), Unit(9)},
				Point{Unit(54), Unit(9)},
				Point{Unit(48), Unit(12)},
				Point{Unit(42), Unit(18)},
				Point{Unit(39), Unit(27)},
				Point{Unit(36), Unit(39)},
				Point{Unit(33), Unit(45)},
				Point{Unit(27), Unit(48)},
				Point{Unit(21), Unit(48)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(36)},
				Point{Unit(12), Unit(24)},
			},
		},
		W: Unit(60),
	},
	'9': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(48)},
				Point{Unit(42), Unit(45)},
				Point{Unit(48), Unit(39)},
				Point{Unit(51), Unit(30)},
				Point{Unit(51), Unit(27)},
				Point{Unit(48), Unit(18)},
				Point{Unit(42), Unit(12)},
				Point{Unit(33), Unit(9)},
				Point{Unit(30), Unit(9)},
				Point{Unit(21), Unit(12)},
				Point{Unit(15), Unit(18)},
				Point{Unit(12), Unit(27)},
				Point{Unit(12), Unit(30)},
				Point{Unit(15), Unit(39)},
				Point{Unit(21), Unit(45)},
				Point{Unit(33), Unit(48)},
				Point{Unit(48), Unit(48)},
				Point{Unit(63), Unit(45)},
				Point{Unit(72), Unit(39)},
				Point{Unit(75), Unit(30)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(66), Unit(12)},
			},
		},
		W: Unit(60),
	},
	':': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(15)},
				Point{Unit(36), Unit(12)},
				Point{Unit(39), Unit(15)},
				Point{Unit(36), Unit(18)},
				Point{Unit(33), Unit(15)},
			},
			Path{
				Point{Unit(69), Unit(15)},
				Point{Unit(72), Unit(12)},
				Point{Unit(75), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(69), Unit(15)},
			},
		},
		W: Unit(30),
	},
	';': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(15)},
				Point{Unit(36), Unit(12)},
				Point{Unit(39), Unit(15)},
				Point{Unit(36), Unit(18)},
				Point{Unit(33), Unit(15)},
			},
			Path{
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(15)},
				Point{Unit(72), Unit(12)},
				Point{Unit(69), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(78), Unit(18)},
				Point{Unit(84), Unit(15)},
				Point{Unit(87), Unit(12)},
			},
		},
		W: Unit(30),
	},
	'<': Glyph{
		S: Set{
			Path{
				Point{Unit(21), Unit(60)},
				Point{Unit(48), Unit(12)},
				Point{Unit(75), Unit(60)},
			},
		},
		W: Unit(72),
	},
	'=': Glyph{
		S: Set{
			Path{
				Point{Unit(39), Unit(12)},
				Point{Unit(39), Unit(66)},
			},
			Path{
				Point{Unit(57), Unit(12)},
				Point{Unit(57), Unit(66)},
			},
		},
		W: Unit(78),
	},
	'>': Glyph{
		S: Set{
			Path{
				Point{Unit(21), Unit(12)},
				Point{Unit(48), Unit(60)},
				Point{Unit(75), Unit(12)},
			},
		},
		W: Unit(72),
	},
	'?': Glyph{
		S: Set{
			Path{
				Point{Unit(27), Unit(9)},
				Point{Unit(24), Unit(9)},
				Point{Unit(18), Unit(12)},
				Point{Unit(15), Unit(15)},
				Point{Unit(12), Unit(21)},
				Point{Unit(12), Unit(33)},
				Point{Unit(15), Unit(39)},
				Point{Unit(18), Unit(42)},
				Point{Unit(24), Unit(45)},
				Point{Unit(30), Unit(45)},
				Point{Unit(36), Unit(42)},
				Point{Unit(39), Unit(39)},
				Point{Unit(45), Unit(27)},
				Point{Unit(54), Unit(27)},
			},
			Path{
				Point{Unit(69), Unit(27)},
				Point{Unit(72), Unit(24)},
				Point{Unit(75), Unit(27)},
				Point{Unit(72), Unit(30)},
				Point{Unit(69), Unit(27)},
			},
		},
		W: Unit(54),
	},
	'@': Glyph{
		S: Set{
			Path{
				Point{Unit(36), Unit(55.5)},
				Point{Unit(30), Unit(52.5)},
				Point{Unit(27), Unit(46.5)},
				Point{Unit(27), Unit(37.5)},
				Point{Unit(30), Unit(31.5)},
				Point{Unit(33), Unit(28.5)},
				Point{Unit(42), Unit(25.5)},
				Point{Unit(51), Unit(25.5)},
				Point{Unit(57), Unit(28.5)},
				Point{Unit(60), Unit(34.5)},
				Point{Unit(60), Unit(43.5)},
				Point{Unit(57), Unit(49.5)},
				Point{Unit(51), Unit(52.5)},
			},
			Path{
				Point{Unit(27), Unit(37.5)},
				Point{Unit(33), Unit(31.5)},
				Point{Unit(42), Unit(28.5)},
				Point{Unit(51), Unit(28.5)},
				Point{Unit(57), Unit(31.5)},
				Point{Unit(60), Unit(34.5)},
			},
			Path{
				Point{Unit(27), Unit(55.5)},
				Point{Unit(51), Unit(52.5)},
				Point{Unit(57), Unit(52.5)},
				Point{Unit(60), Unit(58.5)},
				Point{Unit(60), Unit(64.5)},
				Point{Unit(54), Unit(70.5)},
				Point{Unit(45), Unit(73.5)},
				Point{Unit(39), Unit(73.5)},
				Point{Unit(30), Unit(70.5)},
				Point{Unit(24), Unit(67.5)},
				Point{Unit(18), Unit(61.5)},
				Point{Unit(15), Unit(55.5)},
				Point{Unit(12), Unit(46.5)},
				Point{Unit(12), Unit(37.5)},
				Point{Unit(15), Unit(28.5)},
				Point{Unit(18), Unit(22.5)},
				Point{Unit(24), Unit(16.5)},
				Point{Unit(30), Unit(13.5)},
				Point{Unit(39), Unit(10.5)},
				Point{Unit(48), Unit(10.5)},
				Point{Unit(57), Unit(13.5)},
				Point{Unit(63), Unit(16.5)},
				Point{Unit(69), Unit(22.5)},
				Point{Unit(72), Unit(28.5)},
				Point{Unit(75), Unit(37.5)},
				Point{Unit(75), Unit(46.5)},
				Point{Unit(72), Unit(55.5)},
				Point{Unit(69), Unit(61.5)},
				Point{Unit(66), Unit(64.5)},
			},
			Path{
				Point{Unit(27), Unit(58.5)},
				Point{Unit(51), Unit(55.5)},
				Point{Unit(57), Unit(55.5)},
				Point{Unit(60), Unit(58.5)},
			},
		},
		W: Unit(81),
	},
	'A': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(27)},
				Point{Unit(75), Unit(3)},
			},
			Path{
				Point{Unit(12), Unit(27)},
				Point{Unit(75), Unit(51)},
			},
			Path{
				Point{Unit(54), Unit(12)},
				Point{Unit(54), Unit(42)},
			},
		},
		W: Unit(54),
	},
	'B': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(12), Unit(37.5)},
				Point{Unit(15), Unit(46.5)},
				Point{Unit(18), Unit(49.5)},
				Point{Unit(24), Unit(52.5)},
				Point{Unit(30), Unit(52.5)},
				Point{Unit(36), Unit(49.5)},
				Point{Unit(39), Unit(46.5)},
				Point{Unit(42), Unit(37.5)},
			},
			Path{
				Point{Unit(42), Unit(10.5)},
				Point{Unit(42), Unit(37.5)},
				Point{Unit(45), Unit(46.5)},
				Point{Unit(48), Unit(49.5)},
				Point{Unit(54), Unit(52.5)},
				Point{Unit(63), Unit(52.5)},
				Point{Unit(69), Unit(49.5)},
				Point{Unit(72), Unit(46.5)},
				Point{Unit(75), Unit(37.5)},
				Point{Unit(75), Unit(10.5)},
			},
		},
		W: Unit(63),
	},
	'C': Glyph{
		S: Set{
			Path{
				Point{Unit(27), Unit(55.5)},
				Point{Unit(21), Unit(52.5)},
				Point{Unit(15), Unit(46.5)},
				Point{Unit(12), Unit(40.5)},
				Point{Unit(12), Unit(28.5)},
				Point{Unit(15), Unit(22.5)},
				Point{Unit(21), Unit(16.5)},
				Point{Unit(27), Unit(13.5)},
				Point{Unit(36), Unit(10.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(60), Unit(13.5)},
				Point{Unit(66), Unit(16.5)},
				Point{Unit(72), Unit(22.5)},
				Point{Unit(75), Unit(28.5)},
				Point{Unit(75), Unit(40.5)},
				Point{Unit(72), Unit(46.5)},
				Point{Unit(66), Unit(52.5)},
				Point{Unit(60), Unit(55.5)},
			},
		},
		W: Unit(63),
	},
	'D': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(12), Unit(31.5)},
				Point{Unit(15), Unit(40.5)},
				Point{Unit(21), Unit(46.5)},
				Point{Unit(27), Unit(49.5)},
				Point{Unit(36), Unit(52.5)},
				Point{Unit(51), Unit(52.5)},
				Point{Unit(60), Unit(49.5)},
				Point{Unit(66), Unit(46.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(75), Unit(31.5)},
				Point{Unit(75), Unit(10.5)},
			},
		},
		W: Unit(63),
	},
	'E': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(12), Unit(49.5)},
			},
			Path{
				Point{Unit(42), Unit(10.5)},
				Point{Unit(42), Unit(34.5)},
			},
			Path{
				Point{Unit(75), Unit(10.5)},
				Point{Unit(75), Unit(49.5)},
			},
		},
		W: Unit(57),
	},
	'F': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(75), Unit(9)},
			},
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(12), Unit(48)},
			},
			Path{
				Point{Unit(42), Unit(9)},
				Point{Unit(42), Unit(33)},
			},
		},
		W: Unit(54),
	},
	'G': Glyph{
		S: Set{
			Path{
				Point{Unit(27), Unit(55.5)},
				Point{Unit(21), Unit(52.5)},
				Point{Unit(15), Unit(46.5)},
				Point{Unit(12), Unit(40.5)},
				Point{Unit(12), Unit(28.5)},
				Point{Unit(15), Unit(22.5)},
				Point{Unit(21), Unit(16.5)},
				Point{Unit(27), Unit(13.5)},
				Point{Unit(36), Unit(10.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(60), Unit(13.5)},
				Point{Unit(66), Unit(16.5)},
				Point{Unit(72), Unit(22.5)},
				Point{Unit(75), Unit(28.5)},
				Point{Unit(75), Unit(40.5)},
				Point{Unit(72), Unit(46.5)},
				Point{Unit(66), Unit(52.5)},
				Point{Unit(60), Unit(55.5)},
				Point{Unit(51), Unit(55.5)},
			},
			Path{
				Point{Unit(51), Unit(40.5)},
				Point{Unit(51), Unit(55.5)},
			},
		},
		W: Unit(63),
	},
	'H': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
			Path{
				Point{Unit(12), Unit(54)},
				Point{Unit(75), Unit(54)},
			},
			Path{
				Point{Unit(42), Unit(12)},
				Point{Unit(42), Unit(54)},
			},
		},
		W: Unit(66),
	},
	'I': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
		},
		W: Unit(24),
	},
	'J': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(36)},
				Point{Unit(60), Unit(36)},
				Point{Unit(69), Unit(33)},
				Point{Unit(72), Unit(30)},
				Point{Unit(75), Unit(24)},
				Point{Unit(75), Unit(18)},
				Point{Unit(72), Unit(12)},
				Point{Unit(69), Unit(9)},
				Point{Unit(60), Unit(6)},
				Point{Unit(54), Unit(6)},
			},
		},
		W: Unit(48),
	},
	'K': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(52.5)},
				Point{Unit(54), Unit(10.5)},
			},
			Path{
				Point{Unit(39), Unit(25.5)},
				Point{Unit(75), Unit(52.5)},
			},
		},
		W: Unit(63),
	},
	'L': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(7.5)},
				Point{Unit(75), Unit(7.5)},
			},
			Path{
				Point{Unit(75), Unit(7.5)},
				Point{Unit(75), Unit(43.5)},
			},
		},
		W: Unit(51),
	},
	'M': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(36)},
			},
			Path{
				Point{Unit(12), Unit(60)},
				Point{Unit(75), Unit(36)},
			},
			Path{
				Point{Unit(12), Unit(60)},
				Point{Unit(75), Unit(60)},
			},
		},
		W: Unit(72),
	},
	'N': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(54)},
			},
			Path{
				Point{Unit(12), Unit(54)},
				Point{Unit(75), Unit(54)},
			},
		},
		W: Unit(66),
	},
	'O': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(27)},
				Point{Unit(15), Unit(21)},
				Point{Unit(21), Unit(15)},
				Point{Unit(27), Unit(12)},
				Point{Unit(36), Unit(9)},
				Point{Unit(51), Unit(9)},
				Point{Unit(60), Unit(12)},
				Point{Unit(66), Unit(15)},
				Point{Unit(72), Unit(21)},
				Point{Unit(75), Unit(27)},
				Point{Unit(75), Unit(39)},
				Point{Unit(72), Unit(45)},
				Point{Unit(66), Unit(51)},
				Point{Unit(60), Unit(54)},
				Point{Unit(51), Unit(57)},
				Point{Unit(36), Unit(57)},
				Point{Unit(27), Unit(54)},
				Point{Unit(21), Unit(51)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(39)},
				Point{Unit(12), Unit(27)},
			},
		},
		W: Unit(66),
	},
	'P': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(12), Unit(37.5)},
				Point{Unit(15), Unit(46.5)},
				Point{Unit(18), Unit(49.5)},
				Point{Unit(24), Unit(52.5)},
				Point{Unit(33), Unit(52.5)},
				Point{Unit(39), Unit(49.5)},
				Point{Unit(42), Unit(46.5)},
				Point{Unit(45), Unit(37.5)},
				Point{Unit(45), Unit(10.5)},
			},
		},
		W: Unit(63),
	},
	'Q': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(27)},
				Point{Unit(15), Unit(21)},
				Point{Unit(21), Unit(15)},
				Point{Unit(27), Unit(12)},
				Point{Unit(36), Unit(9)},
				Point{Unit(51), Unit(9)},
				Point{Unit(60), Unit(12)},
				Point{Unit(66), Unit(15)},
				Point{Unit(72), Unit(21)},
				Point{Unit(75), Unit(27)},
				Point{Unit(75), Unit(39)},
				Point{Unit(72), Unit(45)},
				Point{Unit(66), Unit(51)},
				Point{Unit(60), Unit(54)},
				Point{Unit(51), Unit(57)},
				Point{Unit(36), Unit(57)},
				Point{Unit(27), Unit(54)},
				Point{Unit(21), Unit(51)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(39)},
				Point{Unit(12), Unit(27)},
			},
			Path{
				Point{Unit(63), Unit(36)},
				Point{Unit(81), Unit(54)},
			},
		},
		W: Unit(66),
	},
	'R': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(12), Unit(37.5)},
				Point{Unit(15), Unit(46.5)},
				Point{Unit(18), Unit(49.5)},
				Point{Unit(24), Unit(52.5)},
				Point{Unit(30), Unit(52.5)},
				Point{Unit(36), Unit(49.5)},
				Point{Unit(39), Unit(46.5)},
				Point{Unit(42), Unit(37.5)},
				Point{Unit(42), Unit(10.5)},
			},
			Path{
				Point{Unit(42), Unit(31.5)},
				Point{Unit(75), Unit(52.5)},
			},
		},
		W: Unit(63),
	},
	'S': Glyph{
		S: Set{
			Path{
				Point{Unit(21), Unit(51)},
				Point{Unit(15), Unit(45)},
				Point{Unit(12), Unit(36)},
				Point{Unit(12), Unit(24)},
				Point{Unit(15), Unit(15)},
				Point{Unit(21), Unit(9)},
				Point{Unit(27), Unit(9)},
				Point{Unit(33), Unit(12)},
				Point{Unit(36), Unit(15)},
				Point{Unit(39), Unit(21)},
				Point{Unit(45), Unit(39)},
				Point{Unit(48), Unit(45)},
				Point{Unit(51), Unit(48)},
				Point{Unit(57), Unit(51)},
				Point{Unit(66), Unit(51)},
				Point{Unit(72), Unit(45)},
				Point{Unit(75), Unit(36)},
				Point{Unit(75), Unit(24)},
				Point{Unit(72), Unit(15)},
				Point{Unit(66), Unit(9)},
			},
		},
		W: Unit(60),
	},
	'T': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(24)},
				Point{Unit(75), Unit(24)},
			},
			Path{
				Point{Unit(12), Unit(3)},
				Point{Unit(12), Unit(45)},
			},
		},
		W: Unit(48),
	},
	'U': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(57), Unit(12)},
				Point{Unit(66), Unit(15)},
				Point{Unit(72), Unit(21)},
				Point{Unit(75), Unit(30)},
				Point{Unit(75), Unit(36)},
				Point{Unit(72), Unit(45)},
				Point{Unit(66), Unit(51)},
				Point{Unit(57), Unit(54)},
				Point{Unit(12), Unit(54)},
			},
		},
		W: Unit(66),
	},
	'V': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(3)},
				Point{Unit(75), Unit(27)},
			},
			Path{
				Point{Unit(12), Unit(51)},
				Point{Unit(75), Unit(27)},
			},
		},
		W: Unit(54),
	},
	'W': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(6)},
				Point{Unit(75), Unit(21)},
			},
			Path{
				Point{Unit(12), Unit(36)},
				Point{Unit(75), Unit(21)},
			},
			Path{
				Point{Unit(12), Unit(36)},
				Point{Unit(75), Unit(51)},
			},
			Path{
				Point{Unit(12), Unit(66)},
				Point{Unit(75), Unit(51)},
			},
		},
		W: Unit(72),
	},
	'X': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(75), Unit(51)},
			},
			Path{
				Point{Unit(12), Unit(51)},
				Point{Unit(75), Unit(9)},
			},
		},
		W: Unit(60),
	},
	'Y': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(3)},
				Point{Unit(42), Unit(27)},
				Point{Unit(75), Unit(27)},
			},
			Path{
				Point{Unit(12), Unit(51)},
				Point{Unit(42), Unit(27)},
			},
		},
		W: Unit(54),
	},
	'Z': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(51)},
				Point{Unit(75), Unit(9)},
			},
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(12), Unit(51)},
			},
			Path{
				Point{Unit(75), Unit(9)},
				Point{Unit(75), Unit(51)},
			},
		},
		W: Unit(60),
	},
	'[': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(12)},
				Point{Unit(96), Unit(12)},
			},
			Path{
				Point{Unit(0), Unit(15)},
				Point{Unit(96), Unit(15)},
			},
			Path{
				Point{Unit(0), Unit(12)},
				Point{Unit(0), Unit(33)},
			},
			Path{
				Point{Unit(96), Unit(12)},
				Point{Unit(96), Unit(33)},
			},
		},
		W: Unit(42),
	},
	'\\': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(0)},
				Point{Unit(84), Unit(42)},
			},
		},
		W: Unit(42),
	},
	']': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(27)},
				Point{Unit(96), Unit(27)},
			},
			Path{
				Point{Unit(0), Unit(30)},
				Point{Unit(96), Unit(30)},
			},
			Path{
				Point{Unit(0), Unit(9)},
				Point{Unit(0), Unit(30)},
			},
			Path{
				Point{Unit(96), Unit(9)},
				Point{Unit(96), Unit(30)},
			},
		},
		W: Unit(42),
	},
	'^': Glyph{
		S: Set{
			Path{
				Point{Unit(30), Unit(18)},
				Point{Unit(21), Unit(24)},
				Point{Unit(30), Unit(30)},
			},
			Path{
				Point{Unit(39), Unit(9)},
				Point{Unit(24), Unit(24)},
				Point{Unit(39), Unit(39)},
			},
			Path{
				Point{Unit(24), Unit(24)},
				Point{Unit(75), Unit(24)},
			},
		},
		W: Unit(48),
	},
	'_': Glyph{
		S: Set{
			Path{
				Point{Unit(81), Unit(0)},
				Point{Unit(81), Unit(48)},
			},
		},
		W: Unit(48),
	},
	'`': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(18)},
				Point{Unit(15), Unit(15)},
				Point{Unit(21), Unit(12)},
				Point{Unit(27), Unit(12)},
				Point{Unit(30), Unit(15)},
				Point{Unit(27), Unit(18)},
				Point{Unit(24), Unit(15)},
			},
		},
		W: Unit(30),
	},
	'a': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(46.5)},
				Point{Unit(75), Unit(46.5)},
			},
			Path{
				Point{Unit(42), Unit(46.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(34.5)},
				Point{Unit(33), Unit(25.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(57), Unit(10.5)},
				Point{Unit(66), Unit(13.5)},
				Point{Unit(72), Unit(19.5)},
				Point{Unit(75), Unit(25.5)},
				Point{Unit(75), Unit(34.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(66), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'b': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(42), Unit(10.5)},
				Point{Unit(36), Unit(16.5)},
				Point{Unit(33), Unit(22.5)},
				Point{Unit(33), Unit(31.5)},
				Point{Unit(36), Unit(37.5)},
				Point{Unit(42), Unit(43.5)},
				Point{Unit(51), Unit(46.5)},
				Point{Unit(57), Unit(46.5)},
				Point{Unit(66), Unit(43.5)},
				Point{Unit(72), Unit(37.5)},
				Point{Unit(75), Unit(31.5)},
				Point{Unit(75), Unit(22.5)},
				Point{Unit(72), Unit(16.5)},
				Point{Unit(66), Unit(10.5)},
			},
		},
		W: Unit(57),
	},
	'c': Glyph{
		S: Set{
			Path{
				Point{Unit(42), Unit(45)},
				Point{Unit(36), Unit(39)},
				Point{Unit(33), Unit(33)},
				Point{Unit(33), Unit(24)},
				Point{Unit(36), Unit(18)},
				Point{Unit(42), Unit(12)},
				Point{Unit(51), Unit(9)},
				Point{Unit(57), Unit(9)},
				Point{Unit(66), Unit(12)},
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(24)},
				Point{Unit(75), Unit(33)},
				Point{Unit(72), Unit(39)},
				Point{Unit(66), Unit(45)},
			},
		},
		W: Unit(54),
	},
	'd': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(46.5)},
				Point{Unit(75), Unit(46.5)},
			},
			Path{
				Point{Unit(42), Unit(46.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(34.5)},
				Point{Unit(33), Unit(25.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(57), Unit(10.5)},
				Point{Unit(66), Unit(13.5)},
				Point{Unit(72), Unit(19.5)},
				Point{Unit(75), Unit(25.5)},
				Point{Unit(75), Unit(34.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(66), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'e': Glyph{
		S: Set{
			Path{
				Point{Unit(51), Unit(9)},
				Point{Unit(51), Unit(45)},
				Point{Unit(45), Unit(45)},
				Point{Unit(39), Unit(42)},
				Point{Unit(36), Unit(39)},
				Point{Unit(33), Unit(33)},
				Point{Unit(33), Unit(24)},
				Point{Unit(36), Unit(18)},
				Point{Unit(42), Unit(12)},
				Point{Unit(51), Unit(9)},
				Point{Unit(57), Unit(9)},
				Point{Unit(66), Unit(12)},
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(24)},
				Point{Unit(75), Unit(33)},
				Point{Unit(72), Unit(39)},
				Point{Unit(66), Unit(45)},
			},
		},
		W: Unit(54),
	},
	'f': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(33)},
				Point{Unit(12), Unit(27)},
				Point{Unit(15), Unit(21)},
				Point{Unit(24), Unit(18)},
				Point{Unit(75), Unit(18)},
			},
			Path{
				Point{Unit(33), Unit(9)},
				Point{Unit(33), Unit(30)},
			},
		},
		W: Unit(36),
	},
	'g': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(46.5)},
				Point{Unit(81), Unit(46.5)},
				Point{Unit(90), Unit(43.5)},
				Point{Unit(93), Unit(40.5)},
				Point{Unit(96), Unit(34.5)},
				Point{Unit(96), Unit(25.5)},
				Point{Unit(93), Unit(19.5)},
			},
			Path{
				Point{Unit(42), Unit(46.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(34.5)},
				Point{Unit(33), Unit(25.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(57), Unit(10.5)},
				Point{Unit(66), Unit(13.5)},
				Point{Unit(72), Unit(19.5)},
				Point{Unit(75), Unit(25.5)},
				Point{Unit(75), Unit(34.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(66), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'h': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(13.5)},
				Point{Unit(75), Unit(13.5)},
			},
			Path{
				Point{Unit(45), Unit(13.5)},
				Point{Unit(36), Unit(22.5)},
				Point{Unit(33), Unit(28.5)},
				Point{Unit(33), Unit(37.5)},
				Point{Unit(36), Unit(43.5)},
				Point{Unit(45), Unit(46.5)},
				Point{Unit(75), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'i': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(9)},
				Point{Unit(15), Unit(12)},
				Point{Unit(12), Unit(15)},
				Point{Unit(9), Unit(12)},
				Point{Unit(12), Unit(9)},
			},
			Path{
				Point{Unit(33), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
		},
		W: Unit(24),
	},
	'j': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(15)},
				Point{Unit(15), Unit(18)},
				Point{Unit(12), Unit(21)},
				Point{Unit(9), Unit(18)},
				Point{Unit(12), Unit(15)},
			},
			Path{
				Point{Unit(33), Unit(18)},
				Point{Unit(84), Unit(18)},
				Point{Unit(93), Unit(15)},
				Point{Unit(96), Unit(9)},
				Point{Unit(96), Unit(3)},
			},
		},
		W: Unit(30),
	},
	'k': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(33), Unit(40.5)},
				Point{Unit(63), Unit(10.5)},
			},
			Path{
				Point{Unit(51), Unit(22.5)},
				Point{Unit(75), Unit(43.5)},
			},
		},
		W: Unit(51),
	},
	'l': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
		},
		W: Unit(24),
	},
	'm': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(12)},
				Point{Unit(75), Unit(12)},
			},
			Path{
				Point{Unit(45), Unit(12)},
				Point{Unit(36), Unit(21)},
				Point{Unit(33), Unit(27)},
				Point{Unit(33), Unit(36)},
				Point{Unit(36), Unit(42)},
				Point{Unit(45), Unit(45)},
				Point{Unit(75), Unit(45)},
			},
			Path{
				Point{Unit(45), Unit(45)},
				Point{Unit(36), Unit(54)},
				Point{Unit(33), Unit(60)},
				Point{Unit(33), Unit(69)},
				Point{Unit(36), Unit(75)},
				Point{Unit(45), Unit(78)},
				Point{Unit(75), Unit(78)},
			},
		},
		W: Unit(90),
	},
	'n': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(13.5)},
				Point{Unit(75), Unit(13.5)},
			},
			Path{
				Point{Unit(45), Unit(13.5)},
				Point{Unit(36), Unit(22.5)},
				Point{Unit(33), Unit(28.5)},
				Point{Unit(33), Unit(37.5)},
				Point{Unit(36), Unit(43.5)},
				Point{Unit(45), Unit(46.5)},
				Point{Unit(75), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'o': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(25.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(57), Unit(10.5)},
				Point{Unit(66), Unit(13.5)},
				Point{Unit(72), Unit(19.5)},
				Point{Unit(75), Unit(25.5)},
				Point{Unit(75), Unit(34.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(66), Unit(46.5)},
				Point{Unit(57), Unit(49.5)},
				Point{Unit(51), Unit(49.5)},
				Point{Unit(42), Unit(46.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(34.5)},
				Point{Unit(33), Unit(25.5)},
			},
		},
		W: Unit(57),
	},
	'p': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(10.5)},
				Point{Unit(96), Unit(10.5)},
			},
			Path{
				Point{Unit(42), Unit(10.5)},
				Point{Unit(36), Unit(16.5)},
				Point{Unit(33), Unit(22.5)},
				Point{Unit(33), Unit(31.5)},
				Point{Unit(36), Unit(37.5)},
				Point{Unit(42), Unit(43.5)},
				Point{Unit(51), Unit(46.5)},
				Point{Unit(57), Unit(46.5)},
				Point{Unit(66), Unit(43.5)},
				Point{Unit(72), Unit(37.5)},
				Point{Unit(75), Unit(31.5)},
				Point{Unit(75), Unit(22.5)},
				Point{Unit(72), Unit(16.5)},
				Point{Unit(66), Unit(10.5)},
			},
		},
		W: Unit(57),
	},
	'q': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(46.5)},
				Point{Unit(96), Unit(46.5)},
			},
			Path{
				Point{Unit(42), Unit(46.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(34.5)},
				Point{Unit(33), Unit(25.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(51), Unit(10.5)},
				Point{Unit(57), Unit(10.5)},
				Point{Unit(66), Unit(13.5)},
				Point{Unit(72), Unit(19.5)},
				Point{Unit(75), Unit(25.5)},
				Point{Unit(75), Unit(34.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(66), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'r': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(10.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(51), Unit(10.5)},
				Point{Unit(42), Unit(13.5)},
				Point{Unit(36), Unit(19.5)},
				Point{Unit(33), Unit(25.5)},
				Point{Unit(33), Unit(34.5)},
			},
		},
		W: Unit(39),
	},
	's': Glyph{
		S: Set{
			Path{
				Point{Unit(42), Unit(43.5)},
				Point{Unit(36), Unit(40.5)},
				Point{Unit(33), Unit(31.5)},
				Point{Unit(33), Unit(22.5)},
				Point{Unit(36), Unit(13.5)},
				Point{Unit(42), Unit(10.5)},
				Point{Unit(48), Unit(13.5)},
				Point{Unit(51), Unit(19.5)},
				Point{Unit(54), Unit(34.5)},
				Point{Unit(57), Unit(40.5)},
				Point{Unit(63), Unit(43.5)},
				Point{Unit(66), Unit(43.5)},
				Point{Unit(72), Unit(40.5)},
				Point{Unit(75), Unit(31.5)},
				Point{Unit(75), Unit(22.5)},
				Point{Unit(72), Unit(13.5)},
				Point{Unit(66), Unit(10.5)},
			},
		},
		W: Unit(51),
	},
	't': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(18)},
				Point{Unit(63), Unit(18)},
				Point{Unit(72), Unit(21)},
				Point{Unit(75), Unit(27)},
				Point{Unit(75), Unit(33)},
			},
			Path{
				Point{Unit(33), Unit(9)},
				Point{Unit(33), Unit(30)},
			},
		},
		W: Unit(36),
	},
	'u': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(13.5)},
				Point{Unit(63), Unit(13.5)},
				Point{Unit(72), Unit(16.5)},
				Point{Unit(75), Unit(22.5)},
				Point{Unit(75), Unit(31.5)},
				Point{Unit(72), Unit(37.5)},
				Point{Unit(63), Unit(46.5)},
			},
			Path{
				Point{Unit(33), Unit(46.5)},
				Point{Unit(75), Unit(46.5)},
			},
		},
		W: Unit(57),
	},
	'v': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(6)},
				Point{Unit(75), Unit(24)},
			},
			Path{
				Point{Unit(33), Unit(42)},
				Point{Unit(75), Unit(24)},
			},
		},
		W: Unit(48),
	},
	'w': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(9)},
				Point{Unit(75), Unit(21)},
			},
			Path{
				Point{Unit(33), Unit(33)},
				Point{Unit(75), Unit(21)},
			},
			Path{
				Point{Unit(33), Unit(33)},
				Point{Unit(75), Unit(45)},
			},
			Path{
				Point{Unit(33), Unit(57)},
				Point{Unit(75), Unit(45)},
			},
		},
		W: Unit(66),
	},
	'x': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(10.5)},
				Point{Unit(75), Unit(43.5)},
			},
			Path{
				Point{Unit(33), Unit(43.5)},
				Point{Unit(75), Unit(10.5)},
			},
		},
		W: Unit(51),
	},
	'y': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(6)},
				Point{Unit(75), Unit(24)},
			},
			Path{
				Point{Unit(33), Unit(42)},
				Point{Unit(75), Unit(24)},
				Point{Unit(87), Unit(18)},
				Point{Unit(93), Unit(12)},
				Point{Unit(96), Unit(6)},
				Point{Unit(96), Unit(3)},
			},
		},
		W: Unit(48),
	},
	'z': Glyph{
		S: Set{
			Path{
				Point{Unit(33), Unit(43.5)},
				Point{Unit(75), Unit(10.5)},
			},
			Path{
				Point{Unit(33), Unit(10.5)},
				Point{Unit(33), Unit(43.5)},
			},
			Path{
				Point{Unit(75), Unit(10.5)},
				Point{Unit(75), Unit(43.5)},
			},
		},
		W: Unit(51),
	},
	'{': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(27)},
				Point{Unit(3), Unit(21)},
				Point{Unit(6), Unit(18)},
				Point{Unit(12), Unit(15)},
				Point{Unit(18), Unit(15)},
				Point{Unit(24), Unit(18)},
				Point{Unit(27), Unit(21)},
				Point{Unit(33), Unit(24)},
				Point{Unit(39), Unit(24)},
				Point{Unit(45), Unit(18)},
			},
			Path{
				Point{Unit(3), Unit(21)},
				Point{Unit(9), Unit(18)},
				Point{Unit(15), Unit(18)},
				Point{Unit(21), Unit(21)},
				Point{Unit(24), Unit(24)},
				Point{Unit(30), Unit(27)},
				Point{Unit(36), Unit(27)},
				Point{Unit(42), Unit(24)},
				Point{Unit(48), Unit(12)},
				Point{Unit(54), Unit(24)},
				Point{Unit(60), Unit(27)},
				Point{Unit(66), Unit(27)},
				Point{Unit(72), Unit(24)},
				Point{Unit(75), Unit(21)},
				Point{Unit(81), Unit(18)},
				Point{Unit(87), Unit(18)},
				Point{Unit(93), Unit(21)},
			},
			Path{
				Point{Unit(51), Unit(18)},
				Point{Unit(57), Unit(24)},
				Point{Unit(63), Unit(24)},
				Point{Unit(69), Unit(21)},
				Point{Unit(72), Unit(18)},
				Point{Unit(78), Unit(15)},
				Point{Unit(84), Unit(15)},
				Point{Unit(90), Unit(18)},
				Point{Unit(93), Unit(21)},
				Point{Unit(96), Unit(27)},
			},
		},
		W: Unit(42),
	},
	'|': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(12)},
				Point{Unit(96), Unit(12)},
			},
		},
		W: Unit(24),
	},
	'}': Glyph{
		S: Set{
			Path{
				Point{Unit(0), Unit(15)},
				Point{Unit(3), Unit(21)},
				Point{Unit(6), Unit(24)},
				Point{Unit(12), Unit(27)},
				Point{Unit(18), Unit(27)},
				Point{Unit(24), Unit(24)},
				Point{Unit(27), Unit(21)},
				Point{Unit(33), Unit(18)},
				Point{Unit(39), Unit(18)},
				Point{Unit(45), Unit(24)},
			},
			Path{
				Point{Unit(3), Unit(21)},
				Point{Unit(9), Unit(24)},
				Point{Unit(15), Unit(24)},
				Point{Unit(21), Unit(21)},
				Point{Unit(24), Unit(18)},
				Point{Unit(30), Unit(15)},
				Point{Unit(36), Unit(15)},
				Point{Unit(42), Unit(18)},
				Point{Unit(48), Unit(30)},
				Point{Unit(54), Unit(18)},
				Point{Unit(60), Unit(15)},
				Point{Unit(66), Unit(15)},
				Point{Unit(72), Unit(18)},
				Point{Unit(75), Unit(21)},
				Point{Unit(81), Unit(24)},
				Point{Unit(87), Unit(24)},
				Point{Unit(93), Unit(21)},
			},
			Path{
				Point{Unit(51), Unit(24)},
				Point{Unit(57), Unit(18)},
				Point{Unit(63), Unit(18)},
				Point{Unit(69), Unit(21)},
				Point{Unit(72), Unit(24)},
				Point{Unit(78), Unit(27)},
				Point{Unit(84), Unit(27)},
				Point{Unit(90), Unit(24)},
				Point{Unit(93), Unit(21)},
				Point{Unit(96), Unit(15)},
			},
		},
		W: Unit(42),
	},
	'~': Glyph{
		S: Set{
			Path{
				Point{Unit(57), Unit(9)},
				Point{Unit(51), Unit(9)},
				Point{Unit(42), Unit(12)},
				Point{Unit(39), Unit(18)},
				Point{Unit(39), Unit(24)},
				Point{Unit(42), Unit(30)},
				Point{Unit(51), Unit(42)},
				Point{Unit(54), Unit(48)},
				Point{Unit(54), Unit(54)},
				Point{Unit(51), Unit(60)},
				Point{Unit(45), Unit(63)},
			},
			Path{
				Point{Unit(51), Unit(9)},
				Point{Unit(45), Unit(12)},
				Point{Unit(42), Unit(18)},
				Point{Unit(42), Unit(24)},
				Point{Unit(45), Unit(30)},
				Point{Unit(54), Unit(42)},
				Point{Unit(57), Unit(48)},
				Point{Unit(57), Unit(54)},
				Point{Unit(54), Unit(60)},
				Point{Unit(45), Unit(63)},
				Point{Unit(39), Unit(63)},
			},
		},
		W: Unit(72),
	},
	'\u007f': Glyph{
		S: Set{
			Path{
				Point{Unit(12), Unit(18)},
				Point{Unit(15), Unit(12)},
				Point{Unit(21), Unit(9)},
				Point{Unit(27), Unit(9)},
				Point{Unit(33), Unit(12)},
				Point{Unit(36), Unit(18)},
				Point{Unit(36), Unit(24)},
				Point{Unit(33), Unit(30)},
				Point{Unit(27), Unit(33)},
				Point{Unit(21), Unit(33)},
				Point{Unit(15), Unit(30)},
				Point{Unit(12), Unit(24)},
				Point{Unit(12), Unit(18)},
			},
		},
		W: Unit(42),
	},
}
