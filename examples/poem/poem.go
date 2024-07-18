package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jensilo/gwu"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"sync"
)

const (
	IDCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	IDLength  = 10
)

var (
	// ErrNotFound for external use, safe to display to the client.
	ErrNotFound = errors.New("poem(s) do(es) not exist")
	// ErrAuthorNotFound for external use, safe to display to the client.
	ErrAuthorNotFound = errors.New("the requested author does not exist")
	// ErrCouldNotCreate for external use, safe to display to the client.
	ErrCouldNotCreate = errors.New("internal error: could not create")

	// errNotFound to simulate some internal, application specific error.
	errNotFound = errors.New("not found - internally")
	// errDuplicate to simulate some internal, application specific error.
	errDuplicate = errors.New("duplicate poem - internally")
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	store := NewStore()
	ctrl := PoemController{store: store}

	mux := http.NewServeMux()
	mux.Handle("GET /poem/{id}", gwu.Handle(IDIn("id"), ctrl.ByID,
		gwu.Log(log.With("method", "GET", "route", "/poem/{id}"))),
	)
	mux.Handle("GET /poems", gwu.Handle(gwu.Empty(), ctrl.All,
		gwu.Log(log.With("method", "GET", "route", "/poems"))),
	)
	mux.Handle("POST /poem", gwu.Handle(gwu.JSON[Poem](), gwu.ValIn(ctrl.Create, ValidateToCreate),
		gwu.Log(log.With("method", "POST", "route", "/poem"))),
	)
	mux.Handle("GET /poems/author/{author}", gwu.Handle(gwu.PathVal("author"), ctrl.ByAuthor,
		gwu.Log(log.With("method", "GET", "route", "/poems/author/{author}"))),
	)

	server := http.Server{Addr: ":8080", Handler: mux}

	log.Info("start server...")
	log.Info("server killed", "error", server.ListenAndServe())
}

type ID string

func NewID() ID {
	b := make([]byte, IDLength)
	for i := range b {
		b[i] = IDCharset[rand.Intn(len(IDCharset))]
	}

	return ID(b)
}

func IDIn(key string) gwu.CnIn[ID] {
	return func(r *http.Request, _ gwu.HandleOpts) (ID, error) {
		return ID(r.PathValue(key)), nil
	}
}

type Poem struct {
	ID     ID     `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

func ValidateToCreate(p Poem) error {
	reqErr := func(key string) error { return fmt.Errorf("%s required to create poem", key) }

	if p.Name == "" {
		return reqErr("name")
	}

	if p.Author == "" {
		return reqErr("author")
	}

	if p.Text == "" {
		return reqErr("text")
	}

	return nil
}

type Store struct {
	poems map[ID]Poem
	mu    sync.RWMutex
}

func NewStore() *Store {
	store := &Store{poems: make(map[ID]Poem)}
	store.mock()

	return store
}

func (s *Store) Poem(id ID) (Poem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	poem, exists := s.poems[id]
	if !exists {
		return poem, errNotFound
	}

	return poem, nil
}

func (s *Store) PoemsByAuthor(author string) []Poem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	poems := make([]Poem, 0)
	for _, poem := range s.poems {
		if poem.Author == author {
			poems = append(poems, poem)
		}
	}

	return poems
}

func (s *Store) Add(poem Poem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.poems[poem.ID]
	if exists {
		return errDuplicate
	}

	s.poems[poem.ID] = poem

	return nil
}

func (s *Store) All() []Poem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	poems := make([]Poem, 0, len(s.poems))
	for _, poem := range s.poems {
		poems = append(poems, poem)
	}

	return poems
}

type PoemController struct {
	store *Store
}

func (c *PoemController) Create(_ context.Context, poem Poem, opts gwu.HandleOpts) (Poem, int, error) {
	poem.ID = NewID()
	err := c.store.Add(poem)
	if err != nil {
		opts.Log.Debug("could not create poem", "error", err, "poem", poem)
		return poem, http.StatusInternalServerError, ErrCouldNotCreate
	}

	return poem, http.StatusCreated, nil
}

func (c *PoemController) ByID(_ context.Context, id ID, opts gwu.HandleOpts) (Poem, int, error) {
	poem, err := c.store.Poem(id)
	if err != nil {
		opts.Log.Debug("requested non-existent poem", "id", id)
		return poem, http.StatusNotFound, ErrNotFound
	}

	return poem, http.StatusOK, nil
}

func (c *PoemController) All(_ context.Context, _ any, opts gwu.HandleOpts) ([]Poem, int, error) {
	poems := c.store.All()
	return poems, http.StatusOK, nil
}

func (c *PoemController) ByAuthor(_ context.Context, author string, opts gwu.HandleOpts) ([]Poem, int, error) {
	poems := c.store.PoemsByAuthor(author)
	if len(poems) == 0 {
		opts.Log.Debug("no poems found for author", "author", author)
		return nil, http.StatusNotFound, ErrAuthorNotFound
	}

	return poems, http.StatusOK, nil
}

func (s *Store) mock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.poems["1234567890"] = Poem{
		ID:     "1234567890",
		Name:   "The Raven",
		Author: "Edgar Allan Poe",
		Text: `Once upon a midnight dreary, while I pondered, weak and weary,
Over many a quaint and curious volume of forgotten lore—
While I nodded, nearly napping, suddenly there came a tapping,
As of some one gently rapping, rapping at my chamber door.
“’Tis some visitor,” I muttered, “tapping at my chamber door—
Only this and nothing more.”`,
	}

	s.poems["abc123defx"] = Poem{
		ID:     "abc123defx",
		Name:   "The Road Not Taken",
		Author: "Robert Frost",
		Text: `Two roads diverged in a yellow wood,
And sorry I could not travel both
And be one traveler, long I stood
And looked down one as far as I could
To where it bent in the undergrowth;`,
	}

	s.poems["abcdefghij"] = Poem{
		ID:     "abcdefghi",
		Name:   "Der Erlkönig",
		Author: "Goethe",
		Text: `Wer reitet so spät durch Nacht und Wind?
Es ist der Vater mit seinem Kind;
Er hat den Knaben wohl in dem Arm,
Er faßt ihn sicher, er hält ihn warm.`,
	}

	s.poems["isjzB57elf"] = Poem{
		ID:     "isjzB57elf",
		Name:   "Der Zauberlehrling",
		Author: "Goethe",
		Text: `Hat der alte Hexenmeister
Sich doch einmal wegbegeben!
Und nun sollen seine Geister
Auch nach meinem Willen leben.`,
	}
}
