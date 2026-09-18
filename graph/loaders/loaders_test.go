package loaders

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"example/hello/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeFetcher struct {
	calls     atomic.Int32
	lastBatch []pgtype.UUID
	filesFor  map[pgtype.UUID][]database.File
	err       error
	mu        sync.Mutex
}

func (f *fakeFetcher) GetFilesByProjectIDs(_ context.Context, ids []pgtype.UUID) (map[pgtype.UUID][]database.File, error) {
	f.calls.Add(1)
	f.mu.Lock()
	f.lastBatch = append([]pgtype.UUID(nil), ids...)
	f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.filesFor, nil
}

func newID() pgtype.UUID { return pgtype.UUID{Bytes: uuid.New(), Valid: true} }

func TestFilesByProject_BatchesNLoadsIntoOneFetch(t *testing.T) {
	const n = 20

	ids := make([]pgtype.UUID, n)
	filesFor := map[pgtype.UUID][]database.File{}
	for i := range ids {
		ids[i] = newID()
		filesFor[ids[i]] = []database.File{{ID: newID(), ProjectID: ids[i], Name: "f"}}
	}
	fetcher := &fakeFetcher{filesFor: filesFor}
	l := New(fetcher, nil)

	// Resolvers for sibling projects run concurrently; mimic that.
	var wg sync.WaitGroup
	results := make([][]database.File, n)
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			files, err := l.FilesByProject.Load(context.Background(), ids[i])
			if err != nil {
				t.Errorf("load %d: %v", i, err)
			}
			results[i] = files
		}(i)
	}
	wg.Wait()

	if got := fetcher.calls.Load(); got != 1 {
		t.Fatalf("expected 1 batched fetch for %d loads, got %d", n, got)
	}
	if len(fetcher.lastBatch) != n {
		t.Fatalf("expected the single batch to carry %d keys, got %d", n, len(fetcher.lastBatch))
	}
	for i := range ids {
		if len(results[i]) != 1 || results[i][0].ProjectID != ids[i] {
			t.Fatalf("result %d was not the files of its own project", i)
		}
	}
}

func TestFilesByProject_ProjectWithNoFilesReturnsEmpty(t *testing.T) {
	id := newID()
	l := New(&fakeFetcher{filesFor: map[pgtype.UUID][]database.File{id: nil}}, nil)

	files, err := l.FilesByProject.Load(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %d", len(files))
	}
}

func TestFilesByProject_FetchErrorReachesEveryCaller(t *testing.T) {
	boom := errors.New("db down")
	l := New(&fakeFetcher{err: boom}, nil)

	a, b := newID(), newID()
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range []pgtype.UUID{a, b} {
		wg.Add(1)
		go func(i int, id pgtype.UUID) {
			defer wg.Done()
			_, errs[i] = l.FilesByProject.Load(context.Background(), id)
		}(i, id)
	}
	wg.Wait()

	for i, err := range errs {
		if err == nil {
			t.Fatalf("caller %d got no error", i)
		}
	}
}

func TestFrom_ReturnsNilWithoutMiddleware(t *testing.T) {
	if From(context.Background()) != nil {
		t.Fatal("expected nil loaders on a bare context")
	}
}

type fakeUsers struct {
	calls     atomic.Int32
	mu        sync.Mutex
	lastBatch []pgtype.UUID
	users     []database.User // returned in exactly this order, filtered to the requested IDs
	err       error
}

func (f *fakeUsers) GetUsersByIDs(_ context.Context, ids []pgtype.UUID) ([]database.User, error) {
	f.calls.Add(1)
	f.mu.Lock()
	f.lastBatch = append([]pgtype.UUID(nil), ids...)
	f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}

	wanted := map[pgtype.UUID]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	var out []database.User
	for _, u := range f.users {
		if wanted[u.ID] {
			out = append(out, u)
		}
	}
	return out, nil
}

func TestUserByID_BatchesNLoadsIntoOneFetchAndIgnoresRowOrder(t *testing.T) {
	const n = 20

	// The real query orders by username, not by the requested IDs - simulate
	// that by returning rows in an order unrelated to the keys.
	users := make([]database.User, n)
	for i := range users {
		users[i] = database.User{ID: newID(), Username: string(rune('a' + i))}
	}
	fetcher := &fakeUsers{users: users}
	l := New(nil, fetcher)

	var wg sync.WaitGroup
	got := make([]database.User, n)
	for i := range users {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u, err := l.UserByID.Load(context.Background(), users[i].ID)
			if err != nil {
				t.Errorf("load %d: %v", i, err)
			}
			got[i] = u
		}(i)
	}
	wg.Wait()

	if c := fetcher.calls.Load(); c != 1 {
		t.Fatalf("expected 1 batched fetch for %d loads, got %d", n, c)
	}
	for i := range users {
		if got[i].ID != users[i].ID || got[i].Username != users[i].Username {
			t.Fatalf("caller %d received the wrong user (%q)", i, got[i].Username)
		}
	}
}

func TestUserByID_MissingUserReturnsErrNotFound(t *testing.T) {
	present := database.User{ID: newID(), Username: "here"}
	l := New(nil, &fakeUsers{users: []database.User{present}})

	missing := newID()
	var wg sync.WaitGroup
	var okErr, missErr error
	var okUser database.User
	wg.Add(2)
	go func() { defer wg.Done(); okUser, okErr = l.UserByID.Load(context.Background(), present.ID) }()
	go func() { defer wg.Done(); _, missErr = l.UserByID.Load(context.Background(), missing) }()
	wg.Wait()

	if okErr != nil || okUser.ID != present.ID {
		t.Fatalf("existing user should load fine, got %v / %v", okUser.ID, okErr)
	}
	if !errors.Is(missErr, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for a missing user, got %v", missErr)
	}
}

func TestUserByID_SameUserRequestedManyTimesIsFetchedOnce(t *testing.T) {
	u := database.User{ID: newID(), Username: "repeat"}
	fetcher := &fakeUsers{users: []database.User{u}}
	l := New(nil, fetcher)

	// e.g. one person is the owner, a member, and the sender of 10 messages.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := l.UserByID.Load(context.Background(), u.ID); err != nil {
				t.Errorf("load: %v", err)
			}
		}()
	}
	wg.Wait()

	if c := fetcher.calls.Load(); c != 1 {
		t.Fatalf("expected 1 fetch, got %d", c)
	}
	if len(fetcher.lastBatch) != 1 {
		t.Fatalf("expected the key to appear once in the batch, got %d entries", len(fetcher.lastBatch))
	}
}

func TestUserByID_FetchErrorReachesCallers(t *testing.T) {
	boom := errors.New("db down")
	l := New(nil, &fakeUsers{err: boom})

	if _, err := l.UserByID.Load(context.Background(), newID()); !errors.Is(err, boom) {
		t.Fatalf("expected the fetch error, got %v", err)
	}
}
