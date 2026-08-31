package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type RoleRouter struct {
}

type CreateRoleRequest struct {
	Name        string         `json:"name" validate:"required,max=128"`
	Description string         `json:"description" validate:"omitempty,max=512"`
	Permissions map[string]int `json:"permissions"`
}

type GetRoleResponse struct {
	ID     string `json:"id"`
	System bool   `json:"system"`
	CreateRoleRequest
}

type GetPermissionDefinitionResponse struct {
	Key           string `json:"key"`
	AllowedLevels []int  `json:"allowedLevels"`
	PluginID      string `json:"pluginId"`
}

type SetUserRolesRequest struct {
	RoleIDs []string `json:"roleIds" validate:"dive,uuid"`
}

type GetUserPermissionsResponse struct {
	Permissions map[string]int `json:"permissions"`
}

func (router *RoleRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/permissions", router.getPermissionCatalogue).Methods("GET")
	s.HandleFunc("/{id}/users", router.getUsers).Methods("GET")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.getAll).Methods("GET")
	s.HandleFunc("/", router.create).Methods("POST")
}

// getPermissionCatalogue lists every permission the role editor can offer,
// including those contributed by plugins. It is available to any authenticated
// user: it describes the model, not the caller's own access.
func (router *RoleRouter) getPermissionCatalogue(w http.ResponseWriter, r *http.Request) {
	res := []*GetPermissionDefinitionResponse{}
	for _, d := range GetPermissionDefinitions() {
		levels := []int{}
		for _, l := range d.AllowedLevels {
			levels = append(levels, int(l))
		}
		res = append(res, &GetPermissionDefinitionResponse{
			Key:           string(d.Key),
			AllowedLevels: levels,
			PluginID:      d.PluginID,
		})
	}
	SendJSON(w, res)
}

func (router *RoleRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CheckPermission(w, user, user.OrganizationID, PermissionRoles, PermissionLevelRead) {
		return
	}
	list, err := GetRoleRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetRoleResponse{}
	for _, e := range list {
		res = append(res, router.copyToRestModel(e))
	}
	SendJSON(w, res)
}

func (router *RoleRouter) getOne(w http.ResponseWriter, r *http.Request) {
	e := router.loadRole(w, r, PermissionLevelRead)
	if e == nil {
		return
	}
	SendJSON(w, router.copyToRestModel(e))
}

