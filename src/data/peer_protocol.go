package data

import (
	"bytes"
	"encoding/binary"
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

func Request(index uint32, begin uint32, length uint32) *Message {
	buffer := make([]byte, 4*3)
	binary.BigEndian.PutUint32(buffer[0:], index)
	binary.BigEndian.PutUint32(buffer[4:], begin)
	binary.BigEndian.PutUint32(buffer[8:], length)
	ll := make([]byte, 4)
	binary.BigEndian.PutUint32(ll, 13)
	return &Message{
		Length:    [4]byte(ll),
		MessageId: MsgRequest,
		Payload:   buffer,
	}
}

type BitField struct {
	Field []byte
}

func (b *BitField) NumPieces() uint32 {
	// each byte represents 8 blocks
	return uint32(cap(b.Field))
}

func (b *BitField) HasPiece(idx uint32) bool {
	if idx > b.NumPieces() {
		panic(fmt.Sprintf("We only have %d blocks but requested block number %d", b.NumPieces(), idx))
	}

	// find the relevant byte
	byteIdx := idx / 8
	// blocks are 0-indexed
	offset := byte(1 << (8 - (idx % 8) - 1))
	return b.Field[byteIdx]&offset > 0
}

func (b *BitField) SetPiece(idx uint32) {
	if idx > b.NumPieces() {
		panic(fmt.Sprintf("We only have %d blocks but tried to set block number %d", b.NumPieces(), idx))
	}

	// find the relevant byte
	byteIdx := idx / 8
	// blocks are 0-indexed
	offset := byte(1 << (8 - (idx % 8) - 1))
	b.Field[byteIdx] |= offset
}

const (
	MsgChoke         byte = 0
	MsgUnchoke       byte = 1
	MsgInterested    byte = 2
	MsgNotInterested byte = 3
	MsgHave          byte = 4
	MsgBitfield      byte = 5
	MsgRequest       byte = 6
	MsgPiece         byte = 7
	MsgCancel        byte = 8
)
