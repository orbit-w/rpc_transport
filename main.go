package main

import (
	"context"
	"github.com/orbit-w/rpc_transport/rpc"
	"log"
)

/*
   @Author: orbit-w
   @File: main
   @2023 12月 周三 23:07
*/

func Client() {
	host := "127.0.0.1:6900"
	cli, err := rpc.Dial("node_00", "node_01", host)
	if err != nil {
		panic(err.Error())
	}

	_, err = cli.Call(context.Background(), []byte{1})
	if err != nil {
		panic(err.Error())
	}

	if err = cli.AsyncCallC([]byte{3}, 100, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
		return nil
	}); err != nil {
		panic(err.Error())
	}
}

func Server() {
	host := "127.0.0.1:6900"
	err := rpc.Serve(host, func(req rpc.IRequest) error {
		_ = req.Response([]byte{1})
		req.Return()
		return nil
	})
	if err != nil {
		panic(err.Error())
	}
}
