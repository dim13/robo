package main

import (
	"fmt"
	"log"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	fmt.Println(robo.Version(dev))
}
