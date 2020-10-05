package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/vmihailenco/msgpack/v4"
)

func main() {
	buf := bytes.Buffer{}
	enc := msgpack.NewEncoder(&buf)
	for _, s := range []string{"foo", "bar", "buzz"} {
		_ = enc.Encode(s)
	}

	wg := sync.WaitGroup{}
	pr, pw := io.Pipe()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			_ = pw.Close()
		}()
		r := bytes.NewReader(buf.Bytes())
		for {
			// split msgpack strings stream into tiny chunks of byte array
			bs := make([]byte, 1)
			n, err := r.Read(bs)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					panic(err)
				}
				return
			}
			_, _ = pw.Write(bs[:n])
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dec := msgpack.NewDecoder(pr)
		for {
			var s string
			if err := dec.Decode(&s); err != nil {
				if !errors.Is(err, io.EOF) {
					panic(err)
				}
				return
			}
			fmt.Println(s)
		}
	}()

	wg.Wait()
}
