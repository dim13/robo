package robo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type Pic []Path

func parseLine(s string) (pa Path) {
	if strings.HasPrefix(s, "line from ") {
		for _, p := range strings.Split(s[10:], " to ") {
			var po Point
			fmt.Sscanf(p, "%v,%v", &po.X, &po.Y)
			po.X = po.X * IN
			po.Y = po.Y * IN
			pa = append(pa, po)
		}
	}
	return pa
}

func parsePic() (pa Pic) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if p := parseLine(scanner.Text()); p != nil {
			pa = append(pa, p)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return pa
}

func (r Robo) DrawPic() {
	//Landscape.Orientation(c)
	for _, p := range parsePic() {
		r.Line(p...)
	}
}
