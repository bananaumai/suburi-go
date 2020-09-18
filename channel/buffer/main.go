package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

type (
	channelBuffer struct {
		buf []byte
		c   chan []byte
	}
)

func newChannelBuffer(cap int) *channelBuffer {
	return &channelBuffer{c: make(chan []byte, cap)}
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
	wg := sync.WaitGroup{}
	buf := newChannelBuffer(100)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ss := []string{"ab", "cd", "ef"}
		for _, s := range ss {
			if _, err := buf.Write([]byte(s)); err != nil {
				log.Fatalf("failed to write buffer: %s", err)
			}
		}
		_ = buf.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			bs := make([]byte, 1)
			_, err := buf.Read(bs)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("finished to read")
					return
				}
				log.Fatalf("failed to read buffer: %s", err)
			}
			log.Printf("read %s", bs)
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
	log.Printf("completed")
}
