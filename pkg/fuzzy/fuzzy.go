package fuzzy

type ColorizedText struct {
	Highlighted bool
	Text        string
}

func Search(s string, query string) bool {
	sRune := []rune(s)
	queryRune := []rune(query)

	if len(queryRune) == 0 {
		return true
	}

	if len(sRune) == 0 {
		return false
	}

	queryIdx := 0
	for i := range sRune {
		if sRune[i] == queryRune[queryIdx] {
			queryIdx++
			if queryIdx == len(queryRune) {
				return true
			}
		}
	}

	return false
}

func SearchColorized(s string, query string) ([]ColorizedText, bool) {
	results := make([]ColorizedText, 0)

	sRune := []rune(s)
	queryRune := []rune(query)

	if len(queryRune) == 0 {
		return []ColorizedText{{Highlighted: false, Text: string(sRune)}}, true
	}

	if len(sRune) == 0 {
		return []ColorizedText{}, false
	}

	queryIdx := 0
	for runeIdx := range sRune {
		if sRune[runeIdx] == queryRune[queryIdx] {
			results = append(results, ColorizedText{Highlighted: true, Text: string(sRune[runeIdx])})
			queryIdx++

			if queryIdx == len(queryRune) {
				results = append(results, ColorizedText{Highlighted: false, Text: string(sRune[runeIdx+1:])})

				return results, true
			}
		} else {
			results = append(results, ColorizedText{Highlighted: false, Text: string(sRune[runeIdx])})
		}
	}

	return results, false
}
