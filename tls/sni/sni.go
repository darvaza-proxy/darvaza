package sni

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ClientHelloInfo struct {
	Version string
	SNI     string
	Valid   bool
}

const (
	tlsHandshakeType   = 22
	tlsSNIExtension    = 0
	tlsClientHelloType = 1
	tls10              = 769
	tls11              = 770
	tls12              = 771
	tlsMaxLength       = 16384
	tlsSessionIDLength = 32
)

var chi = new(ClientHelloInfo)

// Extracts SNI and version information from a TLS handshake.
func GetInfo(r io.Reader) (*ClientHelloInfo, error) {
	if err := parseRecord(r); err != nil {
		chi.reset()
		return chi, err
	}

	b, err := decodeVector(r, 2)
	if err != nil {
		// No extension (not an error).
		if err == io.EOF {
			err = nil
		}
		chi.reset()
		return chi, err
	}

	// Loop over the extensions.
	for len(b) >= 4 {
		extType := binary.BigEndian.Uint16(b[:2])
		length := binary.BigEndian.Uint16(b[2:4])
		b = b[4:]

		if extType == tlsSNIExtension {
			chi.SNI, err = parseSNI(b[:length])
			if err != nil {
				chi.reset()
				break
			}
		}
		b = b[length:]
	}
	chi.Valid = true
	return chi, err
}

func parseRecord(r io.Reader) error {
	var record struct {
		Type          uint8
		Major         uint8
		Minor         uint8
		Length        uint16
		MessageType   uint8
		MessageLength [3]byte
		Version       uint16
		Random        [32]byte
	}
	if err := binary.Read(r, binary.BigEndian, &record); err != nil {
		chi.reset()
		return err
	}

	if record.Type != tlsHandshakeType {
		chi.reset()
		return fmt.Errorf("not a TLS handshake")
	}

	if record.Length > tlsMaxLength {
		chi.reset()
		return fmt.Errorf("TLS record length exceed maximum (%d > 2^14)", record.Length)
	}
	if record.MessageType != tlsClientHelloType {
		chi.reset()
		return fmt.Errorf("not a ClientHello message (%d)", record.MessageType)
	}

	chi.Version = fmt.Sprintf("%d.%d", record.Major, record.Minor)

	// one byte SessionID
	b, err := decodeVector(r, 1)
	if err != nil {
		chi.reset()
		return fmt.Errorf("could not read ClientHello session ID (%s)", err)
	}

	if len(b) > tlsSessionIDLength {
		chi.reset()
		return fmt.Errorf("SessionID is bigger than allowed")
	}

	// two bytes cipher suites.
	b, err = decodeVector(r, 2)
	if err != nil {
		chi.reset()
		return err
	}
	if len(b) < 2 || len(b)%2 != 0 {
		chi.reset()
		return fmt.Errorf("ClientHello cipher suites has an invalid length (%d)", len(b))
	}

	// one byte compression methods.
	b, err = decodeVector(r, 1)
	if err != nil {
		chi.reset()
		return err
	}
	if len(b) < 1 {
		chi.reset()
		return fmt.Errorf("invalid length %d compression methods", len(b))
	}

	return nil
}

// Parse the SNI from an SNI extension.
func parseSNI(b []byte) (string, error) {
	if len(b) < 2 {
		return "", fmt.Errorf("SNI extension is empty")
	}

	length := binary.BigEndian.Uint16(b[:2])
	if int(length) > len(b[2:]) {
		chi.reset()
		return "", fmt.Errorf("SNI extension is too short")
	}

	b = b[2 : 2+length]

	for len(b) >= 3 {
		nameType := b[0]
		vectLength := binary.BigEndian.Uint16(b[1:3])
		if int(vectLength) > len(b[3:]) {
			chi.reset()
			return "", fmt.Errorf("SNI vector is too short")
		}

		if nameType != 0 {
			b = b[3+vectLength:]
			continue
		}

		return string(b[3 : 3+vectLength]), nil
	}

	// No DNS-based SNI.
	return "", nil
}

func decodeVector(r io.Reader, l uint) ([]byte, error) {
	rawLen := make([]byte, l)
	if err := binary.Read(r, binary.BigEndian, &rawLen); err != nil {
		// No data to read. This can be valid.
		if err == io.EOF {
			chi.reset()
			return nil, err
		}
		chi.reset()
		return nil, fmt.Errorf("could not read the vector length (%s)", err)
	}

	var length uint = 0
	for _, b := range rawLen {
		length = (length << 8) + uint(b)
	}

	if length == 0 {
		chi.reset()
		return nil, nil
	}

	data := make([]byte, length)
	if err := binary.Read(r, binary.BigEndian, &data); err != nil {
		chi.reset()
		return nil, fmt.Errorf("could not read the vector data (%s)", err)
	}

	return data, nil
}
func (c *ClientHelloInfo) reset() {
	c.SNI = ""
	c.Valid = false
	c.Version = ""
}
