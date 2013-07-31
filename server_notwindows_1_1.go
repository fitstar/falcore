// +build !windows
// +build go1.1

package falcore

import (
	"net"
	"runtime"
	"syscall"
)

// only valid on non-windows
func (srv *Server) setupNonBlockingListener(err error, l *net.TCPListener) error {
	// FIXME: File() returns a copied pointer.  we're leaking it.  probably doesn't matter
	if srv.listenerFile, err = l.File(); err != nil {
		return err
	}
	fd := int(srv.listenerFile.Fd())
	if e := syscall.SetNonblock(fd, true); e != nil {
		return e
	}
	return nil
}

// Used NoDelay (Nagle's algorithm) where available
func (srv *Server) setNoDelay(c net.Conn, noDelay bool) bool {
	switch runtime.GOOS {
	case "linux", "freebsd", "darwin":
		if tcpC, ok := c.(*net.TCPConn); ok {
			if noDelay {
				// Disable TCP_CORK/TCP_NOPUSH
				tcpC.SetNoDelay(true)
				// For TCP_NOPUSH, we need to force flush
				c.Write([]byte{})
			} else {
				// Re-enable TCP_CORK/TCP_NOPUSH
				tcpC.SetNoDelay(false)
			}
		}
		return true
	default:
		return false
	}
}
