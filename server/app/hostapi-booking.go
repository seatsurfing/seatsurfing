package app

import (
	"errors"
	"net/http"
	"time"

	"github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
)

// User-scoped booking operations of the host API. They are implemented by
// the services in package service, which the REST handlers use as well, so a
// plugin acting on behalf of a user gets exactly that user's permissions.

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

func searchAttributesFromAPI(in []api.SearchAttributeFilter) ([]service.SearchAttribute, error) {
	out := make([]service.SearchAttribute, 0, len(in))
	for _, a := range in {
		out = append(out, service.SearchAttribute{AttributeID: a.AttributeID, Comparator: a.Comparator, Value: a.Value})
	}
	return out, service.ValidateSearchAttributes(out)
}

func attributeValuesToAPI(in []*SpaceAttributeValue) []api.AttributeValue {
	out := make([]api.AttributeValue, 0, len(in))
	for _, v := range in {
		out = append(out, api.AttributeValue{AttributeID: v.AttributeID, Value: v.Value})
	}
	return out
}

func (h *hostAPIImpl) SearchLocationsForUser(userID string, enter, leave time.Time, attributes []api.SearchAttributeFilter) ([]*api.LocationInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	attrs, err := searchAttributesFromAPI(attributes)
	if err != nil {
		return nil, err
	}
	locations, err := service.GetLocationService().SearchLocationsForUser(user, enter, leave, attrs)
	if err != nil {
		return nil, err
	}
	res := []*api.LocationInfo{}
	for _, l := range locations {
		res = append(res, &api.LocationInfo{
			Location:     *l.Location,
			Attributes:   attributeValuesToAPI(l.Attributes),
			Allowed:      l.AllowedForUser,
			BookableDays: service.WeekdaysFromString(l.Location.BookableDays),
		})
	}
	return res, nil
}

func (h *hostAPIImpl) GetSpaceAvailabilityForUser(userID, locationID string, enter, leave time.Time, attributes []api.SearchAttributeFilter) ([]*api.SpaceAvailabilityInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	location, err := GetLocationRepository().GetOne(locationID)
	if err != nil || location == nil || !service.CanAccessOrg(user, location.OrganizationID) {
		return nil, errors.New("location not found")
	}
	attrs, err := searchAttributesFromAPI(attributes)
	if err != nil {
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
	list, err := service.GetSpaceService().GetAvailabilityForUser(user, location, "", enter, leave, attrs)
	if err != nil {
		return nil, err
	}
	// Other users' bookings (e.Bookings) are deliberately not exposed.
	res := []*api.SpaceAvailabilityInfo{}
	for _, s := range list {
		res = append(res, &api.SpaceAvailabilityInfo{
			Space:            s.Space,
			Attributes:       attributeValuesToAPI(s.Attributes),
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
		})
	}
	return res, nil
}

func (h *hostAPIImpl) GetSpaceAttributesForUser(userID string) ([]*api.SpaceAttributeDefinition, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	list, err := GetSpaceAttributeRepository().GetAll(user.OrganizationID)
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

// bookingErrorStatus maps a booking service error to the HTTP status the REST
// API responds with, which is what BookingCreateResult reports.
func bookingErrorStatus(kind service.BookingErrorKind) int {
	switch kind {
	case service.BookingErrorForbidden:
		return http.StatusForbidden
	case service.BookingErrorConflict:
		return http.StatusConflict
	case service.BookingErrorInternal:
		return http.StatusInternalServerError
	case service.BookingErrorNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func (h *hostAPIImpl) CreateBookingForUser(userID, spaceID string, enter, leave time.Time, subject string) (*api.BookingCreateResult, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	e, bErr := service.GetBookingService().CreateBooking(user, &service.BookingInput{
		SpaceID: spaceID,
		Subject: subject,
		Enter:   enter,
		Leave:   leave,
	})
	if bErr != nil {
		return &api.BookingCreateResult{StatusCode: bookingErrorStatus(bErr.Kind), ErrorCode: bErr.Code}, nil
	}
	return &api.BookingCreateResult{BookingID: e.ID, Approved: e.Approved, StatusCode: http.StatusCreated}, nil
}

func (h *hostAPIImpl) GetUpcomingBookingsForUser(userID string) ([]*api.BookingDetails, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	return service.GetBookingService().GetUpcomingBookingsForUser(user)
}

func (h *hostAPIImpl) DeleteBookingForUser(userID, bookingID string) (*api.BookingDeleteResult, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	if bErr := service.GetBookingService().DeleteBooking(user, bookingID); bErr != nil {
		return &api.BookingDeleteResult{StatusCode: bookingErrorStatus(bErr.Kind), ErrorCode: bErr.Code}, nil
	}
	return &api.BookingDeleteResult{StatusCode: http.StatusNoContent}, nil
}
