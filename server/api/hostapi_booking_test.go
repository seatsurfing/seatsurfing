package api

import (
	"reflect"
	"testing"
)

func TestSearchAttributeFiltersRoundTrip(t *testing.T) {
	in := []SearchAttributeFilter{
		{AttributeID: "a1", Comparator: "eq", Value: "1"},
		{AttributeID: "numFreeSpaces", Comparator: "gte", Value: "3"},
	}
	got := searchAttributesFromProto(searchAttributesToProto(in))
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, in)
	}
}

func TestSpaceAttributeDefinitionsRoundTrip(t *testing.T) {
	in := []*SpaceAttributeDefinition{
		{ID: "a1", OrganizationID: "o1", Label: "Monitor", Type: 2, SpaceApplicable: true},
		{ID: "a2", OrganizationID: "o1", Label: "Floor", Type: 1, LocationApplicable: true},
	}
	got := spaceAttributesFromProto(spaceAttributesToProto(in))
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, in)
	}
}

func TestLocationInfosRoundTrip(t *testing.T) {
	in := []*LocationInfo{
		{
			Location:     Location{ID: "l1", OrganizationID: "o1", Name: "HQ", Timezone: "Europe/Berlin", Enabled: true},
			Attributes:   []AttributeValue{{AttributeID: "a2", Value: "3"}},
			Allowed:      true,
			BookableDays: []int{1, 2, 3, 4, 5},
		},
		{
			Location:     Location{ID: "l2", OrganizationID: "o1", Name: "Branch"},
			Attributes:   []AttributeValue{},
			BookableDays: []int{},
		},
	}
	got := locationInfosFromProto(locationInfosToProto(in))
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, in)
	}
}

func TestSpaceAvailabilityInfosRoundTrip(t *testing.T) {
	in := []*SpaceAvailabilityInfo{
		{
			Space:            Space{ID: "s1", LocationID: "l1", Name: "Desk 1", Enabled: true, RequireSubject: true},
			Attributes:       []AttributeValue{{AttributeID: "a1", Value: "1"}},
			Available:        true,
			Allowed:          true,
			ApprovalRequired: true,
		},
	}
	got := spaceAvailabilityInfosFromProto(spaceAvailabilityInfosToProto(in))
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, in)
	}
}

func TestLocationInfosDropInvalidWeekdays(t *testing.T) {
	in := []*LocationInfo{{BookableDays: []int{-1, 0, 6, 7, 1 << 40}}}
	got := locationInfosFromProto(locationInfosToProto(in))
	if !reflect.DeepEqual(got[0].BookableDays, []int{0, 6}) {
		t.Fatalf("unexpected weekdays: %v", got[0].BookableDays)
	}
}
