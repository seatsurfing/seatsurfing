package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/seatsurfing/seatsurfing/server/config"
)

// Simulates a Next.js static export tree that has a page nested below a
// dynamic segment (e.g. /admin/users/[id]/calendar), and verifies that
// setupStaticUIRoutes serves it for an arbitrary id, not just the literal
// "[id]" folder name. This guards against a regression where the
// bracket->mux-var conversion only worked when the dynamic segment was the
// last path component.
func TestSetupStaticUIRoutesNestedDynamicSegment(t *testing.T) {
	if err := os.Setenv("CRYPT_KEY", "12345678901234567890123456789012"); err != nil {
		t.Fatal(err)
	}
	staticDir := t.TempDir()

	writeFile(t, filepath.Join(staticDir, "admin", "users", "[id]", "index.html"), "user-page")
	writeFile(t, filepath.Join(staticDir, "admin", "users", "[id]", "calendar", "index.html"), "calendar-page")
	writeFile(t, filepath.Join(staticDir, "_attr.json"),
		`["/admin/users/[id]","/admin/users/[id]/calendar"]`)

	cfg := GetConfig()
	cfg.StaticUiPath = staticDir + "/"
	cfg.Development = false

	a := &App{}
	router := mux.NewRouter()
	a.setupStaticUIRoutes(router)

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"dynamic leaf page", "/ui/admin/users/some-real-uuid/", http.StatusOK, "user-page"},
		{"static page nested below dynamic segment", "/ui/admin/users/some-real-uuid/calendar/", http.StatusOK, "calendar-page"},
		{"different id resolves the same route", "/ui/admin/users/another-uuid/calendar/", http.StatusOK, "calendar-page"},
		{"unrelated nested path under a real id is not served", "/ui/admin/users/some-real-uuid/does-not-exist/", http.StatusNotFound, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("GET %s: expected status %d, got %d", tc.path, tc.wantStatus, rec.Code)
			}
			if tc.wantStatus == http.StatusOK && rec.Body.String() != tc.wantBody {
				t.Fatalf("GET %s: expected body %q, got %q", tc.path, tc.wantBody, rec.Body.String())
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
