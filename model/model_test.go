package model

import (
	"database/sql"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database pre-loaded with test data.
func setupTestDB(t *testing.T) *DBHandler {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	schema := `
		CREATE TABLE adjectives (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE adverbs   (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE nouns     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE verbs     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE counts    (table_name VARCHAR(255), row_count INT);
		CREATE TABLE favorite_sentences (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			former_pos TEXT,
			latter_pos TEXT,
			particle   TEXT,
			former_word TEXT,
			latter_word TEXT
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	fixtures := `
		INSERT INTO adjectives VALUES (1, '美しい'), (2, '激しい');
		INSERT INTO adverbs   VALUES (1, '素早く'),  (2, '静かに');
		INSERT INTO nouns     VALUES (1, '猫'),      (2, '山');
		INSERT INTO verbs     VALUES (1, '走る'),    (2, '眠る');
		INSERT INTO counts VALUES
			('adjectives', 2),
			('adverbs',    2),
			('nouns',      2),
			('verbs',      2);
	`
	if _, err := db.Exec(fixtures); err != nil {
		t.Fatalf("failed to insert fixtures: %v", err)
	}

	sizeMap, err := wordSizeMap(db)
	if err != nil {
		t.Fatalf("failed to build sizeMap: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return &DBHandler{db: db, sizeMap: sizeMap}
}

// setupEmptyDB creates an in-memory SQLite database with no word rows.
func setupEmptyDB(t *testing.T) *DBHandler {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	schema := `
		CREATE TABLE adjectives (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE adverbs   (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE nouns     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE verbs     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE counts    (table_name VARCHAR(255), row_count INT);
		CREATE TABLE favorite_sentences (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			former_pos TEXT, latter_pos TEXT, particle TEXT,
			former_word TEXT, latter_word TEXT
		);
		INSERT INTO counts VALUES
			('adjectives', 0), ('adverbs', 0), ('nouns', 0), ('verbs', 0);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	sizeMap, err := wordSizeMap(db)
	if err != nil {
		t.Fatalf("failed to build sizeMap: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return &DBHandler{db: db, sizeMap: sizeMap}
}

// ---- GetWord tests ----

func TestGetWord(t *testing.T) {
	h := setupTestDB(t)

	tests := []struct {
		id   int
		pos  POS
		want string
	}{
		{1, Adj, "美しい"},
		{2, Adj, "激しい"},
		{1, Noun, "猫"},
		{2, Noun, "山"},
		{1, Adverb, "素早く"},
		{1, Verb, "走る"},
	}

	for _, tt := range tests {
		w, err := h.GetWord(tt.id, tt.pos)
		if err != nil {
			t.Errorf("GetWord(%d, %v) returned error: %v", tt.id, tt.pos, err)
			continue
		}
		if w.Word != tt.want {
			t.Errorf("GetWord(%d, %v).Word = %q, want %q", tt.id, tt.pos, w.Word, tt.want)
		}
		if w.POS != tt.pos {
			t.Errorf("GetWord(%d, %v).POS = %v, want %v", tt.id, tt.pos, w.POS, tt.pos)
		}
	}
}

func TestGetWord_NotFound(t *testing.T) {
	h := setupTestDB(t)

	_, err := h.GetWord(999, Noun)
	if err == nil {
		t.Error("GetWord(999, Noun) expected error for non-existent ID, got nil")
	}
}

// ---- GetRandWord tests ----

func TestGetRandWord(t *testing.T) {
	h := setupTestDB(t)

	for _, pos := range []POS{Adj, Adverb, Noun, Verb} {
		w, err := h.GetRandWord(pos)
		if err != nil {
			t.Errorf("GetRandWord(%v) returned error: %v", pos, err)
			continue
		}
		if w == nil {
			t.Errorf("GetRandWord(%v) returned nil word", pos)
			continue
		}
		if w.Word == "" {
			t.Errorf("GetRandWord(%v) returned empty word", pos)
		}
		if w.POS != pos {
			t.Errorf("GetRandWord(%v).POS = %v, want %v", pos, w.POS, pos)
		}
	}
}

func TestGetRandWord_EmptyTable(t *testing.T) {
	h := setupEmptyDB(t)

	for _, pos := range []POS{Adj, Adverb, Noun, Verb} {
		_, err := h.GetRandWord(pos)
		if err == nil {
			t.Errorf("GetRandWord(%v) expected error for empty table, got nil", pos)
		}
	}
}

func TestGetRandWord_ResultIsInTable(t *testing.T) {
	h := setupTestDB(t)

	validWords := map[POS][]string{
		Adj:    {"美しい", "激しい"},
		Adverb: {"素早く", "静かに"},
		Noun:   {"猫", "山"},
		Verb:   {"走る", "眠る"},
	}

	for pos, words := range validWords {
		for range 20 {
			w, err := h.GetRandWord(pos)
			if err != nil {
				t.Fatalf("GetRandWord(%v) unexpected error: %v", pos, err)
			}
			if !slices.Contains(words, w.Word) {
				t.Errorf("GetRandWord(%v) = %q, not in expected set %v", pos, w.Word, words)
			}
		}
	}
}

// ---- StoreSentence tests ----

func TestStoreSentence(t *testing.T) {
	h := setupTestDB(t)

	s := FavSentence{
		FormerPOS:  Noun,
		LatterPOS:  Verb,
		Particle:   "を",
		FormerWord: "猫",
		LatterWord: "食べる",
	}

	if err := h.StoreSentence(s); err != nil {
		t.Fatalf("StoreSentence() unexpected error: %v", err)
	}
}

func TestStoreSentence_AllValidParticles(t *testing.T) {
	h := setupTestDB(t)

	particles := []string{"を", "に", "が", "で", "と"}
	for _, p := range particles {
		s := FavSentence{
			FormerPOS:  Noun,
			LatterPOS:  Verb,
			Particle:   p,
			FormerWord: "猫",
			LatterWord: "走る",
		}
		if err := h.StoreSentence(s); err != nil {
			t.Errorf("StoreSentence(particle=%q) unexpected error: %v", p, err)
		}
	}
}

func TestStoreSentence_EmptyFormerWord(t *testing.T) {
	h := setupTestDB(t)

	s := FavSentence{
		FormerPOS:  Noun,
		LatterPOS:  Verb,
		Particle:   "を",
		FormerWord: "",
		LatterWord: "走る",
	}
	err := h.StoreSentence(s)
	if err == nil {
		t.Error("StoreSentence() expected error for empty FormerWord, got nil")
	}
}

func TestStoreSentence_EmptyLatterWord(t *testing.T) {
	h := setupTestDB(t)

	s := FavSentence{
		FormerPOS:  Noun,
		LatterPOS:  Verb,
		Particle:   "を",
		FormerWord: "猫",
		LatterWord: "",
	}
	err := h.StoreSentence(s)
	if err == nil {
		t.Error("StoreSentence() expected error for empty LatterWord, got nil")
	}
}

func TestStoreSentence_InvalidParticle(t *testing.T) {
	h := setupTestDB(t)

	invalidParticles := []string{"", "は", "の", "へ", "から", "invalid"}
	for _, p := range invalidParticles {
		s := FavSentence{
			FormerPOS:  Noun,
			LatterPOS:  Verb,
			Particle:   p,
			FormerWord: "猫",
			LatterWord: "走る",
		}
		err := h.StoreSentence(s)
		if err == nil {
			t.Errorf("StoreSentence(particle=%q) expected error for invalid particle, got nil", p)
		}
	}
}

// ---- GetSentences tests ----

func TestGetSentences(t *testing.T) {
	h := setupTestDB(t)

	sentences := []FavSentence{
		{Noun, Verb, "を", "猫", "食べる"},
		{Noun, Verb, "に", "山", "登る"},
		{Adj, Noun, "が", "空", "青い"},
	}
	for _, s := range sentences {
		if err := h.StoreSentence(s); err != nil {
			t.Fatalf("StoreSentence() setup failed: %v", err)
		}
	}

	got, err := h.GetSentences(1)
	if err != nil {
		t.Fatalf("GetSentences(1) unexpected error: %v", err)
	}
	if len(got) != len(sentences) {
		t.Errorf("GetSentences(1) returned %d sentences, want %d", len(got), len(sentences))
	}
}

func TestGetSentences_PageZero(t *testing.T) {
	h := setupTestDB(t)

	_, err := h.GetSentences(0)
	if err == nil {
		t.Error("GetSentences(0) expected error, got nil")
	}
}

func TestGetSentences_NegativePage(t *testing.T) {
	h := setupTestDB(t)

	_, err := h.GetSentences(-1)
	if err == nil {
		t.Error("GetSentences(-1) expected error, got nil")
	}
}

func TestGetSentences_EmptyPage(t *testing.T) {
	h := setupTestDB(t)

	got, err := h.GetSentences(1)
	if err != nil {
		t.Fatalf("GetSentences(1) on empty table unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetSentences(1) on empty table returned %d entries, want 0", len(got))
	}
}

func TestGetSentences_Pagination(t *testing.T) {
	h := setupTestDB(t)

	// Insert PageSize+1 sentences to span two pages.
	for i := range PageSize + 1 {
		s := FavSentence{
			FormerPOS:  Noun,
			LatterPOS:  Verb,
			Particle:   "を",
			FormerWord: "猫",
			LatterWord: "走る",
		}
		_ = i
		if err := h.StoreSentence(s); err != nil {
			t.Fatalf("StoreSentence() setup failed: %v", err)
		}
	}

	page1, err := h.GetSentences(1)
	if err != nil {
		t.Fatalf("GetSentences(1): %v", err)
	}
	if len(page1) != PageSize {
		t.Errorf("GetSentences(1) returned %d, want %d", len(page1), PageSize)
	}

	page2, err := h.GetSentences(2)
	if err != nil {
		t.Fatalf("GetSentences(2): %v", err)
	}
	if len(page2) != 1 {
		t.Errorf("GetSentences(2) returned %d, want 1", len(page2))
	}
}

// ---- Validate tests ----

func TestValidate_OK(t *testing.T) {
	h := setupTestDB(t)
	if err := h.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestValidate_EmptyDB(t *testing.T) {
	h := setupEmptyDB(t)
	if err := h.Validate(); err == nil {
		t.Error("Validate() expected error for empty database, got nil")
	}
}

// ---- wordSizeMap tests ----

func TestWordSizeMap(t *testing.T) {
	h := setupTestDB(t)

	for _, pos := range []POS{Adj, Adverb, Noun, Verb} {
		count, ok := h.sizeMap[pos]
		if !ok {
			t.Errorf("sizeMap missing entry for POS %v", pos)
			continue
		}
		if count != 2 {
			t.Errorf("sizeMap[%v] = %d, want 2", pos, count)
		}
	}
}

