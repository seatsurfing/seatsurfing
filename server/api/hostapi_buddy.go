package api

import (
	"context"
	"time"

	"github.com/seatsurfing/seatsurfing/server/api/hostapipb"
)

// User-scoped buddy operations exposed to plugins. Buddies are the
// colleagues a user follows (see repository.BuddyRepository); every call
// acts as the given user and returns an empty/zero result without an error
// if the buddies feature is disabled for their organization.

// BuddyInfo is one of a user's buddies.
type BuddyInfo struct {
	ID          string
	BuddyUserID string
	Email       string
	Firstname   string
	Lastname    string
}

// BuddyScheduleInfo is where a buddy is booked in a time slot.
type BuddyScheduleInfo struct {
	Buddy        BuddyInfo
	LocationID   string
	LocationName string
	SpaceID      string
	SpaceName    string
	Enter, Leave time.Time
}

// NearBuddySpaceInfo is a space available to a user, ranked by proximity to
// where their buddies are seated.
type NearBuddySpaceInfo struct {
	Space            Space
	Available        bool
	Allowed          bool
	ApprovalRequired bool
	// Distance is the (average, if several) Euclidean distance in floor-plan
	// units to the buddies seated in this space's location.
	Distance     float64
	NearBuddyIDs []string
}

// NearBuddiesResult is the outcome of FindSpacesNearBuddiesForUser.
// LocationID is empty if none of the requested buddies have a booking in
// the time slot.
type NearBuddiesResult struct {
	LocationID    string
	LocationName  string
	Spaces        []*NearBuddySpaceInfo
	FoundBuddyIDs []string
}

// ─── Plugin side (gRPC client) ───────────────────────────────────────────────

func (h *HostAPIGRPC) GetBuddiesForUser(userID string) ([]*BuddyInfo, error) {
	reply, err := h.client.GetBuddiesForUser(context.Background(), &hostapipb.GetBuddiesForUserArgs{UserId: userID})
	if err != nil {
		return nil, err
	}
	return buddyInfosFromProto(reply.Buddies), strErr(reply.Err)
}

func (h *HostAPIGRPC) GetBuddyScheduleForUser(userID string, buddyIDs []string, enter, leave time.Time) ([]*BuddyScheduleInfo, error) {
	reply, err := h.client.GetBuddyScheduleForUser(context.Background(), &hostapipb.GetBuddyScheduleForUserArgs{
		UserId:   userID,
		BuddyIds: buddyIDs,
		Enter:    timeToProto(enter),
		Leave:    timeToProto(leave),
	})
	if err != nil {
		return nil, err
	}
	return buddyScheduleInfosFromProto(reply.Schedule), strErr(reply.Err)
}

func (h *HostAPIGRPC) FindSpacesNearBuddiesForUser(userID string, buddyIDs []string, enter, leave time.Time) (*NearBuddiesResult, error) {
	reply, err := h.client.FindSpacesNearBuddiesForUser(context.Background(), &hostapipb.FindSpacesNearBuddiesForUserArgs{
		UserId:   userID,
		BuddyIds: buddyIDs,
		Enter:    timeToProto(enter),
		Leave:    timeToProto(leave),
	})
	if err != nil {
		return nil, err
	}
	if reply.Err != "" {
		return nil, strErr(reply.Err)
	}
	return &NearBuddiesResult{
		LocationID:    reply.LocationId,
		LocationName:  reply.LocationName,
		Spaces:        nearBuddySpaceInfosFromProto(reply.Spaces),
		FoundBuddyIDs: reply.FoundBuddyIds,
	}, nil
}

// ─── Host side (gRPC server) ─────────────────────────────────────────────────

func (s *HostAPIGRPCServer) GetBuddiesForUser(ctx context.Context, a *hostapipb.GetBuddiesForUserArgs) (*hostapipb.GetBuddiesForUserReply, error) {
	v, err := s.impl.GetBuddiesForUser(a.UserId)
	return &hostapipb.GetBuddiesForUserReply{Buddies: buddyInfosToProto(v), Err: errStr(err)}, nil
}

func (s *HostAPIGRPCServer) GetBuddyScheduleForUser(ctx context.Context, a *hostapipb.GetBuddyScheduleForUserArgs) (*hostapipb.GetBuddyScheduleForUserReply, error) {
	v, err := s.impl.GetBuddyScheduleForUser(a.UserId, a.BuddyIds, timeFromProto(a.Enter), timeFromProto(a.Leave))
	return &hostapipb.GetBuddyScheduleForUserReply{Schedule: buddyScheduleInfosToProto(v), Err: errStr(err)}, nil
}

