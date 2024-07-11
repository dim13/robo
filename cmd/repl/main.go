package main

import (
	"bufio"
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
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		robo.Send(dev, s.Text())
	}
}
