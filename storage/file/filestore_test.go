package file

/*
import (
	"context"
	"crypto/x509"
	"fmt"
	"testing"
)

func Test_Get(t *testing.T) {
	q := Options{Directory: "/home/karasz/Downloads/Certs/PEM/"}
	z, err := NewStore(q)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	certdata, err := z.Get(ctx, "www.example.com")
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(certdata)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("found certificate", cert.Subject.CommonName)
}

func Test_Delete(t *testing.T) {
	q := Options{Directory: "/home/karasz/Downloads/Certs/PEM/"}
	z, err := NewStore(q)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = z.Delete(ctx, "www.example.com")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("deleted certificate")
}
*/
