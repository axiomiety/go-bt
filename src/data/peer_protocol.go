package data

import (
	"bytes"
	"fmt"
)

type Handshake struct {
	PstrLen  byte
	Pstr     []byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerId   [20]byte
}

func GetHanshake(peerId [20]byte, infoHash [20]byte) *Handshake {
	pstr := []byte("BitTorrent protocol")
	return &Handshake{
		PstrLen:  byte(len(pstr)),
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerId:   peerId,
	}
}

func (h *Handshake) ToBytes() []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteByte(h.PstrLen)
	buffer.Write(h.Pstr)
	buffer.Write(h.Reserved[:])
	buffer.Write(h.InfoHash[:])
	buffer.Write(h.PeerId[:])
	return buffer.Bytes()
}

type Message struct {
	Length    [4]byte
	MessageId byte
	Payload   []byte
}

func (m *Message) ToBytes() []byte {
	buffer := new(bytes.Buffer)
	buffer.Write(m.Length[:])
	// handle the special keep-alive case
	if m.Length == [4]byte{0, 0, 0, 0} {
		return buffer.Bytes()
	}
	buffer.WriteByte(m.MessageId)
	buffer.Write(m.Payload)
	return buffer.Bytes()
}

func KeepAlive() *Message {
	return &Message{}
}

func Choke() *Message {
	return &Message{
		Length: [4]byte{0, 0, 0, 1},
	}
}

type BitField struct {
	NumBlocks uint64
	Field     []byte
}

func (b *BitField) HasBlock(idx uint64) bool {
	if idx > b.NumBlocks {
		panic(fmt.Sprintf("We only have %d blocks but requested block number %d", b.NumBlocks, idx))
	}

	// find the relevant byte
	byteIdx := idx / 8
	// blocks are 0-indexed
	offset := byte(1 << (8 - (idx % 8) - 1))
	return b.Field[byteIdx]&offset > 0
}

func (b *BitField) SetBlock(idx uint64) {
	if idx > b.NumBlocks {
		panic(fmt.Sprintf("We only have %d blocks but tried to set block number %d", b.NumBlocks, idx))
	}

	// find the relevant byte
	byteIdx := idx / 8
	// blocks are 0-indexed
	offset := byte(1 << (8 - (idx % 8) - 1))
	b.Field[byteIdx] |= offset
}
