package rpc_transport

import (
	"errors"
)

var (
	errPattern    = "rpc_transport err: "
	ErrTimeout    = errors.New("timeout")
	ErrDisconnect = errors.New("disconnect")
)

func NewRpcError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(errPattern + err.Error())
}
