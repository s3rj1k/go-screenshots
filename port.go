package screenshot

import (
	"errors"
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

	defer l.Close()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("unexpected data type from TCP listener")
	}

	return addr.Port, nil
}
