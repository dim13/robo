// Decrypt Graphtec files
package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	magic = []byte{0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99}
	key   = []byte{0x32, 0x5d, 0xbc, 0x97, 0xa8, 0xa1, 0x26, 0x08}
	zero  = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	file  = flag.String("file", "Graphtec.enc", "file name")
)

func main() {
	flag.Parse()

	f, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	if !bytes.Equal(f[:8], magic) {
		log.Fatal("bad magic")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	data := f[0x88:]
	mode := cipher.NewCBCDecrypter(block, zero)
	mode.CryptBlocks(data, data)

	size := len(data)
	size -= int(data[size-1:][0])
	fmt.Print(string(data[:size]))
}
