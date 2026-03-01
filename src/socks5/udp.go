package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/log"
)

// handleUDPAssociate handles the SOCKS5 UDP ASSOCIATE command.
// Keeps TCP connection alive and manages UDP relay.
func (s *Server) handleUDPAssociate(conn net.Conn) error {
	// Send success reply with UDP bind address
	if err := sendReply(conn, repSuccess, s.udpConn.LocalAddr()); err != nil {
		return fmt.Errorf("send UDP reply: %w", err)
	}

	log.Debugf("SOCKS5 UDP ASSOCIATE from %s, UDP relay address: %s", conn.RemoteAddr(), s.udpConn.LocalAddr())

	// Keep TCP connection open - when it closes, UDP association ends
	// Read from TCP connection to detect when client closes it
	bufPtr := s.bufferPool.Get().(*[]byte)
	defer s.bufferPool.Put(bufPtr)

	for {
		_, err := conn.Read(*bufPtr)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				log.Debugf("SOCKS5 UDP associate closed: %s", conn.RemoteAddr())
				return nil
			}
			return err
		}
	}
}

// udpReadLoop reads UDP packets from the shared listener and dispatches them.
func (s *Server) udpReadLoop() {
	// Connection pool for UDP relay
	conns := &sync.Map{}

	defer func() {
		// Clean up all connections on shutdown
		conns.Range(func(key, value interface{}) bool {
			if conn, ok := value.(net.Conn); ok {
				conn.Close()
			}
			return true
		})
	}()

	bufPtr := s.bufferPool.Get().(*[]byte)
	defer s.bufferPool.Put(bufPtr)

	for {
		n, clientAddr, err := s.udpConn.ReadFromUDP((*bufPtr)[:cap(*bufPtr)])
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Errorf("SOCKS5 UDP read: %v", err)
			continue
		}

		// Parse SOCKS5 UDP datagram
		pkt := (*bufPtr)[:n]
		if len(pkt) < 10 {
			continue
		}
		if pkt[0] != 0 || pkt[1] != 0 {
			continue // reserved must be 0
		}
		if pkt[2] != 0 {
			continue // fragmentation not supported
		}

		dest, dataOff, err := parseUDPAddress(pkt)
		if err != nil {
			continue
		}

		data := pkt[dataOff:]

		// Handle packet with connection pooling
		go s.handleUDPPacket(clientAddr, dest, data, conns)
	}
}

// handleUDPPacket processes one incoming SOCKS5 UDP packet using connection pool.
func (s *Server) handleUDPPacket(clientAddr *net.UDPAddr, dest string, data []byte, conns *sync.Map) {
	// Create unique key for this client-destination pair
	connKey := clientAddr.String() + "--" + dest

	// Try to get existing connection
	var target net.Conn
	if val, ok := conns.Load(connKey); ok {
		target = val.(net.Conn)
	} else {
		// Create new connection
		var err error
		target, err = net.Dial("udp", dest)
		if err != nil {
			log.Tracef("SOCKS5 UDP dial to %s failed: %v", dest, err)
			return
		}

		// Store connection
		conns.Store(connKey, target)

		// Start goroutine to read responses from this connection
		go s.udpReadFromTarget(target, clientAddr, dest, connKey, conns)
	}

	// Send data to target
	sent, err := target.Write(data)
	if err != nil {
		log.Tracef("SOCKS5 UDP write to %s failed: %v", dest, err)
		target.Close()
		conns.Delete(connKey)
		return
	}

	// Log metrics (only for sent data, responses are logged in udpReadFromTarget)
	log.Tracef("SOCKS5 UDP sent %d bytes: %s -> %s", sent, clientAddr, dest)
}

// udpReadFromTarget reads responses from target server and sends back to client.
func (s *Server) udpReadFromTarget(target net.Conn, clientAddr *net.UDPAddr, dest string, connKey string, conns *sync.Map) {
	defer func() {
		target.Close()
		conns.Delete(connKey)
	}()

	// Set read timeout
	readTimeout := time.Duration(s.cfg.UDPReadTimeout) * time.Second
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second // Longer timeout for persistent connections
	}

	bufPtr := s.bufferPool.Get().(*[]byte)
	defer s.bufferPool.Put(bufPtr)

	for {
		target.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := target.Read((*bufPtr)[:cap(*bufPtr)])
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			// Timeout or other error - close connection
			return
		}

		// Parse target address
		destUDP, err := net.ResolveUDPAddr("udp", dest)
		if err != nil {
			continue
		}

		// Build SOCKS5 UDP response
		reply := buildUDPReply((*bufPtr)[:n], destUDP)

		// Send response back to client
		_, err = s.udpConn.WriteToUDP(reply, clientAddr)
		if err != nil {
			log.Errorf("SOCKS5 UDP failed to reply to client %s: %v", clientAddr, err)
			return
		}

		// Log metrics
		s.logUDPMetrics(clientAddr, dest, 0, n)
	}
}

