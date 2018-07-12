package utils

import (
	"mime"
	"strconv"
	"strings"
)

type parsedMimeType struct {
	majorType string
	minorType string
	options   map[string]string
	rating    float64
}

type mimeTypeMatch uint32

const (
	mimeTypeMatchNone mimeTypeMatch = iota
	mimeTypeMatchWildcard
	mimeTypeMatchMajor
	mimeTypeMatchExact
)

// Select from available types based on accept header from client
// FIXME: apply options
func PreferredContentType(acceptHeader []string, typesOffered []string) (string, map[string]string) {
	// Parse all the mime types
	acceptTypes := parseMimeTypes(acceptHeader)
	offers := parseMimeTypes(typesOffered)

	var bestType *parsedMimeType = nil
	var bestTypeRating float64 = 0
	var bestTypeMatch mimeTypeMatch = 0

	if len(acceptTypes) == 0 || len(typesOffered) == 0 {
		return "", nil
	}

	for _, offer := range offers {
		for _, accepted := range acceptTypes {
			match := offer.Match(accepted)
			if match == mimeTypeMatchNone {
				continue
			}

			if match > bestTypeMatch {
				// Match is more exact
				bestType = offer
				bestTypeRating = accepted.rating
				bestTypeMatch = match
			} else if match == bestTypeMatch && accepted.rating > bestTypeRating {
				// Is match 'q' bettter/
				bestType = offer
				bestTypeRating = accepted.rating
				bestTypeMatch = match
			}
		}
	}

	if bestType != nil {
		return strings.Join([]string{bestType.majorType, bestType.minorType}, "/"), bestType.options
	} else {
		return "", nil
	}
}

func parseMimeTypes(types []string) []*parsedMimeType {
	parsedTypes := make([]*parsedMimeType, 0, len(types))
	for _, t := range types {
		if typeName, o, e := mime.ParseMediaType(t); e == nil {
			typeParts := strings.Split(typeName, "/")
			if len(typeName) > 1 {
				majorType, minorType := typeParts[0], typeParts[1]

				var rating float64
				if f, err := strconv.ParseFloat(o["q"], 64); err == nil {
					rating = f
				} else {
					rating = 0.1
				}

				parsedTypes = append(parsedTypes, &parsedMimeType{majorType, minorType, o, rating})
			}
		}
	}
	return parsedTypes
}

func (a *parsedMimeType) Match(b *parsedMimeType) mimeTypeMatch {
	if a.majorType == b.majorType {
		if a.minorType == b.minorType {
			return mimeTypeMatchExact
		} else if a.minorType == "*" || b.minorType == "*" {
			return mimeTypeMatchMajor
		}
	} else if a.majorType == "*" || b.majorType == "*" {
		return mimeTypeMatchWildcard
	}
	return mimeTypeMatchNone
}
