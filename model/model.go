package model

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	_ "github.com/mattn/go-sqlite3"
)

const PageSize = 100

// validParticles is the set of allowed Japanese case particles for noun–verb phrases.
var validParticles = map[string]bool{
	"を": true,
	"に": true,
	"が": true,
	"で": true,
	"と": true,
}

// OpenDB opens the SQLite database at the given path and returns a DBHandler.
func OpenDB(dbPath string) (*DBHandler, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return OpenDBFromConn(db)
}

// OpenDBFromConn wraps an existing *sql.DB connection in a DBHandler.
// The caller retains ownership of db; Close must still be called on the handler.
func OpenDBFromConn(db *sql.DB) (*DBHandler, error) {
	sizeMap, err := wordSizeMap(db)
	if err != nil {
		return nil, err
	}

	return &DBHandler{db: db, sizeMap: sizeMap}, nil
}

// DBHandler wraps a database connection and word count metadata.
type DBHandler struct {
	db      *sql.DB
	sizeMap map[POS]int
}

// Close closes the underlying database connection.
func (h *DBHandler) Close() {
	h.db.Close()
}

// Validate checks that every part of speech has at least one word registered.
// Returns an error listing any empty tables so callers can fail fast at startup.
func (h *DBHandler) Validate() error {
	empty := []string{}
	for _, pos := range []POS{Adj, Adverb, Noun, Verb} {
		if h.sizeMap[pos] == 0 {
			empty = append(empty, posTableName(pos))
		}
	}
	if len(empty) > 0 {
		return fmt.Errorf("word tables have no data: %v — rebuild the database (see db/setup-db.bash)", empty)
	}
	return nil
}

// GetWord returns the word at the given ID for the given part of speech.
func (h *DBHandler) GetWord(id int, pos POS) (*Word, error) {
	row := h.db.QueryRow(
		fmt.Sprintf(`SELECT id, word FROM %s WHERE id=?`, posTableName(pos)),
		id,
	)

	w := &Word{POS: pos}
	if err := row.Scan(&w.ID, &w.Word); err != nil {
		return nil, err
	}

	return w, nil
}

// GetRandWord returns a randomly selected word for the given part of speech.
func (h *DBHandler) GetRandWord(pos POS) (*Word, error) {
	count, ok := h.sizeMap[pos]
	if !ok || count == 0 {
		return nil, fmt.Errorf("no words registered for part of speech %q", posName(pos))
	}
	id := rand.Intn(count) + 1
	return h.GetWord(id, pos)
}

// StoreSentence inserts a favorite sentence into the database.
func (h *DBHandler) StoreSentence(s FavSentence) error {
	if s.FormerWord == "" {
		return fmt.Errorf("former_word must not be empty")
	}
	if s.LatterWord == "" {
		return fmt.Errorf("latter_word must not be empty")
	}
	if !validParticles[s.Particle] {
		return fmt.Errorf("invalid particle %q: must be one of を・に・が・で・と", s.Particle)
	}

	_, err := h.db.Exec(
		`INSERT INTO favorite_sentences (former_pos, latter_pos, particle, former_word, latter_word) VALUES (?, ?, ?, ?, ?)`,
		posName(s.FormerPOS),
		posName(s.LatterPOS),
		s.Particle,
		s.FormerWord,
		s.LatterWord,
	)
	return err
}

// GetSentences returns paginated favorite sentences (1-indexed pages).
func (h *DBHandler) GetSentences(page int) ([]FavSentence, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1, got %d", page)
	}

	sentences := []FavSentence{}

	minID := (page-1)*PageSize + 1
	maxID := page * PageSize
	rows, err := h.db.Query(
		`SELECT id, former_pos, latter_pos, particle, former_word, latter_word FROM favorite_sentences WHERE id >= ? AND id <= ?`,
		minID, maxID,
	)
	if err != nil {
		return sentences, err
	}
	defer rows.Close()

	for rows.Next() {
		s := FavSentence{}
		var id int
		var former, latter string
		if err := rows.Scan(&id, &former, &latter, &s.Particle, &s.FormerWord, &s.LatterWord); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		s.FormerPOS = namePOS(former)
		s.LatterPOS = namePOS(latter)
		sentences = append(sentences, s)
	}

	return sentences, rows.Err()
}

func wordSizeMap(db *sql.DB) (map[POS]int, error) {
	sizeMap := map[POS]int{}

	rows, err := db.Query(`SELECT table_name, row_count FROM counts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var n int
		if err := rows.Scan(&name, &n); err != nil {
			return nil, err
		}
		// strip the trailing "s" from table names (e.g. "adjectives" -> "adjective")
		sizeMap[namePOS(name[:len(name)-1])] = n
	}

	return sizeMap, rows.Err()
}