// logUDPMetrics logs UDP connection metrics
func (s *Server) logUDPMetrics(clientAddr *net.UDPAddr, dest string, sent, received int) {
	// Extract client info and destination for logging/metrics
	clientAddrStr := clientAddr.String()
	clientHost := clientAddr.IP.String()
	clientPort := clientAddr.Port

	// Extract domain and destination info
	domain := dest
	destHost, destPortStr, _ := net.SplitHostPort(dest)
	if destHost != "" {
		domain = destHost
	}
	destPort := 0
	if p, err := strconv.Atoi(destPortStr); err == nil {
		destPort = p
	}

	// Match destination against configured sets
	matchedSNI, sniTarget, matchedIP, ipTarget := s.matchDestination(dest)

	// Determine which set to use for metrics
	setName := ""
	if matchedSNI {
		setName = sniTarget
	} else if matchedIP {
		setName = ipTarget
	}

	// Log in CSV format for UI (matching nfq.go format)
	// Format: ,PROTOCOL,sniTarget,host,source:port,ipTarget,destination:port,sourceMac
	// Use P-UDP to indicate proxy traffic
	if !log.IsDiscoveryActive() {
		log.Infof(",P-UDP,%s,%s,%s:%d,%s,%s:%d,", sniTarget, domain, clientHost, clientPort, ipTarget, destHost, destPort)
	}

	// Also log in human-readable format (debug level)
	if sent > 0 || received > 0 {
		log.Debugf("SOCKS5 UDP: %s -> %s (%d bytes sent, %d bytes received, Set: %s)", clientAddrStr, dest, sent, received, setName)
	}

	// Record connection in metrics for UI display
	m := getMetricsCollector()
	if m != nil {
		matched := matchedSNI || matchedIP
		m.RecordConnection("P-UDP", domain, clientAddrStr, dest, matched, "", setName)
	}
}

// parseUDPAddress extracts the destination address from a SOCKS5 UDP packet.
// Returns the address string and the offset where payload data begins.
func parseUDPAddress(pkt []byte) (addr string, dataOffset int, err error) {
	if len(pkt) < 4 {
		return "", 0, fmt.Errorf("packet too short")
	}

	atyp := pkt[3]
	switch atyp {
	case atypIPv4:
		if len(pkt) < 10 {
			return "", 0, fmt.Errorf("packet too short for IPv4")
		}
		ip := net.IP(pkt[4:8])
		port := binary.BigEndian.Uint16(pkt[8:10])
		return net.JoinHostPort(ip.String(), strconv.Itoa(int(port))), 10, nil

	case atypIPv6:
		if len(pkt) < 22 {
			return "", 0, fmt.Errorf("packet too short for IPv6")
		}
		ip := net.IP(pkt[4:20])
		port := binary.BigEndian.Uint16(pkt[20:22])
		return net.JoinHostPort(ip.String(), strconv.Itoa(int(port))), 22, nil

	case atypDomain:
		if len(pkt) < 5 {
			return "", 0, fmt.Errorf("packet too short for domain length")
		}
		dlen := int(pkt[4])
		end := 5 + dlen + 2
		if len(pkt) < end {
			return "", 0, fmt.Errorf("packet too short for domain")
		}
		domain := string(pkt[5 : 5+dlen])
		port := binary.BigEndian.Uint16(pkt[5+dlen : end])
		return net.JoinHostPort(domain, strconv.Itoa(int(port))), end, nil

	default:
		return "", 0, fmt.Errorf("unsupported address type %d", atyp)
	}
}

// buildUDPReply constructs a SOCKS5 UDP response packet.
func buildUDPReply(data []byte, from *net.UDPAddr) []byte {
	// RSV(2) + FRAG(1) + ATYP(1) + ADDR(4|16) + PORT(2) + DATA
	var hdr []byte
	hdr = append(hdr, 0, 0, 0) // RSV, FRAG

	if ip4 := from.IP.To4(); ip4 != nil {
		hdr = append(hdr, atypIPv4)
		hdr = append(hdr, ip4...)
	} else {
		hdr = append(hdr, atypIPv6)
		hdr = append(hdr, from.IP.To16()...)
	}

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, uint16(from.Port))
	hdr = append(hdr, portBuf...)

	return append(hdr, data...)
}
