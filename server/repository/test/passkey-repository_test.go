package test

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func makeTestCredential(rawID []byte) *webauthn.Credential {
	return &webauthn.Credential{
		ID:              rawID,
		PublicKey:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		AttestationType: "none",
		Flags: webauthn.CredentialFlags{
			BackupEligible: true,
			BackupState:    false,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    make([]byte, 16),
			SignCount: 42,
		},
	}
}

func createTestPasskey(user *User, name string, rawID []byte) *Passkey {
	cred := makeTestCredential(rawID)
	pk, err := NewPasskeyFromCredential(user.ID, cred, name)
	if err != nil {
		panic("NewPasskeyFromCredential: " + err.Error())
	}
	if err := GetPasskeyRepository().Create(pk); err != nil {
		panic("Create passkey: " + err.Error())
	}
	return pk
}

func TestPasskeyRepositoryCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte(uuid.New().String())
	pk := createTestPasskey(user, "My YubiKey", rawID)

	CheckStringNotEmpty(t, pk.ID)
	CheckStringNotEmpty(t, pk.CredentialIDEncrypted)
	CheckStringNotEmpty(t, pk.CredentialIDHash)

	got, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, pk.ID, got.ID)
	CheckTestString(t, user.ID, got.UserID)
	CheckTestString(t, "My YubiKey", got.Name)
	CheckTestBool(t, true, got.BackupEligible)
	CheckTestBool(t, false, got.BackupState)
	CheckTestUint(t, 42, uint(got.SignCount))

	all, err := GetPasskeyRepository().GetAllByUserID(user.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(all))
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user.ID))

	err = GetPasskeyRepository().Delete(pk)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, GetPasskeyRepository().GetCountByUserID(user.ID))

	_, err = GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err != nil)
}

func TestPasskeyRepositoryMultiplePasskeys(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	createTestPasskey(user, "Key 1", []byte(uuid.New().String()))
	createTestPasskey(user, "Key 2", []byte(uuid.New().String()))
	createTestPasskey(user, "Key 3", []byte(uuid.New().String()))

	CheckTestInt(t, 3, GetPasskeyRepository().GetCountByUserID(user.ID))

	all, err := GetPasskeyRepository().GetAllByUserID(user.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, len(all))
}

func TestPasskeyRepositoryGetByCredentialIDRaw(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte("unique-credential-id-for-hash-test")
	pk := createTestPasskey(user, "Hash Test Key", rawID)

	found, err := GetPasskeyRepository().GetByCredentialIDRaw(rawID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, pk.ID, found.ID)
	CheckTestString(t, user.ID, found.UserID)
}

func TestPasskeyRepositoryGetByCredentialIDHash(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte("another-unique-raw-id-for-hash")
	pk := createTestPasskey(user, "Direct Hash Key", rawID)

	found, err := GetPasskeyRepository().GetByCredentialIDHash(pk.CredentialIDHash)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, pk.ID, found.ID)
}

func TestPasskeyRepositoryUpdateSignCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte(uuid.New().String())
	pk := createTestPasskey(user, "Sign Count Key", rawID)
	CheckTestUint(t, 42, uint(pk.SignCount))

	pk.SignCount = 100
	err := GetPasskeyRepository().UpdateSignCount(pk)
	CheckTestBool(t, true, err == nil)

	got, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestUint(t, 100, uint(got.SignCount))
}

func TestPasskeyRepositoryUpdateLastUsedAt(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte(uuid.New().String())
	pk := createTestPasskey(user, "Last Used Key", rawID)

	got, _ := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestIsNil(t, got.LastUsedAt)

	err := GetPasskeyRepository().UpdateLastUsedAt(pk)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, pk.LastUsedAt != nil)

	got, err = GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, got.LastUsedAt != nil)
	CheckTestBool(t, true, time.Since(*got.LastUsedAt) < 5*time.Second)
}

func TestPasskeyRepositoryUpdateName(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte(uuid.New().String())
	pk := createTestPasskey(user, "Old Name", rawID)

	pk.Name = "New Name"
	err := GetPasskeyRepository().UpdateName(pk)
	CheckTestBool(t, true, err == nil)

	got, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "New Name", got.Name)
}

func TestPasskeyRepositoryDeleteAllByUserID(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	createTestPasskey(user1, "U1 Key 1", []byte(uuid.New().String()))
	createTestPasskey(user1, "U1 Key 2", []byte(uuid.New().String()))
	createTestPasskey(user2, "U2 Key 1", []byte(uuid.New().String()))

	CheckTestInt(t, 2, GetPasskeyRepository().GetCountByUserID(user1.ID))
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user2.ID))

	err := GetPasskeyRepository().DeleteAllByUserID(user1.ID)
	CheckTestBool(t, true, err == nil)

	CheckTestInt(t, 0, GetPasskeyRepository().GetCountByUserID(user1.ID))
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user2.ID))
}

func TestPasskeyRepositoryCascadeDeleteOnUserDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte(uuid.New().String())
	pk := createTestPasskey(user, "Cascade Key", rawID)
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user.ID))

	err := GetUserRepository().Delete(user)
	CheckTestBool(t, true, err == nil)

	_, err = GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err != nil)
	CheckTestInt(t, 0, GetPasskeyRepository().GetCountByUserID(user.ID))
}

func TestPasskeyRepositoryToWebAuthnCredentialRoundtrip(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte("roundtrip-test-credential-id-12345")
	cred := makeTestCredential(rawID)

	pk, err := NewPasskeyFromCredential(user.ID, cred, "Roundtrip Key")
	CheckTestBool(t, true, err == nil)
	if err := GetPasskeyRepository().Create(pk); err != nil {
		t.Fatal(err)
	}

	got, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)

	converted, err := got.ToWebAuthnCredential()
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, bytes.Equal(rawID, converted.ID))
	CheckTestBool(t, true, bytes.Equal(cred.PublicKey, converted.PublicKey))
	CheckTestBool(t, true, bytes.Equal(cred.Authenticator.AAGUID, converted.Authenticator.AAGUID))
	CheckTestUint(t, 42, uint(converted.Authenticator.SignCount))
	CheckTestBool(t, true, converted.Flags.BackupEligible)
	CheckTestBool(t, false, converted.Flags.BackupState)
}

func TestPasskeyRepositoryEncryptionAtRest(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	rawID := []byte("sensitive-credential-id-bytes")
	pk := createTestPasskey(user, "Encryption Test Key", rawID)

	rawIDBase64 := "c2Vuc2l0aXZlLWNyZWRlbnRpYWwtaWQtYnl0ZXM="
	CheckTestBool(t, false, pk.CredentialIDEncrypted == rawIDBase64)
	CheckTestBool(t, false, pk.CredentialIDHash == "")

	got, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	converted, err := got.ToWebAuthnCredential()
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, bytes.Equal(rawID, converted.ID))
}

func TestPasskeyRepositoryIsolationBetweenUsers(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	createTestPasskey(user1, "User1 Key", []byte(uuid.New().String()))

	CheckTestInt(t, 0, GetPasskeyRepository().GetCountByUserID(user2.ID))
	all, err := GetPasskeyRepository().GetAllByUserID(user2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(all))
}