func (s *HostAPIGRPCServer) FindSpacesNearBuddiesForUser(ctx context.Context, a *hostapipb.FindSpacesNearBuddiesForUserArgs) (*hostapipb.FindSpacesNearBuddiesForUserReply, error) {
	v, err := s.impl.FindSpacesNearBuddiesForUser(a.UserId, a.BuddyIds, timeFromProto(a.Enter), timeFromProto(a.Leave))
	if err != nil || v == nil {
		return &hostapipb.FindSpacesNearBuddiesForUserReply{Err: errStr(err)}, nil
	}
	return &hostapipb.FindSpacesNearBuddiesForUserReply{
		LocationId:    v.LocationID,
		LocationName:  v.LocationName,
		Spaces:        nearBuddySpaceInfosToProto(v.Spaces),
		FoundBuddyIds: v.FoundBuddyIDs,
	}, nil
}

// ─── Conversions ─────────────────────────────────────────────────────────────

func buddyInfoToProto(b *BuddyInfo) *hostapipb.BuddyInfo {
	if b == nil {
		return nil
	}
	return &hostapipb.BuddyInfo{
		Id:          b.ID,
		BuddyUserId: b.BuddyUserID,
		Email:       b.Email,
		Firstname:   b.Firstname,
		Lastname:    b.Lastname,
	}
}

func buddyInfoFromProto(p *hostapipb.BuddyInfo) *BuddyInfo {
	if p == nil {
		return nil
	}
	return &BuddyInfo{
		ID:          p.Id,
		BuddyUserID: p.BuddyUserId,
		Email:       p.Email,
		Firstname:   p.Firstname,
		Lastname:    p.Lastname,
	}
}

func buddyInfosToProto(in []*BuddyInfo) []*hostapipb.BuddyInfo {
	out := make([]*hostapipb.BuddyInfo, 0, len(in))
	for _, b := range in {
		out = append(out, buddyInfoToProto(b))
	}
	return out
}

func buddyInfosFromProto(in []*hostapipb.BuddyInfo) []*BuddyInfo {
	out := make([]*BuddyInfo, 0, len(in))
	for _, b := range in {
		out = append(out, buddyInfoFromProto(b))
	}
	return out
}

func buddyScheduleInfosToProto(in []*BuddyScheduleInfo) []*hostapipb.BuddyScheduleInfo {
	out := make([]*hostapipb.BuddyScheduleInfo, 0, len(in))
	for _, e := range in {
		out = append(out, &hostapipb.BuddyScheduleInfo{
			Buddy:        buddyInfoToProto(&e.Buddy),
			LocationId:   e.LocationID,
			LocationName: e.LocationName,
			SpaceId:      e.SpaceID,
			SpaceName:    e.SpaceName,
			Enter:        timeToProto(e.Enter),
			Leave:        timeToProto(e.Leave),
		})
	}
	return out
}

func buddyScheduleInfosFromProto(in []*hostapipb.BuddyScheduleInfo) []*BuddyScheduleInfo {
	out := make([]*BuddyScheduleInfo, 0, len(in))
	for _, e := range in {
		info := &BuddyScheduleInfo{
			LocationID:   e.LocationId,
			LocationName: e.LocationName,
			SpaceID:      e.SpaceId,
			SpaceName:    e.SpaceName,
			Enter:        timeFromProto(e.Enter),
			Leave:        timeFromProto(e.Leave),
		}
		if b := buddyInfoFromProto(e.Buddy); b != nil {
			info.Buddy = *b
		}
		out = append(out, info)
	}
	return out
}

func nearBuddySpaceInfosToProto(in []*NearBuddySpaceInfo) []*hostapipb.NearBuddySpaceInfo {
	out := make([]*hostapipb.NearBuddySpaceInfo, 0, len(in))
	for _, s := range in {
		out = append(out, &hostapipb.NearBuddySpaceInfo{
			Space:            spaceToProto(&s.Space),
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
			Distance:         s.Distance,
			NearBuddyIds:     s.NearBuddyIDs,
		})
	}
	return out
}

func nearBuddySpaceInfosFromProto(in []*hostapipb.NearBuddySpaceInfo) []*NearBuddySpaceInfo {
	out := make([]*NearBuddySpaceInfo, 0, len(in))
	for _, s := range in {
		info := &NearBuddySpaceInfo{
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
			Distance:         s.Distance,
			NearBuddyIDs:     s.NearBuddyIds,
		}
		if sp := spaceFromProto(s.Space); sp != nil {
			info.Space = *sp
		}
		out = append(out, info)
	}
	return out
}
