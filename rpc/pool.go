package rpc

import (
	"sync"
)

/*
	@Author: orbit-w
	@File: pool
	@2023 12月 周六 11:20
*/

var (
	decodersPool = sync.Pool{New: func() any {
		return new(Decoder)
	}}

	reqPool = sync.Pool{New: func() any {
		return new(Request)
	}}
)
