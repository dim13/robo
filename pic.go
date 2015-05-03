package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type Page []Path

const scale = 24.5 * 20

func truncate(f float64) float64 {
	return float64(int(f*100)) / 100
}

func parseLine(s string) (pa Path) {
	if strings.HasPrefix(s, "line from ") {
		for _, p := range strings.Split(s[10:], " to ") {
			var po Point
			fmt.Sscanf(p, "%v,%v", &po.Y, &po.X)
			po.X = truncate(5440 - po.X*scale)
			po.Y = truncate(po.Y * scale)
			pa = append(pa, po)
		}
	}
	return pa
}

func parsePage() (pa Page) {
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
