// Package qlist provides a processor for HTTP Quality Lists
package qlist

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"darvaza.org/core"
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
	attrs   map[string]string
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

// Get retrieves the value of an attribute
func (q QualityValue) Get(attr string) (string, bool) {
	switch {
	case attr == "q":
		return fmt.Sprint(q.quality), true
	case len(q.attrs) == 0:
		return "", false
	default:
		v, ok := q.attrs[attr]
		return v, ok
	}
}

func (q QualityValue) String() string {
	var s []string

	s = append(s, q.Value())

	if MaximumQuality > q.quality+Epsilon {
		s = append(s, fmt.Sprintf("q=%v", q.quality))
	}

	for k, v := range q.attrs {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(s, ";")
}

// Match answers the question if we match a target.
// For an entry to match another it needs to have the
// same number of parts, and each pair of parts be
// identical or at least one of them a "*" wildcard
func (q QualityValue) Match(t QualityValue) bool {
	if l := len(q.value); l == len(t.value) {
		for i := 0; i < l; i++ {
			a, b := q.value[i], t.value[i]
			if a != "*" && b != "*" && a != b {
				return false
			}
		}
		return true
	}

	return false
}

// MatchFitness answers the question of how well we match a target.
// simple match gives one point, as do matching target attributes.
// exact matches at the lowest part gives 10, and that grows geometrically
// as parts we move up in the chain
func (q QualityValue) MatchFitness(t QualityValue) int {
	fitness := 0
	if q.Match(t) {
		fitness++
		fitness += matchValues(q.value, t.value)
		fitness += matchAttributes(q.attrs, t.attrs)
	}

	return fitness
}

func matchValues(qv, tv []string) int {
	fitness := 0
	l := len(qv)
	w := 10 * l

	for i := 0; i < l; i++ {
		if qv[i] == tv[i] {
			fitness += w
		}
		w /= 10
	}
	return fitness
}

func matchAttributes(qa, ta map[string]string) int {
	fitness := 0
	for key, tav := range ta {
		if av, ok := qa[key]; ok {
			if av == tav {
				fitness++
			}
		}
	}
	return fitness
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
	hdrs := hdr.Values(name)
	return parseQualityHeaders(out, hdrs)
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
		q, m, ok := parseAttributes(fields[1:])

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
			attrs:   m,
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

func parseAttributes(attrs []string) (q float32, m map[string]string, ok bool) {
	q = MaximumQuality
	ok = true

	for _, attr := range attrs {
		if attr != "" {
			q, m, ok = parseAttribute(attr, q, m)

			if !ok {
				break
			}
		}
	}

	return q, m, ok
}

func parseAttribute(attr string,
	prev float32, m map[string]string) (float32, map[string]string, bool) {
	//
	q, ok := prev, false
	k, v := splitAttribute(attr)
	switch k {
	case "":
	case "q":
		if v != "" {
			// parse quality
			q, ok = parseQuality(v)
		} else {
			// skip empty quality
			ok = true
		}
	default:
		// regular
		if m == nil {
			m = make(map[string]string, 1)
		}
		ok = true
		m[k] = v
	}

	return q, m, ok
}

func splitAttribute(s string) (key, value string) {
	// find delimiter
	i := strings.IndexFunc(s, func(r rune) bool {
		return r == '='
	})

	// split
	switch {
	case i < 0:
		key = s
	default:
		key, value = s[:i], s[i+1:]
	}

	// sanitise
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.ToLower(strings.TrimSpace(value))
	return key, value
}

func parseQuality(v string) (float32, bool) {
	ok := true
	q, err := strconv.ParseFloat(v, 32)
	switch {
	case err != nil || math.IsNaN(q) || math.IsInf(q, 0):
		ok = false
	case q < MinimumQuality+Epsilon:
		q = MinimumQuality
	case q+Epsilon > MaximumQuality:
		q = MaximumQuality
	default:
		q = math.Round(q/Epsilon) * Epsilon
	}

	if !ok {
		return 0, false
	}
	return float32(q), true
}
