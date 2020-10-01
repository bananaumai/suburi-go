package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack/v4"
)

type DataPacket struct {
	StreamName  string
	PacketIndex uint32
	Data        []byte
}

type MetadataPacket struct {
	StreamName  string
	PacketIndex uint32
	Content     map[string]interface{}
}

var _ msgpack.Marshaler = MetadataPacket{}

func init() {
	msgpack.RegisterExt(int8(1), MetadataPacket{})
}

func (m MetadataPacket) MarshalMsgpack() ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)

	if err := enc.EncodeString(m.StreamName); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint32(uint32(m.PacketIndex)); err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(m.Content)
	if err != nil {
		return nil, err
	}
	if err := enc.EncodeString(string(jsonBytes)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


func (d DataPacket) MarshalMsgpack() ([]byte, error) {
	var w bytes.Buffer
	enc := msgpack.NewEncoder(&w)

	if err := enc.EncodeString(d.StreamName); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint32(d.PacketIndex); err != nil {
		return nil, err
	}
	if err := enc.EncodeBytes(d.Data); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

var (
	_ msgpack.Marshaler     = DataPacket{}
)

func init() {
	msgpack.RegisterExt(int8(0), DataPacket{})
}

func main() {
	data := [][]byte{[]byte("a")}
	dataPayload, lastPacketIndex := encodeDataPackets("st", 1, data)
	//metadataPayload := encodeMetadata("s", 0, map[string]interface{}{"foo": "bar"})
	//payload := append(dataPayload, metadataPayload...)
	payload := dataPayload

	fmt.Printf("lastPacketIndex: %d\n", lastPacketIndex)
	fmt.Printf("encoded: %s\n", hex.EncodeToString(payload))

	dp := DataPacket{
		StreamName:  "st",
		PacketIndex: 1,
		Data:        []byte("a"),
	}
	r1 := encDP1(dp)
	fmt.Printf("encDP1: %s\n", hex.EncodeToString(r1))
	r2 := encDP2(dp)
	fmt.Printf("encDP2: %s\n", hex.EncodeToString(r2))
	fmt.Printf("encDP1() == encDP2() -> %t", bytes.Equal(r1, r2))
}

func encDP1(d DataPacket) []byte {
	buf1 := bytes.NewBuffer([]byte{})
	enc1 := msgpack.NewEncoder(buf1)
	_ = enc1.Encode(d.StreamName)
	_ = enc1.Encode(d.PacketIndex)
	_ = enc1.Encode(d.Data)
	buf2 := bytes.NewBuffer([]byte{})
	enc2 := msgpack.NewEncoder(buf2)
	_ = enc2.EncodeExtHeader(int8(0), len(buf1.Bytes()))
	return append(buf2.Bytes(), buf1.Bytes()...)
}

func encDP2(d DataPacket) []byte {
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	_ = enc.Encode(d)
	return buf.Bytes()
}

func encodeDataPackets(stream string, initialPacketIndex uint32, data [][]byte) ([]byte, uint32) {
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	lastPacketIndex := initialPacketIndex
	for i, datum := range data {
		// handle encoding error properly in the real code
		_ = enc.Encode(stream)
		_ = enc.Encode(lastPacketIndex)
		_ = enc.Encode(datum)
		if i < len(data)-1 {
			lastPacketIndex++
		}
	}
	return buf.Bytes(), lastPacketIndex
}

func encodeMetadata(stream string, packetIndex uint64, content map[string]interface{}) []byte {
	// handle encoding error properly in the real code
	bs, _ := json.Marshal(content)
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	_ = enc.Encode(stream)
	_ = enc.Encode(packetIndex)
	_ = enc.Encode(string(bs))
	return buf.Bytes()
}
