package certpool

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
	"github.com/darvaza-proxy/slog"
)

// SetLogger binds a slog.Logger to the buffer
func (pb *PoolBuffer) SetLogger(logger slog.Logger) {
	pb.logger.Store(logger)
}

func (pb *PoolBuffer) withLogger(level slog.LogLevel) (slog.Logger, bool) {
	l, ok := pb.logger.Load().(slog.Logger)
	if !ok {
		return nil, false
	}

	return l.WithLevel(level).WithEnabled()
}

func (pb *PoolBuffer) debug() (slog.Logger, bool) {
	return pb.withLogger(slog.Debug)
}

func (pb *PoolBuffer) info() (slog.Logger, bool) {
	return pb.withLogger(slog.Debug)
}

func (pb *PoolBuffer) warn() (slog.Logger, bool) {
	return pb.withLogger(slog.Warn)
}

func (pb *PoolBuffer) error(err error) (slog.Logger, bool) {
	if l, ok := pb.withLogger(slog.Error); ok {
		if err != nil {
			l = l.WithField(slog.ErrorFieldName, err)
		}
		return l, true
	}
	return nil, false
}

func hexString(data []byte) string {
	var buf bytes.Buffer
	for i, x := range data {
		if i > 0 {
			_, _ = buf.WriteRune(':')
		}
		fmt.Fprintf(&buf, "%02X", x)
	}
	return buf.String()
}

func md5String(data []byte) string {
	hash := md5.Sum(data)
	return hexString(hash[:])
}

func pubKeyString(pub crypto.PublicKey) string {
	switch v := pub.(type) {
	case *rsa.PublicKey:
		return rsaPubKeyString(v)
	case *ecdsa.PublicKey:
		return ecdsaPubKeyString(v)
	case *ed25519.PublicKey:
		return ed25519PubKeyString(v)
	default:
		return fmt.Sprintf("%T", pub)
	}
}

func rsaPubKeyString(pub *rsa.PublicKey) string {
	const t = "rsa"
	var s = []string{
		fmt.Sprintf("%s%v", t, pub.Size()),
		fmt.Sprintf("%0X", pub.E),
		fmt.Sprintf("%0X", pub.N),
	}

	return strings.Join(s, ":")
}

func ecdsaPubKeyString(pub *ecdsa.PublicKey) string {
	const t = "ecdsa"
	var cp = pub.Curve.Params()

	var s = []string{
		fmt.Sprintf("%s%v", t, cp.BitSize),
		cp.Name,
		fmt.Sprintf("%0X", pub.X),
		fmt.Sprintf("%0X", pub.Y),
	}
	return strings.Join(s, ":")
}

func ed25519PubKeyString(pub *ed25519.PublicKey) string {
	const t = "ed25519"

	var s = []string{
		t,
		fmt.Sprintf("%0X", pub),
	}
	return strings.Join(s, ":")
}

func (pb *PoolBuffer) printKey(fn string, pk x509utils.PrivateKey) error {
	if log, ok := pb.info(); ok {
		fields := slog.Fields{
			"pub": pubKeyString(pk.Public()),
		}

		if fn != "" {
			fields["filename"] = fn
		}

		log.WithFields(fields).Print("Key Added")
	}
	return nil
}

// revive:disable:cognitive-complexity

func (pb *PoolBuffer) printCert(fn string, cert *x509.Certificate) error {
	// revive:enable:cognitive-complexity
	if log, ok := pb.info(); ok {
		fields := slog.Fields{
			"ca":      cert.IsCA,
			"subject": cert.Subject.String(),
			"md5":     md5String(cert.Raw),
			"pub":     pubKeyString(cert.PublicKey),
		}

		if fn != "" {
			fields["filename"] = fn
		}

		if len(cert.SubjectKeyId) > 0 {
			fields["subject-id"] = hexString(cert.SubjectKeyId)
		}

		names, patterns := x509utils.Names(cert)
		for i, s := range patterns {
			patterns[i] = "*" + s
		}
		if len(patterns) > 0 {
			fields["patterns"] = patterns
		}
		if len(names) > 0 {
			fields["names"] = names
		}

		log.WithFields(fields).Print("Certificate Added")
	}

	return nil
}
