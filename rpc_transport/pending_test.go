package rpc_transport

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
   @Author: orbit-w
   @File: pending_test
   @2024 3月 周日 21:31
*/

func TestPending_Push(t *testing.T) {
	var (
		seq uint32
		p   Pending
	)
	p.Init(nil, time.Second*5)
	for i := 0; i < 100; i++ {
		seq++
		call := NewCall(seq)
		p.Push(call)
	}

	p.RangeAll(func(id uint32) {
		fmt.Println(id)
	})

	p.OnClose()
}

func TestPending_Pop(t *testing.T) {
	var (
		seq uint32
		p   Pending
	)
	p.Init(nil, time.Second*5)
	for i := 0; i < 10; i++ {
		seq++
		call := NewCall(seq)
		p.Push(call)
	}

	for i := uint32(1); i <= 10; i++ {
		_, exist := p.Pop(i)
		assert.True(t, exist)
	}

	p.OnClose()
}
