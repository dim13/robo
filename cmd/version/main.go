package main

import (
	"fmt"
	"log"

	"github.com/dim13/robo"
)

func main() {
	r, err := robo.NewRobo()
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	r.Init()
	r.Wait4Ready()
	ver := r.Version()
	fmt.Println("Version:", ver)
	r.GoHome()
}
