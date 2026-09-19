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

func newID() pgtype.UUID { return pgtype.UUID{Bytes: uuid.New(), Valid: true} }

func ids(n int) []pgtype.UUID {
	out := make([]pgtype.UUID, n)
	for i := range out {
		out[i] = newID()
	}
	return out
}

// spy records how a batch function was called.
type spy struct {
	calls     atomic.Int32
	mu        sync.Mutex
	lastBatch []pgtype.UUID
}

func (s *spy) record(batch []pgtype.UUID) {
	s.calls.Add(1)
	s.mu.Lock()
	s.lastBatch = append([]pgtype.UUID(nil), batch...)
	s.mu.Unlock()
}

// oneRowPerKey is a Grouped source returning exactly one row per requested ID,
// built by mk, so a test can check each caller got its own parent's rows.
func oneRowPerKey[T any](s *spy, mk func(id pgtype.UUID) T) Grouped[T] {
	return func(_ context.Context, batch []pgtype.UUID) (map[pgtype.UUID][]T, error) {
		s.record(batch)
		m := make(map[pgtype.UUID][]T, len(batch))
		for _, id := range batch {
			m[id] = []T{mk(id)}
		}
		return m, nil
	}
}

// rowsInReverse is a Rows source that returns the requested rows in the
// *reverse* of the requested order (like SQL ordering by something unrelated to
// the keys) and omits any ID it has no row for.
func rowsInReverse[T any](s *spy, have map[pgtype.UUID]T) Rows[T] {
	return func(_ context.Context, batch []pgtype.UUID) ([]T, error) {
		s.record(batch)
		var out []T
		for i := len(batch) - 1; i >= 0; i-- {
			if row, ok := have[batch[i]]; ok {
				out = append(out, row)
			}
		}
		return out, nil
	}
}

// loadAll runs one Load per key concurrently (sibling resolvers run
// concurrently in a real operation) and returns results in key order.
func loadAll[T any](t *testing.T, name string, l *ListLoader[T], keys []pgtype.UUID) [][]T {
	t.Helper()
	out := make([][]T, len(keys))
	var wg sync.WaitGroup
	for i := range keys {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := l.Load(context.Background(), keys[i])
			if err != nil {
				t.Errorf("%s: load %d: %v", name, i, err)
			}
			out[i] = v
		}(i)
	}
	wg.Wait()
	return out
}

// ---------------------------------------------------------------------------
// List loaders
// ---------------------------------------------------------------------------

