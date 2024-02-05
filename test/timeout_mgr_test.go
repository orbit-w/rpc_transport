package test

import (
	"fmt"
	"github.com/orbit-w/rpc_transport/rpc"
	"testing"
	"time"
)

func TestTimeout_Push(t *testing.T) {
	to := rpc.NewTimeoutMgr(func(ids []uint32) {
		fmt.Println(ids)
	})
	fmt.Println(time.Now())
	to.Push(100, time.Second*5)
	time.Sleep(time.Second)
	fmt.Println(time.Now())
	to.Push(100, time.Second*5)
	time.Sleep(time.Second)
	fmt.Println(time.Now())
	to.Push(100, time.Second*5)

	//to.Remove(100)

	fmt.Println("complete")
	time.Sleep(time.Second * 15)
}

func TestTimeout_Remove(t *testing.T) {
	to := rpc.NewTimeoutMgr(func(ids []uint32) {
		fmt.Println(ids)
	})

	for i := 0; i < 5; i++ {
		to.Push(uint32(i), time.Second*5)
	}
	to.Remove(10)
	time.Sleep(time.Second)
	for i := 5; i < 10; i++ {
		to.Push(uint32(i), time.Second*5)
	}

	to.Remove(5)
	to.Remove(6)
	fmt.Println("complete")
	time.Sleep(time.Minute * 15)
}
