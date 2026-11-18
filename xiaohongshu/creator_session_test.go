package xiaohongshu

import (
	"testing"
	"time"
)

func TestIsCreatorURL(t *testing.T) {
	for _, tc := range []struct {
		url  string
		want bool
	}{
		{"https://creator.rednote.com/new/home", true},
		{"https://creator.rednote.com/publish/publish?from=menu", true},
		{"https://creator.rednote.com/login?lastUrl=home", false},
		{"https://creator.xiaohongshu.com/new/home", false},
		{"https://creator.rednote.com.evil.test/new/home", false},
		{"https://evil.test/?next=https://creator.rednote.com/new/home", false},
		{"http://creator.rednote.com/new/home", false},
	} {
		if got := isCreatorURL(tc.url); got != tc.want {
			t.Errorf("isCreatorURL(%q)=%v want %v", tc.url, got, tc.want)
		}
	}
}

func TestCreatorScheduleTime(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"2026-11-19T10:30:00-05:00", "2026-11-19 23:30"},
		{"2026-07-19T10:30:00-04:00", "2026-07-19 22:30"},
		{"2026-11-19T23:30:00+08:00", "2026-11-19 23:30"},
	} {
		value, err := time.Parse(time.RFC3339, tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := creatorScheduleTime(value); got != tc.want {
			t.Errorf("got %s want %s", got, tc.want)
		}
	}
}

func TestPublishSuccessURL(t *testing.T) {
	for _, tc := range []struct {
		url  string
		want bool
	}{
		{"https://creator.rednote.com/publish/success?noteId=example", true},
		{"https://creator.rednote.com/new/note-manager", true},
		{"https://creator.rednote.com/login", false},
		{"https://creator.rednote.com/publish/publish", false},
		{"https://creator.rednote.com/new/home", false},
		{"https://creator.xiaohongshu.com/publish/success", false},
		{"about:blank", false},
	} {
		if got := isPublishSuccessURL(tc.url); got != tc.want {
			t.Errorf("%s got %v want %v", tc.url, got, tc.want)
		}
	}
}