// Every list loader must (a) fold N concurrent Loads into ONE fetch and
// (b) hand each caller only its own parent's rows.
func TestListLoaders_BatchNLoadsIntoOneFetchAndKeepParentsApart(t *testing.T) {
	const n = 15
	keys := ids(n)

	cases := []struct {
		name string
		run  func(t *testing.T, s *spy)
	}{
		{"FilesByProject", func(t *testing.T, s *spy) {
			l := New(Sources{FilesByProject: oneRowPerKey(s, func(id pgtype.UUID) database.File { return database.File{ProjectID: id} })})
			for i, rows := range loadAll(t, "FilesByProject", l.FilesByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "FilesByProject", i)
			}
		}},
		{"FoldersByProject", func(t *testing.T, s *spy) {
			l := New(Sources{FoldersByProject: oneRowPerKey(s, func(id pgtype.UUID) database.Folder { return database.Folder{ProjectID: id} })})
			for i, rows := range loadAll(t, "FoldersByProject", l.FoldersByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "FoldersByProject", i)
			}
		}},
		{"MembersByProject", func(t *testing.T, s *spy) {
			l := New(Sources{MembersByProject: oneRowPerKey(s, func(id pgtype.UUID) database.ProjectMember { return database.ProjectMember{ProjectID: id} })})
			for i, rows := range loadAll(t, "MembersByProject", l.MembersByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "MembersByProject", i)
			}
		}},
		{"ChatsByProject", func(t *testing.T, s *spy) {
			l := New(Sources{ChatsByProject: oneRowPerKey(s, func(id pgtype.UUID) database.Chat { return database.Chat{ProjectID: id} })})
			for i, rows := range loadAll(t, "ChatsByProject", l.ChatsByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "ChatsByProject", i)
			}
		}},
		{"ReportsByProject", func(t *testing.T, s *spy) {
			l := New(Sources{ReportsByProject: oneRowPerKey(s, func(id pgtype.UUID) database.GetReportsByProjectIDsRow {
				return database.GetReportsByProjectIDsRow{ProjectID: id}
			})})
			for i, rows := range loadAll(t, "ReportsByProject", l.ReportsByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "ReportsByProject", i)
			}
		}},
		{"FlowchartsByProject", func(t *testing.T, s *spy) {
			l := New(Sources{FlowchartsByProject: oneRowPerKey(s, func(id pgtype.UUID) database.Flowchart { return database.Flowchart{ProjectID: id} })})
			for i, rows := range loadAll(t, "FlowchartsByProject", l.FlowchartsByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "FlowchartsByProject", i)
			}
		}},
		{"ActivityByProject", func(t *testing.T, s *spy) {
			l := New(Sources{ActivityByProject: oneRowPerKey(s, func(id pgtype.UUID) database.ActivityLog { return database.ActivityLog{ProjectID: id} })})
			for i, rows := range loadAll(t, "ActivityByProject", l.ActivityByProject, keys) {
				expect(t, len(rows) == 1 && rows[0].ProjectID == keys[i], "ActivityByProject", i)
			}
		}},
		{"MessagesByChat", func(t *testing.T, s *spy) {
			l := New(Sources{MessagesByChat: oneRowPerKey(s, func(id pgtype.UUID) database.ChatMessage { return database.ChatMessage{ChatID: id} })})
			for i, rows := range loadAll(t, "MessagesByChat", l.MessagesByChat, keys) {
				expect(t, len(rows) == 1 && rows[0].ChatID == keys[i], "MessagesByChat", i)
			}
		}},
		{"ParticipantsByChat", func(t *testing.T, s *spy) {
			l := New(Sources{ParticipantsByChat: oneRowPerKey(s, func(id pgtype.UUID) database.ChatParticipant { return database.ChatParticipant{ChatID: id} })})
			for i, rows := range loadAll(t, "ParticipantsByChat", l.ParticipantsByChat, keys) {
				expect(t, len(rows) == 1 && rows[0].ChatID == keys[i], "ParticipantsByChat", i)
			}
		}},
		{"SharesByFile", func(t *testing.T, s *spy) {
			l := New(Sources{SharesByFile: oneRowPerKey(s, func(id pgtype.UUID) database.FileShare { return database.FileShare{FileID: id} })})
			for i, rows := range loadAll(t, "SharesByFile", l.SharesByFile, keys) {
				expect(t, len(rows) == 1 && rows[0].FileID == keys[i], "SharesByFile", i)
			}
		}},
		{"FilesByFolder", func(t *testing.T, s *spy) {
			l := New(Sources{FilesByFolder: oneRowPerKey(s, func(id pgtype.UUID) database.File { return database.File{FolderID: id} })})
			for i, rows := range loadAll(t, "FilesByFolder", l.FilesByFolder, keys) {
				expect(t, len(rows) == 1 && rows[0].FolderID == keys[i], "FilesByFolder", i)
			}
		}},
		{"ChildFoldersByParent", func(t *testing.T, s *spy) {
			l := New(Sources{ChildFoldersByParent: oneRowPerKey(s, func(id pgtype.UUID) database.Folder { return database.Folder{ParentFolderID: id} })})
			for i, rows := range loadAll(t, "ChildFoldersByParent", l.ChildFoldersByParent, keys) {
				expect(t, len(rows) == 1 && rows[0].ParentFolderID == keys[i], "ChildFoldersByParent", i)
			}
		}},
		{"OwnedProjectsByOwner", func(t *testing.T, s *spy) {
			l := New(Sources{OwnedProjectsByOwner: oneRowPerKey(s, func(id pgtype.UUID) database.Project { return database.Project{OwnerID: id} })})
			for i, rows := range loadAll(t, "OwnedProjectsByOwner", l.OwnedProjectsByOwner, keys) {
				expect(t, len(rows) == 1 && rows[0].OwnerID == keys[i], "OwnedProjectsByOwner", i)
			}
		}},
		{"MembershipsByUser", func(t *testing.T, s *spy) {
			l := New(Sources{MembershipsByUser: oneRowPerKey(s, func(id pgtype.UUID) database.ProjectMember { return database.ProjectMember{UserID: id} })})
			for i, rows := range loadAll(t, "MembershipsByUser", l.MembershipsByUser, keys) {
				expect(t, len(rows) == 1 && rows[0].UserID == keys[i], "MembershipsByUser", i)
			}
		}},
		{"SharesBySharedWith", func(t *testing.T, s *spy) {
			l := New(Sources{SharesBySharedWith: oneRowPerKey(s, func(id pgtype.UUID) database.FileShare { return database.FileShare{SharedWith: id} })})
			for i, rows := range loadAll(t, "SharesBySharedWith", l.SharesBySharedWith, keys) {
				expect(t, len(rows) == 1 && rows[0].SharedWith == keys[i], "SharesBySharedWith", i)
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &spy{}
			tc.run(t, s)
			if got := s.calls.Load(); got != 1 {
				t.Fatalf("expected 1 batched fetch for %d loads, got %d", n, got)
			}
			if len(s.lastBatch) != n {
				t.Fatalf("expected the single batch to carry %d keys, got %d", n, len(s.lastBatch))
			}
		})
	}
}

