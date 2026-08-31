package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// baselineRoutes are the endpoints a user with no roles at all is *supposed*
// to reach: their own bookings, buddies, preferences and profile, plus the
// areas and spaces they can book. Everything else must refuse them.
//
// This list is the point of the test. A new endpoint is refused by default, so
// forgetting a permission check makes the test fail; adding an entry here is a
// deliberate statement that the endpoint needs no permission. Keep it short,
// and justify additions.
var baselineRoutes = map[string]string{
	// The user's own identity and session.
	"GET /user/me":             "own profile",
	"GET /user/session":        "own active sessions",
	"GET /user/merge":          "own merge requests",
	"POST /user/merge/init":    "own account merge",
	"GET /user/passkey/":       "own passkeys",
	"POST /user/totp/generate": "own second factor",
	"GET /user/totp/generate":  "own second factor",

	// Own bookings and the data needed to make one.
	"GET /booking/":                     "own bookings",
	"POST /booking/":                    "book for yourself",
	"GET /booking/precheck/":            "own booking pre-check",
	"GET /recurring-booking/":           "own recurring bookings",
	"POST /recurring-booking/":          "book for yourself",
	"POST /recurring-booking/precheck/": "own booking pre-check",

	"POST /user/totp/disable": "own second factor",

	// Buddies and preferences are entirely self-scoped.
	"GET /preference/":  "own preferences",
	"PUT /preference/":  "own preferences",
	"GET /buddy/":       "own buddies",
	"PUT /buddy/":       "own buddies",
	"GET /buddy/search": "finding colleagues to add",

	// Areas and spaces: without read access the booking UI cannot function.
	"GET /location/":        "areas available to book",
	"GET /location/{id}":    "areas available to book",
	"GET /space-attribute/": "attributes shown while booking",

	// Settings the booking UI needs, filtered server-side to public ones.
	"GET /setting/":       "public settings",
	"GET /setting/{name}": "public settings",

	// Version check and timezone list carry no organization data.
	"GET /uc/":               "update check",
	"GET /setting/timezones": "timezone list",

	// The permission catalogue describes the model, not the caller's access.
	"GET /role/permissions": "permission catalogue",

	// CORS preflight, answered by the security header middleware.
	"OPTIONS /": "CORS preflight",
}

// substitute fills a route template with values that parse, so the request
// reaches the handler's authorization check rather than failing earlier.
func substitute(template string) string {
	r := strings.NewReplacer(
		"{id}", "00000000-0000-0000-0000-000000000000",
		"{locationId}", "00000000-0000-0000-0000-000000000000",
		"{attributeId}", "00000000-0000-0000-0000-000000000000",
		"{uuid}", "00000000-0000-0000-0000-000000000000",
		"{stateId}", "00000000-0000-0000-0000-000000000000",
		"{name}", "hide_reports",
		"{domain}", "example.com",
		"{email}", "nobody@test.com",
		"{keyword}", "x",
	)
	return r.Replace(template)
}

func isUnauthorizedPrefix(path string) bool {
	for _, p := range []string{
		"/auth/", "/organization/domain/", "/organization/deleteorg/",
		"/auth-provider/org/", "/admin/", "/ui/", "/confluence",
		"/robots.txt", "/healthcheck", "/kiosk/", "/uc/",
	} {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// TestEveryEndpointRefusesAUserWithoutPermissions walks the live route table
// and calls every endpoint as a user holding no roles. Anything that answers
// 2xx without being declared baseline is an endpoint missing its permission
// check.
//
// Walking the router rather than listing endpoints by hand is what makes this
// hold for endpoints added later: they are covered the moment they are
// registered.
func TestEveryEndpointRefusesAUserWithoutPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org) // the organization keeps an administrator
	plain := CreateTestUserInOrg(org)
	login := LoginTestUser(plain.ID)

	type route struct{ method, template string }
	var routes []route
	err := GetApp().Router.Walk(func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		tpl, err := r.GetPathTemplate()
		if err != nil {
			return nil
		}
		methods, err := r.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}
		for _, m := range methods {
			routes = append(routes, route{m, tpl})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) < 50 {
		t.Fatalf("expected the walk to find the API surface, got %d routes", len(routes))
	}

	var leaked []string
	checked := 0
	for _, rt := range routes {
		if isUnauthorizedPrefix(rt.template) {
			continue
		}
		key := rt.method + " " + rt.template
		if _, baseline := baselineRoutes[key]; baseline {
			continue
		}
		checked++
		var body *bytes.Buffer = bytes.NewBufferString("{}")
		req := NewHTTPRequest(rt.method, substitute(rt.template), login.UserID, body)
		res := ExecuteTestRequest(req)
		// Any refusal is fine - 403 for the permission, 400 for the empty
		// body, 404 for the placeholder ID. A 2xx means the endpoint let a
		// user with no roles through.
		if res.Code >= 200 && res.Code < 300 {
			leaked = append(leaked, key)
		}
	}
	if checked < 40 {
		t.Fatalf("expected to exercise the guarded surface, only checked %d routes", checked)
	}
	if len(leaked) > 0 {
		t.Fatalf("%d endpoint(s) answered success for a user with no permissions.\n"+
			"Either add the missing permission check, or declare it in baselineRoutes with a reason:\n  %s",
			len(leaked), strings.Join(leaked, "\n  "))
	}
}
