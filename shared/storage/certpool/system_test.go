package certpool

import "testing"

func TestSystemCerts(t *testing.T) {
	if _, err := SystemCertPool(); err != nil {
		t.Error(err)
	}
}
