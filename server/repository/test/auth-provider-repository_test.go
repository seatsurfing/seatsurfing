package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestCallbackURLDomain(t *testing.T) {
	ClearTestDB()
	orgDomain := "test1.com"
	org := CreateTestOrg(orgDomain)
	authProvider := &AuthProvider{
		OrganizationID: org.ID,
		Name:           "test",
		ProviderType:   int(OAuth2),
	}
	GetAuthProviderRepository().Create(authProvider)

	authProviders, _ := GetAuthProviderRepository().GetAll(org.ID)
	CheckTestString(t, orgDomain, authProviders[0].CallbackURLDomain)
}
