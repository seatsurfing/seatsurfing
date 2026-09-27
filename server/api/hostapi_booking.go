package api

import (
	"context"
	"time"

	"github.com/seatsurfing/seatsurfing/server/api/hostapipb"
)

// User-scoped booking operations exposed to plugins. Each call acts as the
// given user and applies exactly the rules of the corresponding REST
// endpoint, so a plugin can't bypass booking restrictions.
//
// Times are wall-clock times: their timezone is ignored and replaced by the
// location's timezone (as the REST API does with incoming times). Callers
// should build them in UTC, which is what survives the gRPC round trip.

// SearchAttributeFilter filters locations or spaces by attribute value. AttributeID
// is a space attribute ID, or one of the synthetic location attributes
// "numSpaces", "numFreeSpaces" and "buddyOnSite". Comparator is one of eq,
// neq, contains, ncontains, gt, gte, lt, lte.
type SearchAttributeFilter struct {
	AttributeID string
	Comparator  string
	Value       string
}

// AttributeValue is the value of one attribute on a location or space.
type AttributeValue struct {
	AttributeID string
	Value       string
}

// SpaceAttributeDefinition is an attribute definition of an organization.
type SpaceAttributeDefinition struct {
	ID                 string
	OrganizationID     string
	Label              string
	Type               int
	SpaceApplicable    bool
	LocationApplicable bool
}

// LocationInfo is a location as seen by a specific user.
type LocationInfo struct {
	Location   Location
	Attributes []AttributeValue
	// Allowed reports whether the user is in the location's allowed booker
	// groups (or the location has none).
	Allowed bool
	// BookableDays are the weekdays (0 = Sunday) bookings may be placed on.
	BookableDays []int
}

// SpaceAvailabilityInfo is a space's availability as seen by a specific user.
// Other users' bookings are deliberately not included.
type SpaceAvailabilityInfo struct {
	Space            Space
	Attributes       []AttributeValue
	Available        bool
	Allowed          bool
	ApprovalRequired bool
}

// BookingCreateResult is the outcome of CreateBookingForUser. On success,
// BookingID is set. Otherwise StatusCode is the HTTP status the REST API
// would have returned and ErrorCode its X-Error-Code value (0 if none).
type BookingCreateResult struct {
	BookingID  string
	Approved   bool
	StatusCode int
	ErrorCode  int
}

// BookingDeleteResult is the outcome of DeleteBookingForUser. StatusCode is
// the HTTP status the REST API would have returned (204 on success) and
// ErrorCode its X-Error-Code value (0 if none).
type BookingDeleteResult struct {
	StatusCode int
	ErrorCode  int
}

// ─── Plugin side (gRPC client) ───────────────────────────────────────────────

func (h *HostAPIGRPC) SearchLocationsForUser(userID string, enter, leave time.Time, attributes []SearchAttributeFilter) ([]*LocationInfo, error) {
	reply, err := h.client.SearchLocationsForUser(context.Background(), &hostapipb.SearchLocationsForUserArgs{
		UserId:     userID,
		Enter:      timeToProto(enter),
		Leave:      timeToProto(leave),
		Attributes: searchAttributesToProto(attributes),
	})
	if err != nil {
		return nil, err
	}
	return locationInfosFromProto(reply.Locations), strErr(reply.Err)
}

func (h *HostAPIGRPC) GetSpaceAvailabilityForUser(userID, locationID string, enter, leave time.Time, attributes []SearchAttributeFilter) ([]*SpaceAvailabilityInfo, error) {
	reply, err := h.client.GetSpaceAvailabilityForUser(context.Background(), &hostapipb.GetSpaceAvailabilityForUserArgs{
		UserId:     userID,
		LocationId: locationID,
		Enter:      timeToProto(enter),
		Leave:      timeToProto(leave),
		Attributes: searchAttributesToProto(attributes),
	})
	if err != nil {
		return nil, err
	}
	return spaceAvailabilityInfosFromProto(reply.Spaces), strErr(reply.Err)
}

func (h *HostAPIGRPC) GetSpaceAttributesForUser(userID string) ([]*SpaceAttributeDefinition, error) {
	reply, err := h.client.GetSpaceAttributesForUser(context.Background(), &hostapipb.GetSpaceAttributesForUserArgs{UserId: userID})
	if err != nil {
		return nil, err
	}
	return spaceAttributesFromProto(reply.Attributes), strErr(reply.Err)
}

func (h *HostAPIGRPC) CreateBookingForUser(userID, spaceID string, enter, leave time.Time, subject string) (*BookingCreateResult, error) {
	reply, err := h.client.CreateBookingForUser(context.Background(), &hostapipb.CreateBookingForUserArgs{
		UserId:  userID,
		SpaceId: spaceID,
		Enter:   timeToProto(enter),
		Leave:   timeToProto(leave),
		Subject: subject,
	})
	if err != nil {
		return nil, err
	}
	if reply.Err != "" {
		return nil, strErr(reply.Err)
	}
	return &BookingCreateResult{
		BookingID:  reply.BookingId,
		Approved:   reply.Approved,
		StatusCode: int(reply.StatusCode),
		ErrorCode:  int(reply.ErrorCode),
	}, nil
}

