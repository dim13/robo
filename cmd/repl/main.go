package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	if err := robo.Initialize(dev, 113, robo.Portrait); err != nil {
		log.Fatal("init: ", err)
	}
	robo.A4.Sub(robo.Margin).UpperRight(dev)
	robo.Triple{U: 100, V: 100, W: 100}.Factor(dev)

	go io.Copy(os.Stdout, dev)

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		robo.Send(dev, s.Text())
	}
}
