package task_test

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"taskmanager/internal/apperr"
	"taskmanager/internal/task"
)

func TestParseListQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		want        task.ListQuery
		wantInvalid string
	}{
		{name: "defaults", query: "", want: task.ListQuery{Page: 1, Limit: 10, Sort: "-created_at"}},
		{name: "invalid status", query: "status=bogus", wantInvalid: "status"},
		{name: "limit too big", query: "limit=101", wantInvalid: "limit"},
		{name: "limit zero", query: "limit=0", wantInvalid: "limit"},
		{name: "page not a number", query: "page=abc", wantInvalid: "page"},
		{name: "page zero", query: "page=0", wantInvalid: "page"},
		{name: "sort not whitelisted", query: "sort=password", wantInvalid: "sort"},
		{name: "sort title", query: "sort=title", want: task.ListQuery{Page: 1, Limit: 10, Sort: "title"}},
		{name: "sort descending", query: "sort=-updated_at", want: task.ListQuery{Page: 1, Limit: 10, Sort: "-updated_at"}},
		{name: "keyword too long", query: "keyword=" + strings.Repeat("a", 101), wantInvalid: "keyword"},
		{name: "keyword at limit", query: "keyword=" + strings.Repeat("é", 100), want: task.ListQuery{Keyword: strings.Repeat("é", 100), Page: 1, Limit: 10, Sort: "-created_at"}},
		{name: "assignee too long", query: "assignee=" + strings.Repeat("a", 101), wantInvalid: "assignee"},
		{name: "keyword trimmed", query: "keyword=%20%20milk%20", want: task.ListQuery{Keyword: "milk", Page: 1, Limit: 10, Sort: "-created_at"}},
		{name: "all params", query: "status=done&assignee=%20Ana&page=2&limit=5", want: task.ListQuery{Status: task.StatusDone, Assignee: "Ana", Page: 2, Limit: 5, Sort: "-created_at"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := url.ParseQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			got, err := task.ParseListQuery(v)
			if tt.wantInvalid != "" {
				var verr *apperr.ValidationError
				if !errors.As(err, &verr) {
					t.Fatalf("err = %v, want ValidationError", err)
				}
				if _, ok := verr.Details[tt.wantInvalid]; !ok || len(verr.Details) != 1 {
					t.Fatalf("details = %v, want only %q", verr.Details, tt.wantInvalid)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseListQuery_Messages(t *testing.T) {
	v, _ := url.ParseQuery("status=x&page=x&limit=x&sort=x&keyword=" + strings.Repeat("a", 101) + "&assignee=" + strings.Repeat("a", 101))
	_, err := task.ParseListQuery(v)
	var verr *apperr.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("err = %v", err)
	}
	want := map[string]string{
		"status":   "must be one of todo, in_progress, done",
		"page":     "must be a positive integer",
		"limit":    "must be between 1 and 100",
		"sort":     "must be one of created_at, updated_at, title, status (prefix - for descending)",
		"keyword":  "must be at most 100 characters",
		"assignee": "must be at most 100 characters",
	}
	if !reflect.DeepEqual(verr.Details, want) {
		t.Fatalf("details = %v", verr.Details)
	}
}
