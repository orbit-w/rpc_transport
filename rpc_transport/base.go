package rpc_transport

import "time"

/*
   @Author: orbit-w
   @File: base
   @2023 12月 周日 11:13
*/

const (
	RpcRaw = iota
	RpcCall
	RpcAsyncCall

	RpcTimeout = time.Second * 5
)

// Client state enum
const (
	TypeNone = iota
	TypeRunning
	TypeStopped
)

type (
	replyMsg struct {
		category int8
		seq      uint32
		pid      int64
		buf      []byte
	}

	timeoutListMsg struct {
		ids []uint32
	}
)
