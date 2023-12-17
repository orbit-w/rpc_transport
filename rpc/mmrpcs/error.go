package mmrpcs

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errPattern          = "rpc err: "
	ErrTimeout          = errors.New("timeout")
	ErrCanceled         = errors.New("context canceled")
	ErrDeadlineExceeded = errors.New("context deadline exceeded")
	ErrDisconnect       = errors.New("disconnect")
	ErrMaxOfRetry       = errors.New(`error_max_of_retry`)
	ErrRpcDisconnected  = errors.New("error_rpc_disconnected")
)

func IsCancelError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "context canceled")
}

func NewRpcError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(errPattern + err.Error())
}

func IsClosedConnError(err error) bool {
	/*
		`use of closed file or network connection` (Go ver > 1.8, internal/pool.ErrClosing)
		`mux: listener closed` (cmux.ErrListenerClosed)
	*/
	return err != nil && strings.Contains(err.Error(), "closed")
}

func ExceedMaxIncomingPacket(size uint32) error {
	return errors.New(fmt.Sprintf("exceed max incoming packet size: %d", size))
}

func ReadBodyFailed(err error) error {
	return errors.New(fmt.Sprintf("read body failed: %s", err.Error()))
}

func ReceiveBufPutErr(err error) error {
	return errors.New(fmt.Sprintf("receiveBuf put failed: %s", err.Error()))
}
