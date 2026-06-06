package model

// POS represents a Japanese part of speech.
type POS int

const (
	Adj POS = iota
	Adverb
	Noun
	Verb
)

// Word holds a single vocabulary entry.
type Word struct {
	ID   int
	Word string
	POS  POS
}

// FavSentence represents a saved favorite phrase combination.
type FavSentence struct {
	FormerPOS  POS
	LatterPOS  POS
	Particle   string
	FormerWord string
	LatterWord string
}

func posName(pos POS) string {
	return map[POS]string{
		Adj:    "adjective",
		Adverb: "adverb",
		Noun:   "noun",
		Verb:   "verb",
	}[pos]
}

func posTableName(pos POS) string {
	return posName(pos) + "s"
}

func namePOS(name string) POS {
	return map[string]POS{
		"adjective": Adj,
		"adverb":    Adverb,
		"noun":      Noun,
		"verb":      Verb,
	}[name]
}
