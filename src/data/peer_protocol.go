package data

import "bytes"

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
