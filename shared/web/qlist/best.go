package qlist

import "net/http"

// FindQuality searches a [QualityList] for a given entry and
// returns its Quality Value
func FindQuality(s string, ql QualityList) (float32, bool) {
	for _, qv := range ql {
		switch qv.Value() {
		case s, "*":
			// match
			return qv.quality, true
		}
	}
	// no match
	return 0.0, false
}

// BestQuality searches for the best option among supported values
// based on the provided QualityList
func BestQuality(supported []string, ql QualityList) (string, float32, bool) {
	bestOption := ""
	bestQuality := float32(0.0)

	if len(ql) > 0 {
		for _, option := range supported {
			quality, _ := FindQuality(option, ql)
			if quality > bestQuality {
				bestQuality = quality
				bestOption = option
			}
		}
	}

	return bestOption, bestQuality, bestOption != ""
}

// BestQualityWithIdentity searches for the best option among supported values
// based on the provided QualityList, but gives special treatment to an
// identity option which is used if it's the best or if nothing was chosen but
// the identity isn't explicitly forbidden
func BestQualityWithIdentity(supported []string,
	ql QualityList, identity string) (string, float32, bool) {
	// pick the best supported match
	bestOption, bestQuality, _ := BestQuality(supported, ql)

	if identity != "" {
		// test for identity
		quality, ok := FindQuality(identity, ql)
		switch {
		case quality > bestQuality:
			// identity is best
		case bestOption == "" && !ok:
			// nothing chosen, but identity wasn't forbidden
		default:
			// no luck with the identity
			goto done
		}

		bestOption = identity
		bestQuality = quality
	}

done:
	return bestOption, bestQuality, bestOption != ""
}

// BestEncoding chooses the best supported Content-Type option considering
// the Accept header
func BestEncoding(supported []string, hdr http.Header) (string, bool) {
	ql, _ := ParseQualityHeader(hdr, "Accept")
	best, _, ok := BestQualityWithIdentity(supported, ql, "identity")
	return best, ok
}
