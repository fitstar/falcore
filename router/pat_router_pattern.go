package router

type Pattern struct {
	raw    string
	tokens []token
}

func ParsePattern(pat string) (*Pattern, error) {
	if t, e := tokenizePattern(pat); e == nil {
		return &Pattern{pat, t}, nil
	} else {
		return nil, e
	}
}

func (p *Pattern) match(path string) (bool, map[string]string) {
	matches := make(map[string]string)

	var optional bool
	var optionalStartPath string

	for i := 0; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if match, length := token.match(path); match {
			switch token.tokenType {
			case tokenBeginOptional:
				optional = true
				optionalStartPath = path
			case tokenEndOptional:
				optional = false
			case tokenCapture:
				matches[token.raw] = path[:length]
			}
			path = path[length:]
		} else {
			if optional {
				// optional match failed.
				// ffwd to end of optional section and restart
				path = optionalStartPath
				optional = false
				for p.tokens[i].tokenType != tokenEndOptional {
					i++
				}
			} else {
				return false, nil
			}
		}
	}
	if len(path) > 0 {
		return false, nil
	}

	return true, matches
}
