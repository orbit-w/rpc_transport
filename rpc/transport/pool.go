package transport

import "sync"

/*
   @Author: orbit-w
   @File: cache
   @2023 11月 周日 14:37
*/

type Buffer struct {
	Bytes []byte
}

var bodyPool *sync.Pool
var headPool = &sync.Pool{
	New: func() any {
		return &Buffer{
			Bytes: make([]byte, HeadLen),
		}
	},
}

func NewBodyPool(size uint32) {
	bodyPool = &sync.Pool{
		New: func() any {
			return &Buffer{
				Bytes: make([]byte, size),
			}
		},
	}
}
