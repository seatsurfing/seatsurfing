package repository

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lib/pq"

	. "github.com/seatsurfing/seatsurfing/server/util"
)

type PasskeyRepository struct {
}

type Passkey struct {
	ID                    string
	UserID                string
	CredentialIDEncrypted string
	CredentialIDHash      string
	PublicKeyEncrypted    string
	AttestationType       string
	AAGUID                []byte
	SignCount             uint32
	BackupEligible        bool
	BackupState           bool
	Transports            []string
	Name                  string
	CreatedAt             time.Time
	LastUsedAt            *time.Time
}

var passkeyRepository *PasskeyRepository
var passkeyRepositoryOnce sync.Once

func GetPasskeyRepository() *PasskeyRepository {
	passkeyRepositoryOnce.Do(func() {
		passkeyRepository = &PasskeyRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS passkeys (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, " +
			"credential_id_encrypted VARCHAR NOT NULL, " +
			"credential_id_hash VARCHAR NOT NULL, " +
			"public_key_encrypted VARCHAR NOT NULL, " +
			"attestation_type VARCHAR NOT NULL DEFAULT '', " +
			"aaguid VARCHAR NOT NULL DEFAULT '', " +
			"sign_count BIGINT NOT NULL DEFAULT 0, " +
			"backup_eligible BOOL NOT NULL DEFAULT false, " +
			"backup_state BOOL NOT NULL DEFAULT false, " +
			"transports VARCHAR[] NOT NULL DEFAULT '{}', " +
			"name VARCHAR(255) NOT NULL DEFAULT '', " +
			"created_at TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"last_used_at TIMESTAMP NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		if _, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_passkeys_user_id ON passkeys(user_id)"); err != nil {
			panic(err)
		}
		if _, err = GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_passkeys_credential_id_hash ON passkeys(credential_id_hash)"); err != nil {
			panic(err)
		}
	})
	return passkeyRepository
}

func (r *PasskeyRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// no schema changes yet
}

// hashBytes returns a hex-encoded SHA-256 hash of the given byte slice.
// Used for credential ID lookups (since encrypted values are non-deterministic).
func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// encryptBytes base64-encodes the byte slice, then encrypts the result.
func encryptBytes(data []byte) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(data)
	return EncryptString(b64)
}

// decryptBytes decrypts the stored value and base64-decodes it back to bytes.
func decryptBytes(encrypted string) ([]byte, error) {
	b64, err := DecryptString(encrypted)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(b64)
}

func (r *PasskeyRepository) Create(e *Passkey) error {
	var id string
	aaguidEncrypted, err := encryptBytes(e.AAGUID)
	if err != nil {
		return err
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	err = GetDatabase().DB().QueryRow("INSERT INTO passkeys "+
		"(user_id, credential_id_encrypted, credential_id_hash, public_key_encrypted, attestation_type, aaguid, sign_count, backup_eligible, backup_state, transports, name, created_at) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) "+
		"RETURNING id",
		e.UserID,
		e.CredentialIDEncrypted,
		e.CredentialIDHash,
		e.PublicKeyEncrypted,
		e.AttestationType,
		aaguidEncrypted,
		e.SignCount,
		e.BackupEligible,
		e.BackupState,
		pq.Array(e.Transports),
		e.Name,
		e.CreatedAt,
	).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *PasskeyRepository) scanRow(row interface{ Scan(...interface{}) error }) (*Passkey, error) {
	e := &Passkey{}
	var aaguidEncrypted string
	var transports pq.StringArray
	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.CredentialIDEncrypted,
		&e.CredentialIDHash,
		&e.PublicKeyEncrypted,
		&e.AttestationType,
		&aaguidEncrypted,
		&e.SignCount,
		&e.BackupEligible,
		&e.BackupState,
		&transports,
		&e.Name,
		&e.CreatedAt,
		&e.LastUsedAt,
	)
	if err != nil {
		return nil, err
	}
	e.Transports = []string(transports)
	// Decrypt AAGUID (try encrypted format first, fall back to legacy hex)
	aaguidBytes, err := decryptBytes(aaguidEncrypted)
	if err != nil {
		// Fallback: legacy unencrypted hex format
		aaguidBytes, err = hex.DecodeString(aaguidEncrypted)
		if err != nil {
			aaguidBytes = nil
		}
	}
	e.AAGUID = aaguidBytes
	return e, nil
}

