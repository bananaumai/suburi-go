package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v4"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

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
	rd, wr := io.Pipe()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, chunk := range chunks {
			size, _ := wr.Write(chunk)
			log.Printf("wrote %d bytes", size)
			time.Sleep(100 * time.Millisecond)
		}
		_ = wr.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dec := msgpack.NewDecoder(rd)

		for i := 0; i < num+1; i++ {
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

type Packet struct {
	StreamName  string
	PacketIndex uint32
	Data        []byte
}

var _ msgpack.CustomDecoder = &Packet{}

func (p *Packet) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&p.StreamName, &p.PacketIndex, &p.Data)
}

func RetrieveDataPacket(
	ctx context.Context,
	homeID uint64,
	smartModuleID string,
	streamName string,
	start int,
	stop int,
	apiKey string,
) ([]Packet, error) {
	reqURL := url.URL{
		Scheme: "https",
		Host: 	"api.tinkermode.com",
	}
	reqURL.Path = fmt.Sprintf("/homes/%d/smartModules/%s/streams/%s/data", homeID, smartModuleID, streamName)
	params := url.Values{}
	params.Add("start", strconv.Itoa(start))
	params.Add("stop", strconv.Itoa(stop))
	reqURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("ModeCloud %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Go's http client automatically follow the redirect and handle chunked response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d %s", resp.StatusCode, resp.Status)
	}

	msgpackDec := msgpack.NewDecoder(resp.Body)
	var ps []Packet
	for {
		var p Packet
		if err := msgpackDec.Decode(&p); err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("failed to decode: %w", err)
			}
			return ps, nil
		}
		ps = append(ps, p)
	}
}
