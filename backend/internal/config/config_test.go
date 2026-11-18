package config

import (
	"reflect"
	"testing"
)

func TestLoad_TrimsCORSOrigins(t *testing.T) {
	t.Setenv("CORS_ORIGINS", " http://a.com , ,http://b.com,")
	got := Load().CORSOrigins
	want := []string{"http://a.com", "http://b.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
