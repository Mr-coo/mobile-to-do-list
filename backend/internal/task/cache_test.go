package task_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"taskmanager/internal/task"
)

type countingRepo struct {
	*fakeRepo
	listCalls int
}

func (c *countingRepo) List(ctx context.Context, q task.ListQuery) ([]task.Task, int64, error) {
	c.listCalls++
	return c.fakeRepo.List(ctx, q)
}

func quietLogs(t *testing.T) {
	t.Helper()
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
}

func newCachedService(t *testing.T) (*task.Service, *countingRepo, *miniredis.Miniredis) {
	t.Helper()
	quietLogs(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr:        mr.Addr(),
		MaxRetries:  -1,
		DialTimeout: 100 * time.Millisecond,
	})
	t.Cleanup(func() { rdb.Close() })
	repo := &countingRepo{fakeRepo: newFakeRepo()}
	return task.NewService(repo, task.NewRedisListCache(rdb)), repo, mr
}

func parse(t *testing.T, raw string) task.ListQuery {
	t.Helper()
	v, _ := url.ParseQuery(raw)
	q, err := task.ParseListQuery(v)
	if err != nil {
		t.Fatal(err)
	}
	return q
}

func listKeys(mr *miniredis.Miniredis) []string {
	var out []string
	for _, k := range mr.Keys() {
		if strings.HasPrefix(k, "tasks:list:") {
			out = append(out, k)
		}
	}
	return out
}

func TestCacheKey(t *testing.T) {
	if got, want := parse(t, "").CacheKey(), "tasks:list:limit=10&page=1&sort=-created_at"; got != want {
		t.Errorf("defaults = %q, want %q", got, want)
	}
	got := parse(t, "status=done&keyword=milk").CacheKey()
	want := "tasks:list:keyword=milk&limit=10&page=1&sort=-created_at&status=done"
	if got != want {
		t.Errorf("filtered = %q, want %q", got, want)
	}
	if parse(t, "").CacheKey() != parse(t, "page=1&limit=10").CacheKey() {
		t.Error("default and explicit-default queries should share a key")
	}
}

func TestServiceList_MissThenHit(t *testing.T) {
	svc, repo, mr := newCachedService(t)
	ctx := context.Background()
	seed(t, svc, "A")
	q := parse(t, "")

	first, hit, err := svc.List(ctx, q)
	if err != nil || hit {
		t.Fatalf("first: hit=%v err=%v", hit, err)
	}
	second, hit, err := svc.List(ctx, q)
	if err != nil || !hit {
		t.Fatalf("second: hit=%v err=%v", hit, err)
	}
	if repo.listCalls != 1 {
		t.Errorf("repo List calls = %d, want 1", repo.listCalls)
	}
	if len(second.Data) != 1 || second.Meta != first.Meta {
		t.Errorf("cached result differs: %+v vs %+v", second, first)
	}
	if ttl := mr.TTL(q.CacheKey()); ttl != 60*time.Second {
		t.Errorf("ttl = %v, want 60s", ttl)
	}
}

func TestServiceList_DifferentParamsDifferentKeys(t *testing.T) {
	svc, _, mr := newCachedService(t)
	ctx := context.Background()
	svc.List(ctx, parse(t, ""))
	svc.List(ctx, parse(t, "status=done"))
	if n := len(mr.Keys()); n != 2 {
		t.Errorf("keys = %d, want 2", n)
	}
}

func warm(t *testing.T, svc *task.Service, mr *miniredis.Miniredis) {
	t.Helper()
	ctx := context.Background()
	svc.List(ctx, parse(t, ""))
	svc.List(ctx, parse(t, "status=done"))
	mr.Set("other:key", "x")
	if len(listKeys(mr)) != 2 {
		t.Fatalf("expected 2 warm list keys, got %v", mr.Keys())
	}
}

func assertInvalidated(t *testing.T, mr *miniredis.Miniredis) {
	t.Helper()
	if keys := listKeys(mr); len(keys) != 0 {
		t.Errorf("list keys left: %v", keys)
	}
	if !mr.Exists("other:key") {
		t.Error("non-list key was removed")
	}
}

func TestCreate_InvalidatesListCache(t *testing.T) {
	svc, _, mr := newCachedService(t)
	warm(t, svc, mr)
	seed(t, svc, "New")
	assertInvalidated(t, mr)
}

func TestUpdate_InvalidatesListCache(t *testing.T) {
	svc, _, mr := newCachedService(t)
	tk := seed(t, svc, "Old")
	warm(t, svc, mr)
	_, err := svc.Update(context.Background(), tk.ID, task.UpdateInput{Title: "New", Status: task.StatusDone})
	if err != nil {
		t.Fatal(err)
	}
	assertInvalidated(t, mr)
}

func TestDelete_InvalidatesListCache(t *testing.T) {
	svc, _, mr := newCachedService(t)
	tk := seed(t, svc, "Old")
	warm(t, svc, mr)
	if err := svc.Delete(context.Background(), tk.ID); err != nil {
		t.Fatal(err)
	}
	assertInvalidated(t, mr)
}

func TestServiceList_RedisDown(t *testing.T) {
	svc, _, mr := newCachedService(t)
	seed(t, svc, "A")
	mr.Close()

	res, hit, err := svc.List(context.Background(), parse(t, ""))
	if err != nil || hit {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
	if len(res.Data) != 1 {
		t.Errorf("data = %d, want 1", len(res.Data))
	}
	if _, err := svc.Create(context.Background(), task.CreateInput{Title: "B"}); err != nil {
		t.Errorf("create with redis down: %v", err)
	}
}

func TestList_XCacheHeader(t *testing.T) {
	svc, _, _ := newCachedService(t)
	r := newRouterWith(svc)
	if got := do(r, http.MethodGet, "/api/tasks", "").Header().Get("X-Cache"); got != "MISS" {
		t.Errorf("first X-Cache = %q, want MISS", got)
	}
	if got := do(r, http.MethodGet, "/api/tasks", "").Header().Get("X-Cache"); got != "HIT" {
		t.Errorf("second X-Cache = %q, want HIT", got)
	}
}
