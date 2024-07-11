package main

import (
	"log"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	defer robo.Home(dev)

	robo.Initialize(dev, 113, robo.Portrait)
	robo.A4.Sub(robo.Margin).UpperRight(dev)
	robo.Triple{100, 100, 100}.Factor(dev)

	size := robo.CM * 15
	for i, p := range ClarkY {
		_, _ = i, p
		origin := robo.Point{5, robo.Unit(5 + 10*i)}.Scale(robo.CM)
		p.Scale(size).Add(origin).Line(dev)
		Box.Scale(size).Add(origin).Line(dev)
		for _, m := range Marks {
			m.Scale(size).Add(origin).Line(dev)
		}
		for _, m := range DrillHoles {
			m.Scale(size).Add(origin).Line(dev)
		}
	}

}
