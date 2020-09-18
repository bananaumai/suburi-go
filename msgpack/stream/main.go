package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v4"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"
)

type (
	channelBuffer struct {
		buf []byte
		c   chan []byte
	}
)

func newChannelBuffer(channelCap int, bufferCap int) *channelBuffer {
	return &channelBuffer{c: make(chan []byte, channelCap), buf: make([]byte, 0, bufferCap)}
}

func (b *channelBuffer) Write(p []byte) (n int, err error) {
	b.c <- p
	return len(p), nil
}

func (b *channelBuffer) Read(p []byte) (n int, err error) {
	if len(b.buf) < len(p) {
		for bs := range b.c {
			b.buf = append(b.buf, bs...)
			if len(b.buf) >= len(p) {
				break
			}
		}
	}
	r := bytes.NewReader(b.buf)
	readLen, err := r.Read(p)
	if err != nil {
		return readLen, err
	}
	b.buf = b.buf[readLen:]

	return readLen, nil
}

func (b channelBuffer) Close() error {
	close(b.c)
	return nil
}

func main() {
	const (
		unit = 256
		num  = 2
	)

	srcs := make([][]byte, 0, num)
	bs := make([]byte, 0, unit*num+1024)
	for i := 0; i < num; i++ {
		src := make([]byte, unit)
		rnd := rand.New(rand.NewSource(time.Now().Unix()))
		rnd.Seed(rand.Int63())
		rnd.Read(src)
		srcs = append(srcs, src)
		//printBytes(src)
		bs = append(bs, encode(src)...)
	}

	r := bytes.NewReader(bs)
	var chunks [][]byte
	for {
		buf := make([]byte, unit/2)
		size, err := r.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("failed to read: %s", err)
		}
		chunks = append(chunks, buf[:size])
	}

	wg := sync.WaitGroup{}
	buf := newChannelBuffer(100, 1024)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, chunk := range chunks {
			size, _ := buf.Write(chunk)
			log.Printf("wrote %d bytes", size)
			time.Sleep(100 * time.Millisecond)
		}
		buf.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dec := msgpack.NewDecoder(buf)

		for i := 0; i < num; i++ {
			bs := make([]byte, unit)
			err := dec.Decode(&bs)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("EOF")
					return
				}
				log.Fatalf("failed to decode: %s", err)
			}
			log.Printf("read %d bytes", len(bs))
			printBytes(srcs[i])
			printBytes(bs)
			log.Printf("expected? %t", bytes.Equal(srcs[i], bs))
		}
	}()

	wg.Wait()
}

func encode(bs []byte) []byte {
	buf := bytes.Buffer{}
	enc := msgpack.NewEncoder(&buf)
	if err := enc.Encode(bs); err != nil {
		log.Fatalf("failed to encode: %s", err)
	}
	return buf.Bytes()
}

func printBytes(bs []byte) {
	for i, b := range bs {
		if i != 0 {
			fmt.Printf(" %02X", b)
		} else {
			fmt.Printf("%02X", b)
		}
	}
	fmt.Println("")
}
