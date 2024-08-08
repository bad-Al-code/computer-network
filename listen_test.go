package listen

import (
	"io"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer func() {
		if closeErr := listener.Close(); closeErr != nil {
			t.Errorf("failed to close listener: %v", closeErr)
		}
	}()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					t.Log("listener timed out")
				} else {
					t.Errorf("failed to accept connection: %v", err)
				}
				return
			}

			go handleConnection(t, conn, done)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close connection: %v", closeErr)
		}
	}()

	<-done
}

func handleConnection(t *testing.T, conn net.Conn, done chan<- struct{}) {
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close connection: %v", closeErr)
		}
		done <- struct{}{}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				t.Log("connection closed by client")
			} else {
				t.Errorf("failed to read from connection: %v", err)
			}
			return
		}
		t.Logf("received: %q", buf[:n])
	}
}
