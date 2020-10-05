package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/vmihailenco/msgpack/v4"
)

func main() {
	buf := bytes.Buffer{}
	enc := msgpack.NewEncoder(&buf)
	_ = enc.Encode(1)
	//_ = enc.Encode(1)
	fmt.Printf("%s\n", hex.EncodeToString(buf.Bytes()))

	r := bytes.NewReader([]byte{0x01})
	dec := msgpack.NewDecoder(r)
	i, _ := dec.DecodeInt()
	fmt.Printf("%d\n", i)
}
