package main

import (
	"fmt"

	"github.com/dim13/robo"
)

func main() {
	dev := robo.NewDevice()
	defer dev.Close()
	fmt.Println(robo.Version(dev))
}