func (h *HostAPIGRPC) GetUpcomingBookingsForUser(userID string) ([]*BookingDetails, error) {
	reply, err := h.client.GetUpcomingBookingsForUser(context.Background(), &hostapipb.GetUpcomingBookingsForUserArgs{UserId: userID})
	if err != nil {
		return nil, err
	}
	res := make([]*BookingDetails, 0, len(reply.Bookings))
	for _, b := range reply.Bookings {
		res = append(res, bookingDetailsFromProto(b))
	}
	return res, strErr(reply.Err)
}

func (h *HostAPIGRPC) DeleteBookingForUser(userID, bookingID string) (*BookingDeleteResult, error) {
	reply, err := h.client.DeleteBookingForUser(context.Background(), &hostapipb.DeleteBookingForUserArgs{UserId: userID, BookingId: bookingID})
	if err != nil {
		return nil, err
	}
	if reply.Err != "" {
		return nil, strErr(reply.Err)
	}
	return &BookingDeleteResult{StatusCode: int(reply.StatusCode), ErrorCode: int(reply.ErrorCode)}, nil
}

// ─── Host side (gRPC server) ─────────────────────────────────────────────────

func (s *HostAPIGRPCServer) SearchLocationsForUser(ctx context.Context, a *hostapipb.SearchLocationsForUserArgs) (*hostapipb.SearchLocationsForUserReply, error) {
	v, err := s.impl.SearchLocationsForUser(a.UserId, timeFromProto(a.Enter), timeFromProto(a.Leave), searchAttributesFromProto(a.Attributes))
	return &hostapipb.SearchLocationsForUserReply{Locations: locationInfosToProto(v), Err: errStr(err)}, nil
}

func (s *HostAPIGRPCServer) GetSpaceAvailabilityForUser(ctx context.Context, a *hostapipb.GetSpaceAvailabilityForUserArgs) (*hostapipb.GetSpaceAvailabilityForUserReply, error) {
	v, err := s.impl.GetSpaceAvailabilityForUser(a.UserId, a.LocationId, timeFromProto(a.Enter), timeFromProto(a.Leave), searchAttributesFromProto(a.Attributes))
	return &hostapipb.GetSpaceAvailabilityForUserReply{Spaces: spaceAvailabilityInfosToProto(v), Err: errStr(err)}, nil
}

func (s *HostAPIGRPCServer) GetSpaceAttributesForUser(ctx context.Context, a *hostapipb.GetSpaceAttributesForUserArgs) (*hostapipb.GetSpaceAttributesReply, error) {
	v, err := s.impl.GetSpaceAttributesForUser(a.UserId)
	return &hostapipb.GetSpaceAttributesReply{Attributes: spaceAttributesToProto(v), Err: errStr(err)}, nil
}

func (s *HostAPIGRPCServer) CreateBookingForUser(ctx context.Context, a *hostapipb.CreateBookingForUserArgs) (*hostapipb.CreateBookingForUserReply, error) {
	v, err := s.impl.CreateBookingForUser(a.UserId, a.SpaceId, timeFromProto(a.Enter), timeFromProto(a.Leave), a.Subject)
	if err != nil || v == nil {
		return &hostapipb.CreateBookingForUserReply{Err: errStr(err)}, nil
	}
	return &hostapipb.CreateBookingForUserReply{
		BookingId:  v.BookingID,
		Approved:   v.Approved,
		StatusCode: int32(v.StatusCode),
		ErrorCode:  int32(v.ErrorCode),
	}, nil
}

func (s *HostAPIGRPCServer) GetUpcomingBookingsForUser(ctx context.Context, a *hostapipb.GetUpcomingBookingsForUserArgs) (*hostapipb.GetUpcomingBookingsForUserReply, error) {
	v, err := s.impl.GetUpcomingBookingsForUser(a.UserId)
	reply := &hostapipb.GetUpcomingBookingsForUserReply{Err: errStr(err)}
	for _, b := range v {
		reply.Bookings = append(reply.Bookings, bookingDetailsToProto(b))
	}
	return reply, nil
}

func (s *HostAPIGRPCServer) DeleteBookingForUser(ctx context.Context, a *hostapipb.DeleteBookingForUserArgs) (*hostapipb.DeleteBookingForUserReply, error) {
	v, err := s.impl.DeleteBookingForUser(a.UserId, a.BookingId)
	if err != nil || v == nil {
		return &hostapipb.DeleteBookingForUserReply{Err: errStr(err)}, nil
	}
	return &hostapipb.DeleteBookingForUserReply{StatusCode: int32(v.StatusCode), ErrorCode: int32(v.ErrorCode)}, nil
}

