package repository

import (
	"log"
	"sync"

	"github.com/lib/pq"
)

type SpaceImageRepository struct {
}

type SpaceImage struct {
	SpaceID  string
	MimeType string
	Width    uint
	Height   uint
	Data     []byte
}

var spaceImageRepository *SpaceImageRepository
var spaceImageRepositoryOnce sync.Once

func GetSpaceImageRepository() *SpaceImageRepository {
	spaceImageRepositoryOnce.Do(func() {
		spaceImageRepository = &SpaceImageRepository{}
		spaceImageRepository.createTable()
	})
	return spaceImageRepository
}

func (r *SpaceImageRepository) createTable() {
	if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS space_images (" +
		"space_id uuid PRIMARY KEY REFERENCES spaces(id) ON DELETE CASCADE, " +
		"mime_type VARCHAR(32) NOT NULL, " +
		"width INTEGER NOT NULL, " +
		"height INTEGER NOT NULL, " +
		"data BYTEA NOT NULL)"); err != nil {
		log.Panicf("failed to create space_images table: %v", err)
	}
}

func (r *SpaceImageRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 67 {
		r.createTable()
	}
}

func (r *SpaceImageRepository) Get(spaceID, organizationID string) (*SpaceImage, error) {
	e := &SpaceImage{}
	err := GetDatabase().DB().QueryRow(
		"SELECT space_images.space_id, space_images.mime_type, space_images.width, space_images.height, space_images.data "+
			"FROM space_images "+
			"INNER JOIN spaces ON spaces.id = space_images.space_id "+
			"INNER JOIN locations ON locations.id = spaces.location_id "+
			"WHERE space_images.space_id = $1 AND locations.organization_id = $2",
		spaceID, organizationID).Scan(&e.SpaceID, &e.MimeType, &e.Width, &e.Height, &e.Data)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SpaceImageRepository) Set(e *SpaceImage) error {
	_, err := GetDatabase().DB().Exec(
		"INSERT INTO space_images (space_id, mime_type, width, height, data) "+
			"VALUES ($1, $2, $3, $4, $5) "+
			"ON CONFLICT (space_id) DO UPDATE SET "+
			"mime_type = EXCLUDED.mime_type, width = EXCLUDED.width, height = EXCLUDED.height, data = EXCLUDED.data",
		e.SpaceID, e.MimeType, e.Width, e.Height, e.Data)
	return err
}

func (r *SpaceImageRepository) Delete(spaceID, organizationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM space_images USING spaces, locations "+
		"WHERE space_images.space_id = $1 AND spaces.id = space_images.space_id "+
		"AND locations.id = spaces.location_id AND locations.organization_id = $2",
		spaceID, organizationID)
	return err
}

// GetSpaceIDsWithImage returns the subset of spaceIDs that have an image.
func (r *SpaceImageRepository) GetSpaceIDsWithImage(spaceIDs []string, organizationID string) (map[string]bool, error) {
	res := map[string]bool{}
	if len(spaceIDs) == 0 {
		return res, nil
	}
	rows, err := GetDatabase().DB().Query("SELECT space_images.space_id FROM space_images "+
		"INNER JOIN spaces ON spaces.id = space_images.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE space_images.space_id = ANY($1) AND locations.organization_id = $2",
		pq.Array(spaceIDs), organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		res[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
