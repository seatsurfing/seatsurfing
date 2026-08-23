package measure

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// reservedActorIndices is the number of leading user indices per org that
// seed.Run reserves for actor users (org admin, space admin) -- see
// seed.Run's userRoles assignment.
const reservedActorIndices = 2

// Scenario is one measured REST API call shape. Request builds a request
// for a given org-admin/space-admin actor; scenarios that need
// cross-org/global access (e.g. listing all organizations) use the shared
// super admin actor instead and ignore the passed-in actor's org.
type Scenario struct {
	// ID identifies this scenario; it is the key looked up in the
	// thresholds YAML file.
	ID      string
	Request func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error)
}

func newGet(baseURL, path, jwt string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}
	return req, nil
}

func randomWeekWindow(rnd *rand.Rand) (start, end time.Time) {
	now := time.Now().UTC()
	// A random week somewhere in the seeded +/- 180 day booking window.
	offsetDays := rnd.Intn(2*180+1) - 180
	start = now.AddDate(0, 0, offsetDays).Truncate(24 * time.Hour)
	end = start.AddDate(0, 0, 7)
	return
}

// randomDayInBookingWindow returns a day within the +/- 180 day range that
// seed.Run spreads booking enter times across (see seed.bookingWindowDays),
// so date-scoped scenarios hit a representative amount of seeded data.
func randomDayInBookingWindow(rnd *rand.Rand) time.Time {
	now := time.Now().UTC()
	offsetDays := rnd.Intn(2*180+1) - 180
	return now.AddDate(0, 0, offsetDays).Truncate(24 * time.Hour)
}

// Scenarios is the full set of REST endpoints exercised by the
// measurement run. Endpoint paths and auth requirements were verified
// against server/router/*.go.
var Scenarios = []Scenario{
	{
		ID: "organization.getOne",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/organization/"+a.OrgID, a.Token())
		},
	},
	{
		ID: "user.list",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/user/", a.Token())
		},
	},
	{
		ID: "user.count",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/user/count", a.Token())
		},
	},
	{
		ID: "user.byEmail",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			if a.UsersPerOrg == 0 {
				return newGet(baseURL, "/user/byEmail/nonexistent@perf.test", a.Token())
			}
			// Indices 0/1 are the org-admin/space-admin actor users
			// (seed.Run); skip them so the query never targets the
			// requesting actor's own email, which the endpoint always
			// 404s on.
			idx := reservedActorIndices
			if a.UsersPerOrg > reservedActorIndices {
				idx += rnd.Intn(a.UsersPerOrg - reservedActorIndices)
			} else {
				idx = 0
			}
			email := fmt.Sprintf("user%06d@org%04d.perf.test", idx, a.OrgIndex)
			return newGet(baseURL, "/user/byEmail/"+url.PathEscape(email), a.Token())
		},
	},
	{
		ID: "space.list",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			if len(a.LocationIDs) == 0 {
				return nil, fmt.Errorf("actor for org %d has no seeded locations", a.OrgIndex)
			}
			locID := a.LocationIDs[rnd.Intn(len(a.LocationIDs))]
			return newGet(baseURL, "/location/"+locID+"/space/", a.Token())
		},
	},
	{
		// The UI's main booking-search page (ui/src/pages/search.tsx) calls
		// this on every location/date change. It joins spaces to
		// bookings-in-range, space attribute values, group memberships and
		// allowed-booker lists, making it one of the most expensive
		// endpoints at scale (see space-router.go _getAvailability).
		ID: "space.availability",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			if len(a.LocationIDs) == 0 {
				return nil, fmt.Errorf("actor for org %d has no seeded locations", a.OrgIndex)
			}
			locID := a.LocationIDs[rnd.Intn(len(a.LocationIDs))]
			day := randomDayInBookingWindow(rnd)
			enter := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.UTC)
			leave := time.Date(day.Year(), day.Month(), day.Day(), 17, 0, 0, 0, time.UTC)
			q := url.Values{}
			q.Set("enter", enter.Format(time.RFC3339Nano))
			q.Set("leave", leave.Format(time.RFC3339Nano))
			return newGet(baseURL, "/location/"+locID+"/space/availability?"+q.Encode(), a.Token())
		},
	},
	{
		ID: "booking.filter",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			start, end := randomWeekWindow(rnd)
			q := url.Values{}
			q.Set("start", start.Format(time.RFC3339Nano))
			q.Set("end", end.Format(time.RFC3339Nano))
			return newGet(baseURL, "/booking/filter/?"+q.Encode(), a.Token())
		},
	},
	{
		ID: "booking.current",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/booking/current/", a.Token())
		},
	},
	{
		ID: "booking.pendingApprovals",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/booking/pendingapprovals/", a.Token())
		},
	},
	{
		ID: "stats.summary",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/stats/", a.Token())
		},
	},
	{
		ID: "stats.load",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/stats/load", a.Token())
		},
	},
	{
		ID: "stats.weekday",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/stats/weekday", a.Token())
		},
	},
	{
		// The admin bookings page (ui/src/pages/admin/bookings/index.tsx)
		// uses this for its user-search typeahead, via
		// UserStore.GetByKeyword (a LIKE query over the org's users). "user"
		// matches the email prefix of every seeded regular user, so this
		// exercises the worst-case (near-full-table) match set.
		ID: "search.users",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/search/?includeUsers=1&query=user", a.Token())
		},
	},
	{
		ID: "user.me",
		Request: func(baseURL string, a *Actor, rnd *rand.Rand) (*http.Request, error) {
			return newGet(baseURL, "/user/me", a.Token())
		},
	},
}
