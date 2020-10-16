package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	msg := []byte("message")
	meta := []byte{0x01, 0x02}
	h := hmac.New(sha256.New, append(msg, meta...))
	bs := h.Sum(meta)
	fmt.Println(hex.EncodeToString(bs))
}