// ─── Conversions ─────────────────────────────────────────────────────────────

func searchAttributesToProto(in []SearchAttributeFilter) []*hostapipb.SearchAttribute {
	out := make([]*hostapipb.SearchAttribute, 0, len(in))
	for _, a := range in {
		out = append(out, &hostapipb.SearchAttribute{AttributeId: a.AttributeID, Comparator: a.Comparator, Value: a.Value})
	}
	return out
}

func searchAttributesFromProto(in []*hostapipb.SearchAttribute) []SearchAttributeFilter {
	out := make([]SearchAttributeFilter, 0, len(in))
	for _, a := range in {
		out = append(out, SearchAttributeFilter{AttributeID: a.AttributeId, Comparator: a.Comparator, Value: a.Value})
	}
	return out
}

func attributeValuesToProto(in []AttributeValue) []*hostapipb.AttributeValue {
	out := make([]*hostapipb.AttributeValue, 0, len(in))
	for _, a := range in {
		out = append(out, &hostapipb.AttributeValue{AttributeId: a.AttributeID, Value: a.Value})
	}
	return out
}

func attributeValuesFromProto(in []*hostapipb.AttributeValue) []AttributeValue {
	out := make([]AttributeValue, 0, len(in))
	for _, a := range in {
		out = append(out, AttributeValue{AttributeID: a.AttributeId, Value: a.Value})
	}
	return out
}

func spaceAttributesToProto(in []*SpaceAttributeDefinition) []*hostapipb.SpaceAttribute {
	out := make([]*hostapipb.SpaceAttribute, 0, len(in))
	for _, a := range in {
		out = append(out, &hostapipb.SpaceAttribute{
			Id:                 a.ID,
			OrganizationId:     a.OrganizationID,
			Label:              a.Label,
			Type:               int32(a.Type),
			SpaceApplicable:    a.SpaceApplicable,
			LocationApplicable: a.LocationApplicable,
		})
	}
	return out
}

func spaceAttributesFromProto(in []*hostapipb.SpaceAttribute) []*SpaceAttributeDefinition {
	out := make([]*SpaceAttributeDefinition, 0, len(in))
	for _, a := range in {
		out = append(out, &SpaceAttributeDefinition{
			ID:                 a.Id,
			OrganizationID:     a.OrganizationId,
			Label:              a.Label,
			Type:               int(a.Type),
			SpaceApplicable:    a.SpaceApplicable,
			LocationApplicable: a.LocationApplicable,
		})
	}
	return out
}

func locationInfosToProto(in []*LocationInfo) []*hostapipb.LocationInfo {
	out := make([]*hostapipb.LocationInfo, 0, len(in))
	for _, l := range in {
		days := make([]int32, 0, len(l.BookableDays))
		for _, d := range l.BookableDays {
			// Weekdays are 0 (Sunday) to 6; anything else is invalid data.
			if d >= 0 && d <= 6 {
				days = append(days, int32(d))
			}
		}
		out = append(out, &hostapipb.LocationInfo{
			Location:     locationToProto(&l.Location),
			Attributes:   attributeValuesToProto(l.Attributes),
			Allowed:      l.Allowed,
			BookableDays: days,
		})
	}
	return out
}

func locationInfosFromProto(in []*hostapipb.LocationInfo) []*LocationInfo {
	out := make([]*LocationInfo, 0, len(in))
	for _, l := range in {
		days := make([]int, 0, len(l.BookableDays))
		for _, d := range l.BookableDays {
			days = append(days, int(d))
		}
		info := &LocationInfo{
			Attributes:   attributeValuesFromProto(l.Attributes),
			Allowed:      l.Allowed,
			BookableDays: days,
		}
		if loc := locationFromProto(l.Location); loc != nil {
			info.Location = *loc
		}
		out = append(out, info)
	}
	return out
}

func spaceAvailabilityInfosToProto(in []*SpaceAvailabilityInfo) []*hostapipb.SpaceAvailabilityInfo {
	out := make([]*hostapipb.SpaceAvailabilityInfo, 0, len(in))
	for _, s := range in {
		out = append(out, &hostapipb.SpaceAvailabilityInfo{
			Space:            spaceToProto(&s.Space),
			Attributes:       attributeValuesToProto(s.Attributes),
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
		})
	}
	return out
}

func spaceAvailabilityInfosFromProto(in []*hostapipb.SpaceAvailabilityInfo) []*SpaceAvailabilityInfo {
	out := make([]*SpaceAvailabilityInfo, 0, len(in))
	for _, s := range in {
		info := &SpaceAvailabilityInfo{
			Attributes:       attributeValuesFromProto(s.Attributes),
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
		}
		if sp := spaceFromProto(s.Space); sp != nil {
			info.Space = *sp
		}
		out = append(out, info)
	}
	return out
}
