package rpc_transport

import (
	"fmt"
	"github.com/orbit-w/rpc_transport/timeout"
	"testing"
	"time"
)

func TestTimeout_Push(t *testing.T) {
	to := timeout.NewTimeout(func(ids []uint32) {
		fmt.Println(ids)
	})
	fmt.Println(time.Now())
	to.Insert(100, time.Second*5)
	time.Sleep(time.Second)
	fmt.Println(time.Now())
	to.Insert(100, time.Second*5)
	time.Sleep(time.Second)
	fmt.Println(time.Now())
	to.Insert(100, time.Second*5)

	//to.Remove(100)

	fmt.Println("complete")
	time.Sleep(time.Second * 15)
}

func TestTimeout_Remove(t *testing.T) {
	to := timeout.NewTimeout(func(ids []uint32) {
		fmt.Println(ids)
	})

	for i := 0; i < 5; i++ {
		to.Insert(uint32(i), time.Second*5)
	}
	to.Remove(10)
	time.Sleep(time.Second)
	for i := 5; i < 20; i++ {
		to.Insert(uint32(i), time.Second*5)
	}

	to.Remove(5)
	to.Remove(6)
	fmt.Println("complete")
	time.Sleep(time.Second * 30)
}
