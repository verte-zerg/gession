package main

type ColorizedText struct {
	Color int // 0 - normal, 1 - highlighted
	Text  string
}

type FuzzySearchResultRepresentation struct {
	results []ColorizedText
	bold    bool
}

func (f FuzzySearchResultRepresentation) GetString(bold bool) string {
	result := "  "
	if f.bold {
		result += "* "
	}
	boldMarker := ""
	if bold {
		boldMarker = BOLD
	}
	for _, r := range f.results {
		if r.Color == 0 {
			result += boldMarker + r.Text + RESET
		} else {
			result += boldMarker + GREEN + r.Text + RESET
		}
	}
	return result
}

func fuzzyFilterSessions(sessions []Session, query string) []Session {
	if query == "" {
		return sessions
	}
	result := make([]Session, 0)
	for _, session := range sessions {
		if fuzzySearch(session.Name, query) {
			result = append(result, session)
		}
	}
	return result
}

func fuzzyFilterSessionsColorized(sessions []Session, query string) []FuzzySearchResultRepresentation {
	result := make([]FuzzySearchResultRepresentation, 0)
	for _, session := range sessions {
		if representation, found := fuzzySearchColorized(session.Name, query); found {
			representation.bold = session.IsAttached
			result = append(result, representation)
		}
	}
	return result
}

func fuzzySearch(s string, query string) bool {
	sRune := []rune(s)
	queryRune := []rune(query)

	if len(queryRune) == 0 {
		return true
	}

	if len(sRune) == 0 {
		return false
	}

	queryIdx := 0
	for i := 0; i < len(sRune); i++ {
		if sRune[i] == queryRune[queryIdx] {
			queryIdx++
			if queryIdx == len(queryRune) {
				return true
			}
		}
	}

	return false
}

func fuzzySearchColorized(s string, query string) (FuzzySearchResultRepresentation, bool) {
	results := make([]ColorizedText, 0)

	sRune := []rune(s)
	queryRune := []rune(query)

	if len(queryRune) == 0 {
		return FuzzySearchResultRepresentation{results: []ColorizedText{{Color: 0, Text: string(sRune)}}}, true
	}

	if len(sRune) == 0 {
		return FuzzySearchResultRepresentation{results: []ColorizedText{}}, false
	}

	queryIdx := 0
	for i := 0; i < len(sRune); i++ {
		if sRune[i] == queryRune[queryIdx] {
			results = append(results, ColorizedText{Color: 1, Text: string(sRune[i])})
			queryIdx++
			if queryIdx == len(queryRune) {
				results = append(results, ColorizedText{Color: 0, Text: string(sRune[i+1:])})
				return FuzzySearchResultRepresentation{results: results}, true
			}
		} else {
			results = append(results, ColorizedText{Color: 0, Text: string(sRune[i])})
		}
	}

	return FuzzySearchResultRepresentation{results: results}, false
}