func expect(t *testing.T, ok bool, name string, i int) {
	t.Helper()
	if !ok {
		t.Errorf("%s: caller %d did not receive its own parent's rows", name, i)
	}
}

func TestListLoader_ParentWithNoRowsLoadsAsEmptyNotError(t *testing.T) {
	id := newID()
	l := New(Sources{FilesByProject: func(context.Context, []pgtype.UUID) (map[pgtype.UUID][]database.File, error) {
		return map[pgtype.UUID][]database.File{id: nil}, nil
	}})

	rows, err := l.FilesByProject.Load(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no rows, got %d", len(rows))
	}
}

func TestListLoader_FetchErrorReachesEveryCaller(t *testing.T) {
	boom := errors.New("db down")
	l := New(Sources{MessagesByChat: func(context.Context, []pgtype.UUID) (map[pgtype.UUID][]database.ChatMessage, error) {
		return nil, boom
	}})

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range ids(2) {
		wg.Add(1)
		go func(i int, id pgtype.UUID) {
			defer wg.Done()
			_, errs[i] = l.MessagesByChat.Load(context.Background(), id)
		}(i, id)
	}
	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, boom) {
			t.Fatalf("caller %d: expected the fetch error, got %v", i, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Lookup loaders
// ---------------------------------------------------------------------------

// Every lookup loader must batch, be immune to the SQL row order, and report
// an ID with no row as ErrNotFound (never someone else's row).
func TestLookupLoaders_BatchIgnoreRowOrderAndReportMissing(t *testing.T) {
	const n = 12

	type lookup struct {
		name string
		// run loads n existing keys plus one missing key and returns the
		// per-key "got the right row" flags and the missing key's error.
		run func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) (right []bool, missErr error)
	}

	run := func(name string, f func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error)) lookup {
		return lookup{name, f}
	}

	cases := []lookup{
		run("UserByID", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.User{}
			for _, k := range keys {
				have[k] = database.User{ID: k}
			}
			l := New(Sources{Users: rowsInReverse(s, have)})
			return check(t, l.UserByID, keys, missing, func(u database.User, k pgtype.UUID) bool { return u.ID == k })
		}),
		run("ProjectByID", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.Project{}
			for _, k := range keys {
				have[k] = database.Project{ID: k}
			}
			l := New(Sources{Projects: rowsInReverse(s, have)})
			return check(t, l.ProjectByID, keys, missing, func(p database.Project, k pgtype.UUID) bool { return p.ID == k })
		}),
		run("FolderByID", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.Folder{}
			for _, k := range keys {
				have[k] = database.Folder{ID: k}
			}
			l := New(Sources{Folders: rowsInReverse(s, have)})
			return check(t, l.FolderByID, keys, missing, func(f database.Folder, k pgtype.UUID) bool { return f.ID == k })
		}),
		run("ProfileByUser", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.UserProfile{}
			for _, k := range keys {
				have[k] = database.UserProfile{UserID: k}
			}
			l := New(Sources{Profiles: rowsInReverse(s, have)})
			return check(t, l.ProfileByUser, keys, missing, func(p database.UserProfile, k pgtype.UUID) bool { return p.UserID == k })
		}),
		run("AIMetadataByFile", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.FileAiMetadatum{}
			for _, k := range keys {
				have[k] = database.FileAiMetadatum{FileID: k}
			}
			l := New(Sources{AIMetadata: rowsInReverse(s, have)})
			return check(t, l.AIMetadataByFile, keys, missing, func(m database.FileAiMetadatum, k pgtype.UUID) bool { return m.FileID == k })
		}),
		run("StorageByFile", func(s *spy, keys []pgtype.UUID, missing pgtype.UUID) ([]bool, error) {
			have := map[pgtype.UUID]database.FileStorage{}
			for _, k := range keys {
				have[k] = database.FileStorage{FileID: k}
			}
			l := New(Sources{Storages: rowsInReverse(s, have)})
			return check(t, l.StorageByFile, keys, missing, func(st database.FileStorage, k pgtype.UUID) bool { return st.FileID == k })
		}),
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &spy{}
			keys := ids(n)
			right, missErr := tc.run(s, keys, newID())

			if got := s.calls.Load(); got != 1 {
				t.Fatalf("expected 1 batched fetch for %d+1 loads, got %d", n, got)
			}
			for i, ok := range right {
				if !ok {
					t.Fatalf("caller %d received the wrong row (row order leaked into the result)", i)
				}
			}
			if !errors.Is(missErr, ErrNotFound) {
				t.Fatalf("expected ErrNotFound for a missing ID, got %v", missErr)
			}
		})
	}
}

