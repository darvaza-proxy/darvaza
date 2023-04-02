package qlist

import (
	"runtime"
	"testing"
)

// revive:disable:argument-limit
func parsedEqual(test *testing.T, mime string,
	t string, st string, q float32, attrs map[string]string) {
	// revive:enable:argument-limit
	r, err := ParseMediaRange(mime)
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		test.Errorf("%s:%d Failed to parse: %s", file, line, err)
		test.FailNow()
	}
	if got := mediaRangeType(r); t != got {
		test.Errorf("%s:%d Failed to parse major type %q from %q, got %q",
			file, line, t, mime, got)
	}
	if got := mediaRangeSubType(r); st != got {
		test.Errorf("%s:%d Failed to parse minor type %q from %q, got %q",
			file, line, st, mime, got)
	}
	if q != r.Quality() {
		test.Errorf("%s:%d Failed to parse quality %v from %q, got %v",
			file, line, st, mime, r.Quality())
	}
	if !equalAttrs(attrs, r.attrs) {
		test.Errorf("%s:%d Failed to parse attributes, expected %v, got %v",
			file, line, attrs, r.attrs)
	}
}

func mediaRangeType(r QualityValue) string {
	var mtype string
	if len(r.value) > 0 {
		mtype = r.value[0]
	}

	if mtype == "" {
		mtype = "*"
	}

	return mtype
}

func mediaRangeSubType(r QualityValue) string {
	var stype string
	if len(r.value) > 1 {
		stype = r.value[1]
	}

	if stype == "" {
		stype = "*"
	}

	return stype
}

func equalAttrs(a, b map[string]string) bool {
	if len(a) == len(b) {
		for k, va := range a {
			vb, ok := b[k]
			if !ok || va != vb {
				return false
			}
		}
		return true
	}
	return false
}

func TestParseMimeType(t *testing.T) {
	parsedEqual(t, "Application/xhtml;q=0.5;vEr=1.2", "application", "xhtml",
		0.5, map[string]string{"ver": "1.2"})
}

func TestParseMediaRange(t *testing.T) {
	parsedEqual(t, "application/xml;q=1", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml;q=", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml;q", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml ; q=", "application", "xml", 1, nil)
	parsedEqual(t, "application/xml ; q=1;b=other", "application", "xml",
		1, map[string]string{"b": "other"})
	parsedEqual(t, "application/xml ; q=2;b=other", "application", "xml",
		1, map[string]string{"b": "other"})
	// Java URLConnection class sends an Accept header that includes a single *
	parsedEqual(t, " *;q=.2", "*", "*", .2, nil)
}

func TestRFC2616Example(t *testing.T) {
	// revive:disable:line-length-limit
	accept := "text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, * /*;q=0.5"
	// revive:enable:line-length-limit
	cond := map[string]float32{
		"text/html;level=1": 1.0,
		"text/html":         0.7,
		"text/plain":        0.3,
		"image/jpeg":        0.5,
		"text/html;level=2": 0.4,
		"text/html;level=3": 0.7,
	}
	for mime, q := range cond {
		if q != MediaRangeQuality(mime, accept) {
			t.Errorf("Failed to match %v at %f, got %f instead",
				mime, q, MediaRangeQuality(mime, accept))
		}
	}
}

func doTestBestMatch(t *testing.T, supported []string, headers map[string]string) {
	for header, result := range headers {
		match := MediaRangeBestQuality(supported, header)
		if match != result {
			t.Errorf("BestMatch(%v, %v) == %s, not %s\n", supported, header, match, result)
		}
	}
}

func TestBestMatch(t *testing.T) {
	supported := []string{"application/xml", "application/xbel+xml"}
	headers := map[string]string{
		"application/xbel+xml":      "application/xbel+xml",
		"application/xbel+xml; q=1": "application/xbel+xml",
		"application/xml; q=1":      "application/xml",
		"application/*; q=1":        "application/xml",
		"*/*":                       "application/xml",
	}
	doTestBestMatch(t, supported, headers)
}

func TestBestMatchDirect(t *testing.T) {
	supported := []string{"application/xbel+xml", "text/xml"}
	headers := map[string]string{
		"text/*;q=0.5,*/*; q=0.1":               "text/xml",
		"text/html,application/atom+xml; q=0.9": "",
	}
	doTestBestMatch(t, supported, headers)
}

func TestBestMatchAjax(t *testing.T) {
	// Common AJAX scenario
	supported := []string{"application/json", "text/html"}
	headers := map[string]string{
		"application/json, text/javascript, */*": "application/json",
		"application/json, text/html;q=0.9":      "application/json",
	}
	doTestBestMatch(t, supported, headers)
}

func TestSupportWildcards(t *testing.T) {
	supported := []string{"image/*", "application/xml"}
	headers := map[string]string{
		"image/png": "image/*",
		"image/*":   "image/*",
	}
	doTestBestMatch(t, supported, headers)
}
