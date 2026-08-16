package measure

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	"github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
)

// orgActorsFile mirrors seed.OrgActors / seed.ActorsFile. Duplicated here
// (rather than importing perftest/seed) so measure only depends on
// perftest/seed's JSON output format, not its implementation.
type orgActorsFile struct {
	OrgIndex     int      `json:"orgIndex"`
	OrgID        string   `json:"orgId"`
	OrgAdminID   string   `json:"orgAdminId"`
	SpaceAdminID string   `json:"spaceAdminId"`
	UsersPerOrg  int      `json:"usersPerOrg"`
	LocationIDs  []string `json:"locationIds"`
	SpaceIDs     []string `json:"spaceIds"`
}

type actorsFile struct {
	SuperAdminID    string          `json:"superAdminId"`
	SuperAdminOrgID string          `json:"superAdminOrgId"`
	Orgs            []orgActorsFile `json:"orgs"`
}

// accessTokenRefreshAfter is how long a minted access token is reused
// before Token() mints a replacement. AuthRouter.CreateAccessToken hardcodes
// a 15 minute expiry, so this leaves a comfortable margin for a request that
// is issued just before the cutoff and then runs into the server's write
// timeout.
const accessTokenRefreshAfter = 10 * time.Minute

// Actor is a fully-authenticated seeded user, ready to issue HTTP requests
// as itself. It holds on to its session so that a long measurement run can
// mint fresh access tokens without creating additional sessions (which would
// trip MAX_SESSIONS_PER_USER).
type Actor struct {
	OrgIndex    int
	OrgID       string
	UserID      string
	LocationIDs []string
	SpaceIDs    []string
	UsersPerOrg int

	auth     *AuthRouter
	user     *User
	session  *repository.Session
	jwt      string
	mintedAt time.Time
}

// Token returns a currently-valid access token for this actor, minting a
// fresh one when the previous token is old enough that it could expire
// mid-request. A full-scale run takes longer than the 15 minute token
// lifetime, so reusing the token minted at load time would make every
// request after that point fail with a bare 401.
func (a *Actor) Token() string {
	if a.jwt != "" && time.Since(a.mintedAt) < accessTokenRefreshAfter {
		return a.jwt
	}
	claims := a.auth.CreateClaims(a.user, a.session)
	a.jwt = a.auth.CreateAccessToken(claims)
	a.mintedAt = time.Now()
	return a.jwt
}

// LoadActors reads the actors file written by perftest-seed and creates one
// session plus an initial access token for every actor user, by talking to
// Postgres directly and reusing the server's own session/token-issuance code
// (server/router AuthRouter).
//
// Sessions are created once and last 90 days, but access tokens expire after
// 15 minutes -- far less than a full-scale run takes -- so each Actor keeps
// its session and re-mints its own token on demand via Actor.Token().
func LoadActors(path string) (orgAdmins, spaceAdmins []Actor, superAdmin *Actor, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	var f actorsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, nil, nil, err
	}

	auth := &AuthRouter{}
	mint := func(userID, orgID string, role UserRole) (Actor, error) {
		user := &User{ID: userID, OrganizationID: orgID, Role: role}
		session := auth.CreateSession(nil, user)
		if session == nil || session.ID == "" {
			return Actor{}, fmt.Errorf("failed to create session for user %s", userID)
		}
		a := Actor{UserID: userID, OrgID: orgID, auth: auth, user: user, session: session}
		if a.Token() == "" {
			return Actor{}, fmt.Errorf("failed to mint access token for user %s", userID)
		}
		return a, nil
	}

	for _, o := range f.Orgs {
		if o.OrgAdminID != "" {
			role := UserRoleOrgAdmin
			if o.OrgID == f.SuperAdminOrgID && o.OrgAdminID == f.SuperAdminID {
				role = UserRoleSuperAdmin
			}
			a, err := mint(o.OrgAdminID, o.OrgID, role)
			if err != nil {
				return nil, nil, nil, err
			}
			a.OrgIndex = o.OrgIndex
			a.LocationIDs = o.LocationIDs
			a.SpaceIDs = o.SpaceIDs
			a.UsersPerOrg = o.UsersPerOrg
			orgAdmins = append(orgAdmins, a)
			if role == UserRoleSuperAdmin {
				sa := a
				superAdmin = &sa
			}
		}
		if o.SpaceAdminID != "" {
			a, err := mint(o.SpaceAdminID, o.OrgID, UserRoleSpaceAdmin)
			if err != nil {
				return nil, nil, nil, err
			}
			a.OrgIndex = o.OrgIndex
			a.LocationIDs = o.LocationIDs
			a.SpaceIDs = o.SpaceIDs
			a.UsersPerOrg = o.UsersPerOrg
			spaceAdmins = append(spaceAdmins, a)
		}
	}

	if superAdmin == nil {
		return nil, nil, nil, fmt.Errorf("actors file has no super admin (superAdminId=%q)", f.SuperAdminID)
	}
	return orgAdmins, spaceAdmins, superAdmin, nil
}
