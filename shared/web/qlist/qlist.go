// Package qlist provides a processor for HTTP Quality Lists
package qlist

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/darvaza-proxy/core"
)

const (
	MinimumQuality = 0.    // MinimumQuality is 0
	MaximumQuality = 1.    // MaximumQuality is 1
	Epsilon        = 0.001 // Epsilon is the accuracy of the quality, 1/100
)

// QualityValue is a parsed item of a [QualityList]
type QualityValue struct {
	value   []string
	quality float32
}

// IsZero tells if the QualityValue doesn't hold any information
func (q QualityValue) IsZero() bool {
	return len(q.value) == 0 && q.quality == 0
}

// Value tells the entry the quality refers to
func (q QualityValue) Value() string {
	switch len(q.value) {
	case 0:
		return ""
	case 1:
		return q.value[0]
	default:
		return strings.Join(q.value, "/")
	}
}

// Quality of the Value entry
func (q QualityValue) Quality() float32 {
	return q.quality
}

func (q QualityValue) String() string {
	if q.quality+Epsilon > MaximumQuality {
		return q.Value()
	}
	return fmt.Sprintf("%s;q=%v", q.Value(), q.quality)
}

// QualityList is a list of [QualityValue]
type QualityList []QualityValue

func (ql QualityList) String() string {
	s := make([]string, len(ql))

	for i, x := range ql {
		s[i] = x.String()
	}

	return strings.Join(s, ", ")
}

// ParseQualityHeader extracts a [QualityList] from the headers of a request
func ParseQualityHeader(hdr http.Header, name string) (out QualityList, err error) {
	for k, v := range hdr {
		if strings.EqualFold(name, k) {
			out, err = parseQualityHeaders(out, v)
			if err != nil {
				return out, err
			}
		}
	}
	return out, nil
}

func parseQualityHeaders(out QualityList, hdrs []string) (QualityList, error) {
	for _, s := range hdrs {
		q, err := ParseQualityString(s)
		if err != nil {
			return out, err
		}
		out = append(out, q...)
	}
	return out, nil
}

// ParseQualityString extracts a [QualityList] from a string representing one
// Header's content
func ParseQualityString(qlist string) (out QualityList, err error) {
	for _, s := range strings.Split(qlist, ",") {
		q, err := ParseQualityValue(s)
		if err != nil {
			return out, err
		}

		if !q.IsZero() {
			out = append(out, q)
		}
	}

	return out, nil
}

// ParseQualityValue parses one [QualityValue]
func ParseQualityValue(s string) (QualityValue, error) {
	var out QualityValue

	fields := splitFields(s)
	if len(fields) > 0 {
		// value
		v := strings.ToLower(fields[0])
		// attributes
		q, ok := qualityAttribute(fields[1:])

		if len(v) == 0 || !ok {
			err := fmt.Errorf("invalid argument: %q", s)
			return out, err
		}

		value := core.SliceReplaceFn(strings.Split(v, "/"),
			func(_ []string, s string) (string, bool) {
				s = strings.TrimSpace(s)
				return s, s != ""
			})

		out = QualityValue{
			value:   value,
			quality: q,
		}
	}

	return out, nil
}

func splitFields(s string) []string {
	fields := strings.Split(s, ";")
	fields = core.SliceReplaceFn(fields, func(partial []string, v string) (string, bool) {
		var keep bool

		// remove whitespace
		v = strings.TrimSpace(v)

		// remove empty attributes
		if len(partial) == 0 || len(v) > 0 {
			keep = true
		}

		return v, keep
	})

	return fields
}

func qualityAttribute(attrs []string) (float32, bool) {
	for _, s := range attrs {
		if strings.HasPrefix(s, "q=") {
			v := s[2:]
			q, err := strconv.ParseFloat(v, 32)
			switch {
			case v == "":
			case err != nil || math.IsNaN(q) || math.IsInf(q, 0):
				return 0., false
			case q < MinimumQuality+Epsilon:
				return MinimumQuality, true
			case q+Epsilon > MaximumQuality:
				return MaximumQuality, true
			default:
				q = math.Round(q/Epsilon) * Epsilon
				return float32(q), true
			}
		}
	}
	return MaximumQuality, true
}
