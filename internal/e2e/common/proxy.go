package e2e_common

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type PausableProxy struct {
	listenAddr         string
	targetAddr         string
	listener           net.Listener
	holdingConnections bool
	heldConns          []net.Conn
	heldLock           sync.Mutex
	stopChan           chan struct{}
	wg                 sync.WaitGroup
}

func NewPausableProxy(listenAddr, targetAddr string) *PausableProxy {
	return &PausableProxy{
		listenAddr: listenAddr,
		targetAddr: targetAddr,
		stopChan:   make(chan struct{}),
	}
}

func (p *PausableProxy) Start() error {
	var err error
	p.listener, err = net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}

	p.wg.Add(1)
	go p.acceptLoop()
	return nil
}

func (p *PausableProxy) Stop() {
	close(p.stopChan)
	p.listener.Close()
	p.wg.Wait()
}

func (p *PausableProxy) HoldNewIncomingConnections() {
	p.heldLock.Lock()
	p.holdingConnections = true
	p.heldLock.Unlock()
}

func (p *PausableProxy) DoNotHoldNewIncomingConnections() {
	p.heldLock.Lock()
	p.holdingConnections = false
	p.heldLock.Unlock()
}

func (p *PausableProxy) ResumeHeldConnections() {
	p.heldLock.Lock()
	held := p.heldConns
	p.heldConns = nil
	p.holdingConnections = false
	p.heldLock.Unlock()

	for _, conn := range held {
		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

func (p *PausableProxy) acceptLoop() {
	defer p.wg.Done()
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.stopChan:
				return
			default:
				log.Println("Accept error:", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		p.heldLock.Lock()
		if p.holdingConnections {
			p.heldConns = append(p.heldConns, conn)
			p.heldLock.Unlock()
			continue
		}
		p.heldLock.Unlock()

		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

func (p *PausableProxy) handleConnection(src net.Conn) {
	defer p.wg.Done()

	dst, err := net.Dial("tcp", p.targetAddr)
	if err != nil {
		log.Println("Failed to connect to target:", err)
		src.Close()
		return
	}

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
	}()

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
	}()
}
