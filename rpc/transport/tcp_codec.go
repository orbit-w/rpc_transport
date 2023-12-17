package transport

import (
	"encoding/binary"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"io"
	"log"
	"net"
	"time"
)

/*
   @Author: orbit-w
   @File: tcp_codec
   @2023 11月 周日 14:21
*/

const (
	gzipSize = 1
)

// TcpCodec TODO: 不支持压缩
type TcpCodec struct {
	isGzip          bool //压缩标识符（建议超过100byte消息进行压缩）
	maxIncomingSize uint32
}

func NewTcpCodec(max uint32, _isGzip bool) *TcpCodec {
	return &TcpCodec{
		isGzip:          _isGzip,
		maxIncomingSize: max,
	}
}

// EncodeBody 消息编码协议 body: size<int32> | gzipped<bool> | body<bytes>
func (tcd *TcpCodec) EncodeBody(body packet.IPacket) packet.IPacket {
	defer body.Return()
	pack := packet.Writer()
	//TODO: 默认 gzip 为 false
	tcd.buildPacket(pack, body, false)
	return pack
}

func (tcd *TcpCodec) BlockDecodeBody(conn net.Conn, header, body []byte) (packet.IPacket, error) {
	err := conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(conn, header)
	if err != nil {
		if err != io.EOF && !mmrpcs.IsClosedConnError(err) {
			log.Println("[TcpCodec] [func:BlockDecodeBody] receive data head failed: ", err.Error())
		}
		return nil, err
	}

	size := binary.BigEndian.Uint32(header)
	if size > tcd.maxIncomingSize {
		return nil, mmrpcs.ExceedMaxIncomingPacket(size)
	}

	body = body[:size]
	if _, err = io.ReadFull(conn, body); err != nil {
		return nil, transport_err.ReadBodyFailed(err)
	}
	buf := packet.Writer()
	buf.Write(body)

	//TODO:gzip
	_, err = buf.ReadBool()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// body: size<int32> | gzipped<byte> | body<bytes>
func (tcd *TcpCodec) buildPacket(buf, data packet.IPacket, gzipped bool) {
	size := data.Len()
	buf.WriteInt32(int32(size) + gzipSize)
	buf.WriteBool(gzipped)
	buf.Write(data.Data())
}

func (tcd *TcpCodec) checkPacketSize(header []byte) error {
	if size := binary.BigEndian.Uint32(header); size > tcd.maxIncomingSize {
		return transport_err.ExceedMaxIncomingPacket(size)
	}
	return nil
}
