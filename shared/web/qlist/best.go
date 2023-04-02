package qlist

import "net/http"

const (
	// AcceptEncoding is the canonical name given to the header used
	// to indicate compression options
	AcceptEncoding = "Accept-Encoding"
)

// FitnessAndQualityParsed finds the best accepted match
// for a target, returning fitness and quality.
// (-1, 0) if no match was found
func FitnessAndQualityParsed(target QualityValue,
	accepted QualityList) (fitness int, quality float32) {
	//
	bestfitness := -1
	bestquality := float32(0.0)

	for _, r := range accepted {
		if fitness := r.MatchFitness(target); fitness > 0 {
			if fitness > bestfitness {
				bestfitness = fitness
				bestquality = r.Quality()
			}
		}
	}

	return bestfitness, bestquality
}

// FitnessAndQuality finds the best accepted match
// for a target, returning fitness and quality.
// (-1, 0) if no match was found
func FitnessAndQuality(target string,
	accepted QualityList) (fitness int, quality float32) {
	q, err := ParseQualityValue(target)
	if err == nil {
		return FitnessAndQualityParsed(q, accepted)
	}
	return -1, 0.
}

// BestQualityParsed takes a list of supported and accepted quality entries
// and finds the best match among all combinations.
func BestQualityParsed(supported, accepted []QualityValue) (string, float32, bool) {
	var bestQuality float32
	var bestOption string

	for _, v := range supported {
		_, quality := FitnessAndQualityParsed(v, accepted)
		if quality > bestQuality {
			bestQuality = quality
			bestOption = v.String()
		}
	}

	return bestOption, bestQuality, bestOption != ""
}

// BestQuality searches for the best option among supported values
// based on the provided QualityList
func BestQuality(supported []string, ql QualityList) (string, float32, bool) {
	var sql QualityList

	for _, s := range supported {
		q, err := ParseQualityValue(s)
		if err != nil {
			sql = append(sql, q)
		}
	}

	return BestQualityParsed(sql, ql)
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
		fitness, quality := FitnessAndQuality(identity, ql)
		switch {
		case quality > bestQuality:
			// identity is best
		case bestOption == "" && fitness >= 0:
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

// BestEncoding chooses the best supported compression option considering
// the Accept-Encoding header
func BestEncoding(supported []string, hdr http.Header) (string, bool) {
	ql, _ := ParseQualityHeader(hdr, AcceptEncoding)
	best, _, ok := BestQualityWithIdentity(supported, ql, "identity")
	return best, ok
}
