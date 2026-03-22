package socks5

import (
	"testing"
)

func TestFindSNIInPayload_ValidClientHello(t *testing.T) {
	// Minimal TLS 1.2 ClientHello with SNI "example.com"
	hello := buildTestClientHello("example.com")

	start, end, ok := findSNIInPayload(hello)
	if !ok {
		t.Fatal("expected to find SNI")
	}
	sni := string(hello[start:end])
	if sni != "example.com" {
		t.Fatalf("expected 'example.com', got %q", sni)
	}
}

func TestFindSNIInPayload_NotTLS(t *testing.T) {
	// MTProto-like data (not TLS)
	data := make([]byte, 105)
	data[0] = 0xef
	_, _, ok := findSNIInPayload(data)
	if ok {
		t.Fatal("should not find SNI in non-TLS data")
	}
}

func TestFindSNIInPayload_TooShort(t *testing.T) {
	_, _, ok := findSNIInPayload([]byte{0x16, 0x03, 0x01})
	if ok {
		t.Fatal("should not find SNI in too-short data")
	}
}

func TestFindSNIInPayload_EmptyPayload(t *testing.T) {
	_, _, ok := findSNIInPayload(nil)
	if ok {
		t.Fatal("should not find SNI in nil data")
	}
}

func TestFindSNIInPayload_LongDomain(t *testing.T) {
	domain := "subdomain.deep.nested.example.co.uk"
	hello := buildTestClientHello(domain)
	start, end, ok := findSNIInPayload(hello)
	if !ok {
		t.Fatal("expected to find SNI")
	}
	if got := string(hello[start:end]); got != domain {
		t.Fatalf("expected %q, got %q", domain, got)
	}
}

// buildTestClientHello builds a minimal TLS ClientHello with the given SNI.
func buildTestClientHello(serverName string) []byte {
	sniBytes := []byte(serverName)
	sniLen := len(sniBytes)

	// SNI extension: type(2) + len(2) + sni_list_len(2) + type(1) + name_len(2) + name
	sniExt := []byte{
		0x00, 0x00, // extension type: server_name
		byte((sniLen + 5) >> 8), byte((sniLen + 5) & 0xff), // extension data length
		byte((sniLen + 3) >> 8), byte((sniLen + 3) & 0xff), // server name list length
		0x00,                              // host name type
		byte(sniLen >> 8), byte(sniLen),   // host name length
	}
	sniExt = append(sniExt, sniBytes...)

	extLen := len(sniExt)

	// ClientHello body: version(2) + random(32) + session_id_len(1) + cipher_suites_len(2) + ciphers(2) + comp_len(1) + comp(1) + ext_len(2) + extensions
	chBody := make([]byte, 0, 2+32+1+2+2+1+1+2+extLen)
	chBody = append(chBody, 0x03, 0x03) // TLS 1.2
	chBody = append(chBody, make([]byte, 32)...) // random
	chBody = append(chBody, 0x00) // session ID length = 0
	chBody = append(chBody, 0x00, 0x02, 0x00, 0x2f) // cipher suites: length=2, TLS_RSA_WITH_AES_128_CBC_SHA
	chBody = append(chBody, 0x01, 0x00) // compression: length=1, null
	chBody = append(chBody, byte(extLen>>8), byte(extLen&0xff)) // extensions length
	chBody = append(chBody, sniExt...)

	// Handshake header: type(1) + length(3)
	chLen := len(chBody)
	handshake := []byte{0x01, byte(chLen >> 16), byte(chLen >> 8), byte(chLen)}
	handshake = append(handshake, chBody...)

	// TLS record header: type(1) + version(2) + length(2)
	recLen := len(handshake)
	record := []byte{0x16, 0x03, 0x01, byte(recLen >> 8), byte(recLen)}
	record = append(record, handshake...)

	return record
}