// check loads every key plus one missing key concurrently.
func check[T any](t *testing.T, l *LookupLoader[T], keys []pgtype.UUID, missing pgtype.UUID, isRight func(T, pgtype.UUID) bool) ([]bool, error) {
	t.Helper()
	right := make([]bool, len(keys))
	var missErr error
	var wg sync.WaitGroup
	for i := range keys {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := l.Load(context.Background(), keys[i])
			if err != nil {
				t.Errorf("load %d: %v", i, err)
				return
			}
			right[i] = isRight(v, keys[i])
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, missErr = l.Load(context.Background(), missing)
	}()
	wg.Wait()
	return right, missErr
}

func TestLookupLoader_SameKeyRequestedManyTimesIsFetchedOnce(t *testing.T) {
	u := database.User{ID: newID(), Username: "repeat"}
	s := &spy{}
	l := New(Sources{Users: rowsInReverse(s, map[pgtype.UUID]database.User{u.ID: u})})

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

	if c := s.calls.Load(); c != 1 {
		t.Fatalf("expected 1 fetch, got %d", c)
	}
	if len(s.lastBatch) != 1 {
		t.Fatalf("expected the key to appear once in the batch, got %d entries", len(s.lastBatch))
	}
}

func TestLookupLoader_FetchErrorReachesCallers(t *testing.T) {
	boom := errors.New("db down")
	l := New(Sources{Users: func(context.Context, []pgtype.UUID) ([]database.User, error) { return nil, boom }})

	if _, err := l.UserByID.Load(context.Background(), newID()); !errors.Is(err, boom) {
		t.Fatalf("expected the fetch error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Operation scoping
// ---------------------------------------------------------------------------

func TestFrom_ReturnsNilWithoutMiddleware(t *testing.T) {
	if From(context.Background()) != nil {
		t.Fatal("expected nil loaders on a bare context")
	}
}

// Loaders are per operation: the same key loaded in two operations must hit the
// source twice (no cache leaking between operations on a long-lived socket).
func TestOperationMiddleware_GivesEachOperationFreshLoaders(t *testing.T) {
	s := &spy{}
	id := newID()
	src := Sources{Users: rowsInReverse(s, map[pgtype.UUID]database.User{id: {ID: id}})}

	for op := 0; op < 2; op++ {
		l := New(src)
		if _, err := l.UserByID.Load(context.Background(), id); err != nil {
			t.Fatalf("op %d: %v", op, err)
		}
	}
	if c := s.calls.Load(); c != 2 {
		t.Fatalf("expected one fetch per operation (2), got %d", c)
	}
}
