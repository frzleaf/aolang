package proxy

import (
	"encoding/binary"
)

const (
	PackageTypeSelect = iota
	PackageTypeInform
	PackageTypeFindHost
	PackageTypeFindHostResponse
	PackageTypeConnectHost
)

type Packet struct {
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

func (p *Packet) ToBytes() []byte {
	res := make([]byte, 12+len(p.data))
	binary.BigEndian.PutUint32(res[0:4], uint32(p.pkgType))
	binary.BigEndian.PutUint32(res[4:8], uint32(p.src))
	binary.BigEndian.PutUint32(res[8:12], uint32(p.dst))
	copy(res[12:], p.data)
	return res
}

func NewSelectPacket(src, dst int, data []byte) *Packet {
	return &Packet{
		PackageTypeSelect,
		src,
		dst,
		data,
	}
}

func NewInformPacket(src, dst int, data []byte) *Packet {
	return &Packet{
		PackageTypeInform,
		src,
		dst,
		data,
	}
}

func NewPacket(pkgType, src, dst int, data []byte) *Packet {
	return &Packet{
		pkgType,
		src,
		dst,
		data,
	}
}

func PacketFromBytes(data []byte) *Packet {
	src, dst, pkgType := -1, -1, -1
	if len(data) >= 4 {
		pkgType = int(binary.BigEndian.Uint32(data[0:4]))
	}
	if len(data) >= 8 {
		src = int(binary.BigEndian.Uint32(data[4:8]))
	}
	if len(data) >= 12 {
		dst = int(binary.BigEndian.Uint32(data[8:12]))
	}
	var dataMsg []byte
	if len(data) > 12 {
		dataMsg = data[12:]
	}
	return &Packet{
		pkgType, src, dst, dataMsg,
	}
}
