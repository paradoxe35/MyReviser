package history

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/paradoxe35/encre/internal/utils"
)

// Kind identifies where a history entry came from.
type Kind string

const (
	KindRevise    Kind = "revise"
	KindTranslate Kind = "translate"
	KindSpeech    Kind = "speech"
)

// MaxEntries bounds the file. The oldest entries are dropped when exceeded.
const MaxEntries = 200

// Entry is one finished action. The Original/Result pair is shared by every
// kind; language fields are only meaningful for translations.
type Entry struct {
	ID         string    `json:"id"`
	Kind       Kind      `json:"kind"`
	At         time.Time `json:"at"`
	Original   string    `json:"original"`
	Result     string    `json:"result"`
	FromLang   string    `json:"from_lang,omitempty"`
	ToLang     string    `json:"to_lang,omitempty"`
	Model      string    `json:"model,omitempty"`
	Provider   string    `json:"provider,omitempty"`
	Characters int       `json:"characters"`
}

// Store is an append-only JSONL log capped at MaxEntries. Writes are one
// line each; reads load the whole file, which stays small by construction.
type Store struct {
	mu   sync.Mutex
	path string

	onChange func()
}

func NewStore() *Store { return &Store{path: utils.AppHomeDir("history.jsonl")} }

// OnChange registers a callback fired after every successful append. It runs
// on the writing goroutine, outside the lock, so a handler may read the store
// back without deadlocking.
func (s *Store) OnChange(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onChange = fn
}

// Add appends an entry and trims the file to the cap. A failed write is
// logged and dropped: history must never break the action it records.
func (s *Store) Add(entry Entry) {
	var notify func()
	defer func() {
		if notify != nil {
			notify()
		}
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.At.IsZero() {
		entry.At = time.Now()
	}
	if entry.ID == "" {
		entry.ID = entry.At.Format("20060102150405.000000000")
	}

	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(entry); err != nil {
		return
	}
	s.trimLocked()
	notify = s.onChange
}

// Recent returns the newest entries first, filtered by kind. An empty kind
// means everything.
func (s *Store) Recent(kind Kind) []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var entry Entry
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		if kind == "" || entry.Kind == kind {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}

// Clear removes every entry.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.Remove(s.path)
}

// trimLocked rewrites the file without the oldest entries when over the cap.
// Called after each append; with the cap at 200 the rewrite amortises to one
// every MaxEntries appends.
func (s *Store) trimLocked() {
	info, err := os.Stat(s.path)
	if err != nil || info.Size() < 64*MaxEntries {
		return
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= MaxEntries {
		return
	}
	keep := strings.Join(lines[len(lines)-MaxEntries:], "\n") + "\n"
	_ = os.WriteFile(s.path, []byte(keep), 0o644)
}
