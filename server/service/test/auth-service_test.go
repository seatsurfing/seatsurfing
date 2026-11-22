package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthServiceIsBoundToAuthProvider(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	p1 := &AuthProvider{OrganizationID: org.ID, Name: "P1", ProviderType: int(OAuth2)}
	GetAuthProviderRepository().Create(p1)
	p2 := &AuthProvider{OrganizationID: org.ID, Name: "P2", ProviderType: int(OAuth2)}
	GetAuthProviderRepository().Create(p2)
	s := service.GetAuthService()

	// Positive: bound to the provider
	CheckTestBool(t, true, s.IsBoundToAuthProvider(&User{AuthProviderID: NullUUID(p1.ID)}, p1))

	// Negative: unbound (empty and null UUID), other provider, nil values
	CheckTestBool(t, false, s.IsBoundToAuthProvider(&User{AuthProviderID: NullUUID("")}, p1))
	CheckTestBool(t, false, s.IsBoundToAuthProvider(&User{AuthProviderID: NullUUID(GetSettingsRepository().GetNullUUID())}, p1))
	CheckTestBool(t, false, s.IsBoundToAuthProvider(&User{AuthProviderID: NullUUID(p2.ID)}, p1))
	CheckTestBool(t, false, s.IsBoundToAuthProvider(nil, p1))
	CheckTestBool(t, false, s.IsBoundToAuthProvider(&User{AuthProviderID: NullUUID(p1.ID)}, nil))
}
