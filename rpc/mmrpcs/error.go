package mmrpcs

import (
	"errors"
	"strings"
)

var (
	errPattern          = "rpc err: "
	ErrTimeout          = errors.New("rpc err: timeout")
	ErrCanceled         = errors.New("rpc err: context canceled")
	ErrDeadlineExceeded = errors.New("rpc err: context deadline exceeded")
	ErrDisconnect       = errors.New("rpc err: disconnect")
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
