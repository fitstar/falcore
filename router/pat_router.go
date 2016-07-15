package router

type PatternRouter struct {
	Patterns []Pattern
}

type Pattern struct {
	tokens []token
}

func ParsePattern(pat string) (*Pattern, error) {
	if t, e := tokenizePattern(pat); e == nil {
		return &Pattern{t}, nil
	} else {
		return nil, e
	}
}

func (p *Pattern) match(path string) (bool, map[string]string) {
	matches := make(map[string]string)

	for _, token := range p.tokens {
		if match, length := token.match(path); match {
			switch token.tokenType {
			case tokenCapture:
				matches[token.raw] = path[:length]
			}
			path = path[length:]
		} else {
			return false, nil
		}
	}
	if len(path) > 0 {
		return false, nil
	}

	return true, matches
}
