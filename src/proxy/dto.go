package proxy

import (
	"encoding/binary"
)

const (
	PackageTypeSelect = iota
	PackageTypeInform
	PackageTypeBroadCast
	PackageTypeBroadCastResponse
	PackageTypeConnectHost
)

type Packet struct {
	size    int
	pkgType int
	src     int
	dst     int
	data    []byte
}

func (p *Packet) PacketType() int {
	return p.pkgType
}

func (p *Packet) SrcAddr() int {
	return p.src
}

func (p *Packet) DstAddr() int {
	return p.dst
}

func (p *Packet) Data() []byte {
	return p.data
}

func (p *Packet) Len() int {
	return p.size
}

func (p *Packet) ToBytes() []byte {
	res := make([]byte, 16+len(p.data))
	binary.BigEndian.PutUint32(res[0:4], uint32(p.size))
	binary.BigEndian.PutUint32(res[4:8], uint32(p.pkgType))
	binary.BigEndian.PutUint32(res[8:12], uint32(p.src))
	binary.BigEndian.PutUint32(res[12:16], uint32(p.dst))
	copy(res[16:], p.data)
	return res
}

func NewSelectPacket(src, dst int, data []byte) *Packet {
	return NewPacket(PackageTypeSelect, src, dst, data)
}

func NewInformPacket(src, dst int, data []byte) *Packet {
	return NewPacket(PackageTypeInform, src, dst, data)
}

func NewPacket(pkgType, src, dst int, data []byte) *Packet {
	size := 16 + len(data)
	return &Packet{
		size,
		pkgType,
		src,
		dst,
		data,
	}
}

func PacketFromBytes(data []byte) (result []*Packet) {
	result = make([]*Packet, 0)
	if data == nil {
		return
	}
	read := 0
	for read < len(data)-4 {
		size := int(binary.BigEndian.Uint32(data[read : read+4]))
		endBytes := size + read
		if endBytes > len(data) {
			endBytes = len(data)
		}
		result = append(result, singlePacketFromBytes(data[read:endBytes]))
		read = read + size
	}
	return result
}

func singlePacketFromBytes(data []byte) *Packet {
	size, src, dst, pkgType := -1, -1, -1, -1
	if len(data) >= 4 {
		size = int(binary.BigEndian.Uint32(data[0:4]))
	}
	if len(data) >= 8 {
		pkgType = int(binary.BigEndian.Uint32(data[4:8]))
	}
	if len(data) >= 12 {
		src = int(binary.BigEndian.Uint32(data[8:12]))
	}
	if len(data) >= 16 {
		dst = int(binary.BigEndian.Uint32(data[12:16]))
	}
	var dataMsg []byte
	if len(data) > 16 {
		dataMsg = data[16:]
	}
	return &Packet{
		size, pkgType, src, dst, dataMsg,
	}
}
