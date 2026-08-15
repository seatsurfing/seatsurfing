package measure

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/seatsurfing/seatsurfing/server/api"
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

// Actor is a fully-authenticated seeded user, ready to issue HTTP requests
// as itself.
type Actor struct {
	OrgIndex    int
	OrgID       string
	UserID      string
	JWT         string
	LocationIDs []string
	SpaceIDs    []string
	UsersPerOrg int
}

// LoadActors reads the actors file written by perftest-seed and mints a
// fresh, short-lived JWT for every actor user by talking to Postgres
// directly and reusing the server's own session/token-issuance code
// (server/router AuthRouter). This intentionally happens once, right
// before the timed HTTP measurement run starts, since access tokens are
// only valid for 15 minutes.
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
		claims := auth.CreateClaims(user, session)
		token := auth.CreateAccessToken(claims)
		if token == "" {
			return Actor{}, fmt.Errorf("failed to mint access token for user %s", userID)
		}
		return Actor{UserID: userID, OrgID: orgID, JWT: token}, nil
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
