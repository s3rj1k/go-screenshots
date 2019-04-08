package screenshot

import (
	"net"
)

// GetFreePort - returns free TCP port.
func GetFreePort() (int, error) {

	addr, err := net.ResolveTCPAddr("tcp", "[::]:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer func(l *net.TCPListener) {
		_ = l.Close()
	}(l)

	return l.Addr().(*net.TCPAddr).Port, nil
}
