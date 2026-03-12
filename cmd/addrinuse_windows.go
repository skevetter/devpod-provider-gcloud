//go:build windows

package cmd

import (
	"errors"
	"syscall"
)

// WSAEADDRINUSE is the Windows socket error for "address already in use".
const wsaeaddrinuse = syscall.Errno(10048)

func isAddrInUse(err error) bool {
	return errors.Is(err, wsaeaddrinuse)
}
