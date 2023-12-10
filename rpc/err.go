package rpc

import (
	"errors"
	"strings"
)

/*
   @Author: orbit-w
   @File: i_model
   @2023 12月 周六 00:07
*/

var (
	ErrTimeout = errors.New("error_timeout")
)

func IsCancelError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "context canceled")
}
