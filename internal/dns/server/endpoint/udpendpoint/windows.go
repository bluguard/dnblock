//go:build windows

package udpendpoint

import "syscall"

// SO_REUSEPORT reuseport flag for sockets
const SO_REUSEPORT = 4

func reusePort(_, _ string, conn syscall.RawConn) error {

	return conn.Control(func(descriptor uintptr) {
		_ = syscall.SetsockoptInt(syscall.Handle(descriptor), syscall.SOL_SOCKET, SO_REUSEPORT, 1)
	})
}
