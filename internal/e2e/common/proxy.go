package e2e_common

import (
	"io"
	"log"
	"net"
	"sync"
)

type PausableProxy struct {
	listenAddr string
	targetAddr string

	mu          sync.RWMutex
	paused      bool
	connections map[net.Conn]struct{}

	listener net.Listener
	done     chan struct{}
	once     sync.Once
}

func NewPausableProxy(listenAddr, targetAddr string) *PausableProxy {
	return &PausableProxy{
		listenAddr:  listenAddr,
		targetAddr:  targetAddr,
		connections: make(map[net.Conn]struct{}),
	}
}

func (p *PausableProxy) Start() error {
	ln, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}

	p.listener = ln
	p.done = make(chan struct{})

	go func() {
		for {
			clientConn, err := ln.Accept()
			if err != nil {
				select {
				case <-p.done:
					return
				default:
					log.Printf("Accept error: %v", err)
					continue
				}
			}

			p.mu.RLock()
			isPaused := p.paused
			p.mu.RUnlock()

			if isPaused {
				p.registerConn(clientConn)
				go p.holdConnection(clientConn)
			} else {
				go p.handleConnection(clientConn)
			}
		}
	}()

	return nil
}

func (p *PausableProxy) Stop() {
	p.once.Do(func() {
		if p.listener != nil {
			_ = p.listener.Close()
		}

		close(p.done)
		p.CloseHeldConnections()
	})
}

func (p *PausableProxy) PauseConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
}

func (p *PausableProxy) ResumeConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = false
}

func (p *PausableProxy) CloseHeldConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for conn := range p.connections {
		conn.Close()
		delete(p.connections, conn)
	}
}

func (p *PausableProxy) registerConn(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections[conn] = struct{}{}
}

func (p *PausableProxy) unregisterConn(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.connections, conn)
}

func (p *PausableProxy) holdConnection(conn net.Conn) {
	// Block until the connection is closed externally
	buf := make([]byte, 1)
	_, _ = conn.Read(buf) // Do nothing; just hold
	p.unregisterConn(conn)
}

func (p *PausableProxy) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	serverConn, err := net.Dial("tcp", p.targetAddr)
	if err != nil {
		log.Printf("Error dialing target: %v", err)
		return
	}
	defer serverConn.Close()

	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}
