package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"dim13.org/robo/gc"
	"github.com/llgcode/ps"
)

var postscriptContent string

func main() {
	src, err := os.OpenFile("tiger.ps", 0, 0)
	if err != nil {
		log.Println("can't find postscript file.")
		return
	}
	defer src.Close()
	bytes, err := ioutil.ReadAll(src)
	postscriptContent = string(bytes)
	if err != nil {
		panic(err)
	}

	GC := gc.NewGraphicContext()

	interpreter := ps.NewInterpreter(GC)
	reader := strings.NewReader(postscriptContent)
	interpreter.Execute(reader)

}
