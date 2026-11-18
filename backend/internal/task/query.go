package task

import (
	"net/url"
	"strconv"
	"strings"

	"taskmanager/internal/apperr"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
	defaultSort  = "-created_at"
)

var sortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"title":      "title",
	"status":     "status",
}

type ListQuery struct {
	Status   Status
	Keyword  string
	Assignee string
	Page     int
	Limit    int
	Sort     string
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListResult struct {
	Data []Task `json:"data"`
	Meta Meta   `json:"meta"`
}

// ParseListQuery applies defaults and reports every invalid parameter at once.
func ParseListQuery(v url.Values) (ListQuery, error) {
	q := ListQuery{
		Status:   Status(strings.TrimSpace(v.Get("status"))),
		Keyword:  strings.TrimSpace(v.Get("keyword")),
		Assignee: strings.TrimSpace(v.Get("assignee")),
		Page:     defaultPage,
		Limit:    defaultLimit,
		Sort:     defaultSort,
	}
	details := map[string]string{}

	if q.Status != "" && !q.Status.Valid() {
		details["status"] = "must be one of todo, in_progress, done"
	}
	if s := strings.TrimSpace(v.Get("page")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			details["page"] = "must be a positive integer"
		} else {
			q.Page = n
		}
	}
	if s := strings.TrimSpace(v.Get("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > maxLimit {
			details["limit"] = "must be between 1 and 100"
		} else {
			q.Limit = n
		}
	}
	if s := strings.TrimSpace(v.Get("sort")); s != "" {
		if _, ok := sortFields[strings.TrimPrefix(s, "-")]; !ok {
			details["sort"] = "must be one of created_at, updated_at, title, status (prefix - for descending)"
		} else {
			q.Sort = s
		}
	}

	if len(details) > 0 {
		return ListQuery{}, &apperr.ValidationError{Details: details}
	}
	return q, nil
}

// CacheKey is built from the normalised query, so "/tasks" and
// "/tasks?page=1&limit=10" share one entry. Encode sorts the keys.
func (q ListQuery) CacheKey() string {
	v := url.Values{}
	v.Set("page", strconv.Itoa(q.Page))
	v.Set("limit", strconv.Itoa(q.Limit))
	v.Set("sort", q.Sort)
	if q.Status != "" {
		v.Set("status", string(q.Status))
	}
	if q.Keyword != "" {
		v.Set("keyword", q.Keyword)
	}
	if q.Assignee != "" {
		v.Set("assignee", q.Assignee)
	}
	return listKeyPrefix + v.Encode()
}
