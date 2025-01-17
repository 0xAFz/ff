package udp

import (
	"fmt"
	"log"
	"net"
)

func ListenForwarder(addr, dest string) error {
	c, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("failed to bind socket on %s: %v", addr, err)
		return err
	}

	defer c.Close()

	log.Printf("Listening on %s and forwarding to %s\n", addr, dest)

	d, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		log.Fatalf("failed to resolve udp address %s: %v", dest, err)
		return err
	}

	buf := make([]byte, 1500)

	for {
		n, a, err := c.ReadFrom(buf)
		if err != nil {
			fmt.Printf("failed to read packet from %s: %v", a, err)
			continue
		}

		log.Printf("recived %d bytes from %s", n, a)

		b := make([]byte, n)
		copy(b, buf[:n])

		go handleConn(c, b, n, d)
	}
}

func handleConn(c net.PacketConn, buf []byte, n int, remoteAddr net.Addr) {
	_, err := c.WriteTo(buf[:n], remoteAddr)
	if err != nil {
		log.Printf("failed to send packet to dest: %v", err)
		return
	}
	log.Printf("%d bytes forwarded to %s", n, remoteAddr)
}
