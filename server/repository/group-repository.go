package repository

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lib/pq"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type GroupStore struct {
}

var groupRepository *GroupStore
var groupRepositoryOnce sync.Once

func GetGroupRepository() *GroupStore {
	groupRepositoryOnce.Do(func() {
		groupRepository = &GroupStore{}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS groups (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"PRIMARY KEY (id))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS users_groups (" +
			"group_id uuid NOT NULL, " +
			"user_id uuid NOT NULL, " +
			"PRIMARY KEY (group_id, user_id))"); err != nil {
			panic(err)
		}
	})
	return groupRepository
}

func (r *GroupStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	// Nothing yet
}

func (r *GroupStore) Create(e *Group) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO groups "+
		"(organization_id, name) "+
		"VALUES ($1, $2) "+
		"RETURNING id",
		e.OrganizationID, e.Name).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *GroupStore) GetOne(id string) (*Group, error) {
	e := &Group{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Name)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *GroupStore) GetAll(organizationID string) ([]*Group, error) {
	var result []*Group
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE organization_id = $1 "+
		"ORDER BY name",
		organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Group{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *GroupStore) GetAllByIDs(groupIDs []string) ([]*Group, error) {
	var result []*Group
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE id = ANY($1) "+
		"ORDER BY name",
		pq.Array(groupIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Group{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *GroupStore) GetAllWhereUserIsMember(userID string) ([]*Group, error) {
	var result []*Group
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE id IN (SELECT group_id FROM users_groups WHERE user_id = $1) "+
		"ORDER BY name",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Group{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *GroupStore) GetByKeyword(organizationID string, keyword string) ([]*Group, error) {
	var result []*Group
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE organization_id = $1 AND LOWER(name) LIKE '%' || $2 || '%' "+
		"ORDER BY name", organizationID, strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Group{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *GroupStore) GetByName(organizationID string, name string) (*Group, error) {
	e := &Group{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name "+
		"FROM groups "+
		"WHERE organization_id = $1 AND LOWER(name) = $2", organizationID, strings.ToLower(name)).Scan(&e.ID, &e.OrganizationID, &e.Name)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *GroupStore) GroupsExistAndBelongToOrg(organizationID string, groupIDs []string) (bool, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) "+
		"FROM groups "+
		"WHERE id = ANY($1) AND organization_id = $2",
		pq.Array(groupIDs), organizationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == len(groupIDs), nil
}

func (r *GroupStore) Update(e *Group) error {
	_, err := GetDatabase().DB().Exec("UPDATE groups SET "+
		"organization_id = $1, "+
		"name = $2 "+
		"WHERE id = $3",
		e.OrganizationID, e.Name, e.ID)
	return err
}

func (r *GroupStore) Delete(e *Group) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_groups WHERE "+
		"group_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM groups WHERE id = $1", e.ID)
	return err
}

func (r *GroupStore) DeleteAll(organizationID string) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM users_groups WHERE "+
		"group_id IN (SELECT id from groups WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM groups WHERE organization_id = $1", organizationID)
	return err
}

func (r *GroupStore) GetMemberUserIDs(e *Group) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT user_id "+
		"FROM users_groups "+
		"WHERE group_id = $1 "+
		"ORDER BY user_id",
		e.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func (r *GroupStore) GetUserCountMap(organizationID string) (map[string]int, error) {
	res := make(map[string]int)
	rows, err := GetDatabase().DB().Query("SELECT users_groups.group_id, COUNT(users_groups.user_id) "+
		"FROM users_groups "+
		"INNER JOIN groups ON groups.id = users_groups.group_id "+
		"WHERE groups.organization_id = $1 "+
		"GROUP BY users_groups.group_id",
		organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var groupID string
		var count int
		err = rows.Scan(&groupID, &count)
		if err != nil {
			return nil, err
		}
		res[groupID] = count
	}
	return res, nil
}

func (r *GroupStore) AddMembers(e *Group, userIDs []string) error {
	sqlStr := "INSERT INTO users_groups (group_id, user_id) VALUES "
	vals := []interface{}{}
	i := 1
	for _, userID := range userIDs {
		sqlStr += fmt.Sprintf("($%d, $%d),", i, i+1)
		i += 2
		vals = append(vals, e.ID, userID)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	_, err := GetDatabase().DB().Exec(sqlStr, vals...)
	return err
}

func (r *GroupStore) RemoveMembers(e *Group, userIDs []string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM users_groups WHERE group_id = $1 AND user_id = ANY($2)", e.ID, pq.Array(userIDs))
	return err
}