const passkeySelectCols = "id, user_id, credential_id_encrypted, credential_id_hash, public_key_encrypted, attestation_type, aaguid, sign_count, backup_eligible, backup_state, transports, name, created_at, last_used_at"

func (r *PasskeyRepository) GetOne(id string) (*Passkey, error) {
	row := GetDatabase().DB().QueryRow("SELECT "+passkeySelectCols+" FROM passkeys WHERE id = $1", id)
	return r.scanRow(row)
}

func (r *PasskeyRepository) GetByCredentialIDHash(hash string) (*Passkey, error) {
	row := GetDatabase().DB().QueryRow("SELECT "+passkeySelectCols+" FROM passkeys WHERE credential_id_hash = $1", hash)
	return r.scanRow(row)
}

// GetByCredentialIDRaw looks up a passkey by the raw (plaintext) credential ID bytes.
// It computes the SHA-256 hash internally for the lookup.
func (r *PasskeyRepository) GetByCredentialIDRaw(rawID []byte) (*Passkey, error) {
	return r.GetByCredentialIDHash(hashBytes(rawID))
}

func (r *PasskeyRepository) GetAllByUserID(userID string) ([]*Passkey, error) {
	rows, err := GetDatabase().DB().Query("SELECT "+passkeySelectCols+" FROM passkeys WHERE user_id = $1 ORDER BY created_at ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*Passkey
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *PasskeyRepository) GetCountByUserID(userID string) int {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM passkeys WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (r *PasskeyRepository) UpdateSignCount(e *Passkey) error {
	_, err := GetDatabase().DB().Exec("UPDATE passkeys SET sign_count = $1 WHERE id = $2", e.SignCount, e.ID)
	return err
}

func (r *PasskeyRepository) UpdateLastUsedAt(e *Passkey) error {
	now := time.Now()
	e.LastUsedAt = &now
	_, err := GetDatabase().DB().Exec("UPDATE passkeys SET last_used_at = $1 WHERE id = $2", now, e.ID)
	return err
}

func (r *PasskeyRepository) UpdateName(e *Passkey) error {
	_, err := GetDatabase().DB().Exec("UPDATE passkeys SET name = $1 WHERE id = $2", e.Name, e.ID)
	return err
}

func (r *PasskeyRepository) Delete(e *Passkey) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM passkeys WHERE id = $1", e.ID)
	return err
}

func (r *PasskeyRepository) DeleteAllByUserID(userID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM passkeys WHERE user_id = $1", userID)
	return err
}

// ToWebAuthnCredential converts a stored Passkey record into a webauthn.Credential
// that can be passed to the webauthn library for assertion verification.
func (p *Passkey) ToWebAuthnCredential() (*webauthn.Credential, error) {
	credentialID, err := decryptBytes(p.CredentialIDEncrypted)
	if err != nil {
		return nil, err
	}
	publicKey, err := decryptBytes(p.PublicKeyEncrypted)
	if err != nil {
		return nil, err
	}
	transports := make([]protocol.AuthenticatorTransport, len(p.Transports))
	for i, t := range p.Transports {
		transports[i] = protocol.AuthenticatorTransport(t)
	}
	cred := &webauthn.Credential{
		ID:              credentialID,
		PublicKey:       publicKey,
		AttestationType: p.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			BackupEligible: p.BackupEligible,
			BackupState:    p.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    p.AAGUID,
			SignCount: p.SignCount,
		},
	}
	return cred, nil
}

// NewPasskeyFromCredential creates a Passkey from a webauthn.Credential returned
// after a successful registration ceremony. It encrypts the sensitive fields.
func NewPasskeyFromCredential(userID string, cred *webauthn.Credential, name string) (*Passkey, error) {
	encCredID, err := encryptBytes(cred.ID)
	if err != nil {
		return nil, err
	}
	encPubKey, err := encryptBytes(cred.PublicKey)
	if err != nil {
		return nil, err
	}
	transports := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}
	p := &Passkey{
		UserID:                userID,
		CredentialIDEncrypted: encCredID,
		CredentialIDHash:      hashBytes(cred.ID),
		PublicKeyEncrypted:    encPubKey,
		AttestationType:       cred.AttestationType,
		AAGUID:                cred.Authenticator.AAGUID,
		SignCount:             cred.Authenticator.SignCount,
		BackupEligible:        cred.Flags.BackupEligible,
		BackupState:           cred.Flags.BackupState,
		Transports:            transports,
		Name:                  name,
	}
	return p, nil
}