func (router *RoleRouter) getUsers(w http.ResponseWriter, r *http.Request) {
	e := router.loadRole(w, r, PermissionLevelAdmin)
	if e == nil {
		return
	}
	userIDs, err := GetUserRoleRepository().GetUserIDsForRole(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if userIDs == nil {
		userIDs = []string{}
	}
	SendJSON(w, userIDs)
}

func (router *RoleRouter) create(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CheckPermission(w, user, user.OrganizationID, PermissionRoles, PermissionLevelAdmin) {
		return
	}
	var m CreateRoleRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	perms, ok := router.parsePermissions(w, m.Permissions)
	if !ok {
		return
	}
	if !CanGrantPermissions(user, user.OrganizationID, perms) {
		SendBadRequestCode(w, ResponseCodeRoleEscalationNotAllowed)
		return
	}
	if existing, err := GetRoleRepository().GetByName(user.OrganizationID, m.Name); err == nil && existing != nil {
		SendBadRequestCode(w, ResponseCodeRoleNameAlreadyExists)
		return
	}
	e := &Role{
		OrganizationID: user.OrganizationID,
		Name:           m.Name,
		Description:    m.Description,
		Permissions:    perms,
	}
	if err := GetRoleRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *RoleRouter) update(w http.ResponseWriter, r *http.Request) {
	e := router.loadRole(w, r, PermissionLevelAdmin)
	if e == nil {
		return
	}
	if e.System {
		SendBadRequestCode(w, ResponseCodeRoleIsSystemRole)
		return
	}
	user := GetRequestUser(r)
	var m CreateRoleRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	perms, ok := router.parsePermissions(w, m.Permissions)
	if !ok {
		return
	}
	// The caller must be able to grant both what the role is gaining and what
	// it already carries, so that a limited administrator can not take over a
	// more powerful role by editing its name.
	if !CanGrantPermissions(user, user.OrganizationID, perms) || !CanGrantPermissions(user, user.OrganizationID, e.Permissions) {
		SendBadRequestCode(w, ResponseCodeRoleEscalationNotAllowed)
		return
	}
	if existing, err := GetRoleRepository().GetByName(user.OrganizationID, m.Name); err == nil && existing != nil && existing.ID != e.ID {
		SendBadRequestCode(w, ResponseCodeRoleNameAlreadyExists)
		return
	}
	e.Name = m.Name
	e.Description = m.Description
	e.Permissions = perms
	// From here on the permission set belongs to the administrator, so stop
	// adding plugin permissions to it behind their back.
	e.AutoGrantPluginPermissions = false
	if err := GetRoleRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// Reducing a role's permissions can strip the organization's last
	// administrator just as surely as deleting the role.
	if !CheckOrgRetainsAdmin(w, user.OrganizationID) {
		return
	}
	SendUpdated(w)
}

func (router *RoleRouter) delete(w http.ResponseWriter, r *http.Request) {
	e := router.loadRole(w, r, PermissionLevelAdmin)
	if e == nil {
		return
	}
	if e.System {
		SendBadRequestCode(w, ResponseCodeRoleIsSystemRole)
		return
	}
	user := GetRequestUser(r)
	if !CanGrantPermissions(user, user.OrganizationID, e.Permissions) {
		SendBadRequestCode(w, ResponseCodeRoleEscalationNotAllowed)
		return
	}
	if err := GetRoleRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if !CheckOrgRetainsAdmin(w, user.OrganizationID) {
		return
	}
	SendUpdated(w)
}

// loadRole resolves the role named in the path, enforcing the required level
// and organization scoping. It writes the error response and returns nil when
// the request should not proceed.
func (router *RoleRouter) loadRole(w http.ResponseWriter, r *http.Request, min PermissionLevel) *Role {
	user := GetRequestUser(r)
	if !CheckPermission(w, user, user.OrganizationID, PermissionRoles, min) {
		return nil
	}
	vars := mux.Vars(r)
	e, err := GetRoleRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return nil
	}
	if e.OrganizationID != user.OrganizationID {
		SendNotFound(w)
		return nil
	}
	return e
}

// parsePermissions validates a submitted permission map against the
// catalogue. Unknown keys and levels a permission does not declare are
// rejected rather than silently dropped, so that a typo does not quietly
// produce a role weaker than intended.
func (router *RoleRouter) parsePermissions(w http.ResponseWriter, m map[string]int) (map[Permission]PermissionLevel, bool) {
	res := make(map[Permission]PermissionLevel)
	for key, level := range m {
		if PermissionLevel(level) == PermissionLevelNone {
			continue
		}
		if !IsValidPermissionLevel(Permission(key), PermissionLevel(level)) {
			SendBadRequest(w)
			return nil, false
		}
		res[Permission(key)] = PermissionLevel(level)
	}
	return res, true
}

func (router *RoleRouter) copyToRestModel(e *Role) *GetRoleResponse {
	m := &GetRoleResponse{}
	m.ID = e.ID
	m.System = e.System
	m.Name = e.Name
	m.Description = e.Description
	perms := e.Permissions
	if e.System {
		// A system role grants whatever the catalogue currently holds, which
		// includes permissions registered by plugins after the role was
		// seeded. Reporting the stored set would show those as "no access" on
		// a role that in fact grants them.
		perms = make(map[Permission]PermissionLevel)
		for _, d := range GetPermissionDefinitions() {
			perms[d.Key] = d.MaxLevel()
		}
	}
	m.Permissions = PermissionsToRestModel(perms)
	return m
}
