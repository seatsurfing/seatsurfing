package app

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
)

// User-scoped booking operations of the host API. They share their logic
// with the REST handlers (see the *ForUser functions in package router), so
// a plugin acting on behalf of a user gets exactly that user's permissions.

var errHostAPIUserNotAllowed = errors.New("user not found or disabled")

// getActiveUser loads a user who may act via the host API: existing and not
// disabled, like VerifyAuthMiddleware requires for REST calls.
func getActiveUser(userID string) (*api.User, error) {
	user, err := GetUserRepository().GetOne(userID)
	if err != nil || user == nil || user.Disabled {
		return nil, errHostAPIUserNotAllowed
	}
	return user, nil
}

func searchAttributesFromAPI(in []api.SearchAttributeFilter) []SearchAttribute {
	out := make([]SearchAttribute, 0, len(in))
	for _, a := range in {
		out = append(out, SearchAttribute{AttributeID: a.AttributeID, Comparator: a.Comparator, Value: a.Value})
	}
	return out
}

func userGroupIDs(userID string) ([]string, error) {
	groups, err := GetGroupRepository().GetAllWhereUserIsMember(userID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids, nil
}

func (h *hostAPIImpl) SearchLocationsForUser(userID string, enter, leave time.Time, attributes []api.SearchAttributeFilter) ([]*api.LocationInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	req := &SearchLocationRequest{Enter: enter, Leave: leave, Attributes: searchAttributesFromAPI(attributes)}
	if err := GetValidator().Struct(req); err != nil {
		return nil, err
	}
	locations, err := (&LocationRouter{}).SearchLocationsForUser(user, req)
	if err != nil {
		return nil, err
	}
	groupIDs, err := userGroupIDs(user.ID)
	if err != nil {
		return nil, err
	}
	locationIDs := make([]string, 0, len(locations))
	for _, l := range locations {
		locationIDs = append(locationIDs, l.ID)
	}
	attributeValues, err := GetSpaceAttributeValueRepository().GetAllForEntityList(locationIDs, SpaceAttributeValueEntityTypeLocation)
	if err != nil {
		return nil, err
	}
	res := []*api.LocationInfo{}
	for _, l := range locations {
		allowed := len(l.AllowedBookerGroupIDs) == 0
		for _, id := range l.AllowedBookerGroupIDs {
			if slices.Contains(groupIDs, id) {
				allowed = true
				break
			}
		}
		info := &api.LocationInfo{
			Location: api.Location{
				ID:                    l.ID,
				OrganizationID:        l.OrganizationID,
				Name:                  l.Name,
				MapWidth:              l.MapWidth,
				MapHeight:             l.MapHeight,
				MapScale:              l.MapScale,
				MapMimeType:           l.MapMimeType,
				MapType:               l.MapType,
				Description:           l.Description,
				MaxConcurrentBookings: l.MaxConcurrentBookings,
				Timezone:              l.Timezone,
				Enabled:               l.Enabled,
			},
			Attributes:   []api.AttributeValue{},
			Allowed:      allowed,
			BookableDays: l.BookableDays,
		}
		for _, v := range attributeValues {
			if v.EntityID == l.ID {
				info.Attributes = append(info.Attributes, api.AttributeValue{AttributeID: v.AttributeID, Value: v.Value})
			}
		}
		res = append(res, info)
	}
	return res, nil
}

func (h *hostAPIImpl) GetSpaceAvailabilityForUser(userID, locationID string, enter, leave time.Time, attributes []api.SearchAttributeFilter) ([]*api.SpaceAvailabilityInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	location, err := GetLocationRepository().GetOne(locationID)
	if err != nil || location == nil || !CanAccessOrg(user, location.OrganizationID) {
		return nil, errors.New("location not found")
	}
	attrs := searchAttributesFromAPI(attributes)
	if err := GetValidator().Var(attrs, "dive"); err != nil {
		return nil, err
	}
	enter, err = GetLocationRepository().AttachTimezoneInformation(enter, location)
	if err != nil {
		return nil, err
	}
	leave, err = GetLocationRepository().AttachTimezoneInformation(leave, location)
	if err != nil {
		return nil, err
	}
	list, err := (&SpaceRouter{}).GetSpaceAvailabilityForUser(user, location, "", enter, leave, attrs)
	if err != nil {
		return nil, err
	}
	res := []*api.SpaceAvailabilityInfo{}
	for _, s := range list {
		info := &api.SpaceAvailabilityInfo{
			Space: api.Space{
				ID:                   s.ID,
				LocationID:           s.LocationID,
				Name:                 s.Name,
				X:                    s.X,
				Y:                    s.Y,
				Width:                s.Width,
				Height:               s.Height,
				Rotation:             s.Rotation,
				RequireSubject:       s.RequireSubject,
				Enabled:              s.Enabled,
				KioskEnabled:         s.KioskEnabled,
				Shape:                s.Shape,
				FontSize:             s.FontSize,
				PublicBookingEnabled: s.PublicBookingEnabled,
			},
			Attributes:       []api.AttributeValue{},
			Available:        s.Available,
			Allowed:          s.IsAllowed,
			ApprovalRequired: s.IsApprovalRequired,
		}
		for _, a := range s.Attributes {
			info.Attributes = append(info.Attributes, api.AttributeValue{AttributeID: a.AttributeID, Value: a.Value})
		}
		res = append(res, info)
	}
	return res, nil
}

func (h *hostAPIImpl) GetSpaceAttributes(organizationID string) ([]*api.SpaceAttributeDefinition, error) {
	list, err := GetSpaceAttributeRepository().GetAll(organizationID)
	if err != nil {
		return nil, err
	}
	res := make([]*api.SpaceAttributeDefinition, 0, len(list))
	for _, a := range list {
		res = append(res, &api.SpaceAttributeDefinition{
			ID:                 a.ID,
			OrganizationID:     a.OrganizationID,
			Label:              a.Label,
			Type:               int(a.Type),
			SpaceApplicable:    a.SpaceApplicable,
			LocationApplicable: a.LocationApplicable,
		})
	}
	return res, nil
}

func (h *hostAPIImpl) CreateBookingForUser(userID, spaceID string, enter, leave time.Time, subject string) (*api.BookingCreateResult, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	m := &CreateBookingRequest{
		SpaceID: spaceID,
		Subject: subject,
		BookingRequest: BookingRequest{
			Enter: enter,
			Leave: leave,
		},
	}
	if err := GetValidator().Struct(m); err != nil {
		return &api.BookingCreateResult{StatusCode: http.StatusBadRequest}, nil
	}
	e, bErr := (&BookingRouter{}).CreateBookingForUser(user, m)
	if bErr != nil {
		return &api.BookingCreateResult{StatusCode: bErr.StatusCode, ErrorCode: bErr.Code}, nil
	}
	return &api.BookingCreateResult{BookingID: e.ID, Approved: e.Approved, StatusCode: http.StatusCreated}, nil
}

func (h *hostAPIImpl) GetUpcomingBookingsForUser(userID string) ([]*api.BookingDetails, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	return GetUpcomingBookingsForUser(user)
}
