package qlist

import (
	"fmt"
	"net/http"
	"strings"
)

// This module provides basic functions for handling mime-types. It can handle
// matching mime-types against a list of media-ranges. See section 12.5.1 of
// the HTTP Semantics [RFC 9110] for a complete explanation.
//
// https://www.rfc-editor.org/rfc/rfc9110.html#section-12.5.1
//

const (
	// Accept is the canonical header name used for negotiating a MIME Type
	// for the content of the Request response
	Accept = "Accept"
)

// AsMediaRange converts a QualityValue into a valid MediaRange
func AsMediaRange(q QualityValue) (QualityValue, bool) {
	var value []string

	if !q.IsZero() {
		switch len(q.value) {
		case 0:
			value = []string{"*", "*"}
		case 1:
			value = []string{q.value[0], "*"}
		case 2:
			value = q.value
		default:
			return q, false
		}

		if value[0] == "" {
			value[0] = "*"
		}

		if value[1] == "" {
			value[1] = "*"
		}

		q.value = value
	}

	return q, true
}

// ParseMediaRangeHeader extract a accepted media ranges from the
// Accept header
func ParseMediaRangeHeader(hdr http.Header) (QualityList, error) {
	hdrs := hdr.Values(Accept)
	return ParseMediaRangeHeaders(hdrs)
}

// ParseMediaRangeHeaders extracts a [QualityList] from a list of Accept
// headers
func ParseMediaRangeHeaders(hdrs []string) (out QualityList, err error) {
	for _, s := range hdrs {
		r, err := ParseMediaRangeString(s)
		if err != nil {
			return out, err
		}
		out = append(out, r...)
	}
	return out, nil
}

// ParseMediaRangeString extracts a [QualityList] from a string representings on
// Accept header
func ParseMediaRangeString(ranges string) (out QualityList, err error) {
	for _, s := range strings.Split(ranges, ",") {
		r, err := ParseMediaRange(s)
		if err != nil {
			return out, err
		}

		if !r.IsZero() {
			out = append(out, r)
		}
	}

	return out, nil
}

// ParseMediaRange parses a Media-Range
func ParseMediaRange(mimerange string) (QualityValue, error) {
	var ok bool
	q, err := ParseQualityValue(mimerange)
	if err == nil {
		q, ok = AsMediaRange(q)
	}

	if !ok {
		err = fmt.Errorf("%q: invalid Media-Range", q)
		return q, err
	}

	return q, nil
}

// MediaRangeQuality returns the quality 'q' of a mime-type when compared
// against the media-ranges in ranges
func MediaRangeQuality(mimetype string, ranges ...string) (quality float32) {
	ql, _ := ParseMediaRangeHeaders(ranges)
	return MediaRangeQualityParsed(mimetype, ql)
}

// MediaRangeQualityParsed find the best match for a given mime-type against
// a list of media_ranges that have already been
// parsed by ParseMediaRange(). Returns the
// 'q' quality parameter of the best match, 0 if no
// match was found. This function bahaves the same as quality()
// except that 'parsed_ranges' must be a list of
// parsed media ranges.
func MediaRangeQualityParsed(mimetype string, parsedRanges QualityList) (quality float32) {
	_, quality = MediaRangeFitnessAndQuality(mimetype, parsedRanges)
	return quality
}

// MediaRangeFitnessAndQuality finds the best match for a given
// mime-type against a list of media_ranges that have
// already been parsed by ParseMediaRange(). Returns a tuple of
// the fitness value and the value of the 'q' quality
// parameter of the best match, or (-1, 0) if no match
// was found. Just as for QualityParsed(), 'parsedranges'
// must be a list of parsed media ranges.
func MediaRangeFitnessAndQuality(mimetype string,
	accepted QualityList) (fitness int, quality float32) {
	//
	target, _ := ParseMediaRange(mimetype)
	return FitnessAndQualityParsed(target, accepted)
}

// MediaRangeBestQuality takes a list of supported mime-types and finds the best
// match for all the media-ranges listed in header. The value of
// header must be a string that conforms to the format of the
// HTTP Accept: header. The value of 'supported' is a list of
// mime-types.
func MediaRangeBestQuality(supported []string, header string) string {
	parsedSupported, _ := ParseMediaRangeHeaders(supported)
	parsedHeader, _ := ParseMediaRangeString(header)
	best, _, _ := BestQualityParsed(parsedSupported, parsedHeader)
	return best
}
