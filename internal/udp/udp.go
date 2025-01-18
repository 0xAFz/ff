package udp

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	bufferSize = 65535
	numWorkers = 10
)

type Forwarder struct {
	laddr    *net.UDPAddr
	raddr    *net.UDPAddr
	conn     *net.UDPConn
	connLock sync.RWMutex
	packetCh chan []byte
	doneCh   chan struct{}
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		packetCh: make(chan []byte, 1024),
		doneCh:   make(chan struct{}),
	}
}

func (f *Forwarder) Init(laddr, raddr string) error {
	l, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		log.Printf("failed to resolve local address: %v", err)
		return err
	}
	r, err := net.ResolveUDPAddr("udp", raddr)
	if err != nil {
		log.Printf("failed to resolve destination address: %v", err)
		return err
	}

	f.laddr = l
	f.raddr = r

	return nil
}

func (f *Forwarder) checkConnection() error {
	_, err := f.conn.Write([]byte("ping"))
	if err != nil {
		return fmt.Errorf("failed to send ping to %s: %v", f.raddr, err)
	}

	f.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	_, err = f.conn.Read(buf)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return fmt.Errorf("no response to ping, connection timeout: %v", err)
		}
		return fmt.Errorf("failed to receive ping response from %s: %v", f.raddr, err)
	}

	return nil
}

func (f *Forwarder) Connect() {
	for {
		conn, err := net.DialUDP("udp", nil, f.raddr)
		if err != nil {
			log.Printf("failed to connect to %s: %v", f.raddr, err)
			continue
		}

		f.connLock.Lock()
		f.conn = conn
		f.connLock.Unlock()

		err = f.checkConnection()
		if err != nil {
			log.Printf("connection to %s is invalid: %v", f.raddr, err)
			time.Sleep(1 * time.Second)
			conn.Close()
			continue
		}

		log.Printf("connected to %s", f.raddr)
		break
	}
}

func (f *Forwarder) Listen() error {
	f.Connect()

	c, err := net.ListenUDP("udp", f.laddr)
	if err != nil {
		log.Fatalf("failed to bind socket on %s: %v", f.laddr, err)
		return err
	}

	defer c.Close()
	defer f.conn.Close()

	log.Printf("listening on %s and forwarding to %s\n", f.laddr.String(), f.raddr.String())

	buf := make([]byte, bufferSize)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go f.worker()
	}

	for {
		n, addr, err := c.ReadFrom(buf)
		if err != nil {
			fmt.Printf("failed to read packet from %s: %v", addr, err)
			continue
		}

		log.Printf("received %d bytes from %s", n, addr)

		b := make([]byte, n)
		copy(b, buf[:n])

		// Send the data to the channel for workers to handle
		select {
		case f.packetCh <- b: // If channel is not full, send the packet to the channel
		default:
			log.Printf("packet dropped due to full channel")
		}
	}
}

func (f *Forwarder) worker() {
	for {
		select {
		case packet, ok := <-f.packetCh:
			start := time.Now()
			if !ok {
				return // Channel closed, exit worker
			}
			_, err := f.conn.Write(packet)
			if err != nil {
				log.Printf("failed to send packet to %s: %v", f.raddr.String(), err)
				// Reconnect if the connection is down
				log.Printf("reconnecting to %s...", f.raddr.String())
				f.Connect()
				continue
			}

			log.Printf("packet forwarded to %s in %v", f.raddr.String(), time.Since(start))
		case <-f.doneCh:
			return // Signal to stop, exit worker
		}
	}
}

func (f *Forwarder) Stop() {
	close(f.doneCh)
	close(f.packetCh)
}
