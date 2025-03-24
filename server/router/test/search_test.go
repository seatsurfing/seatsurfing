package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationsMatchesSearchAttributesSuccess(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
		{AttributeID: "2", Comparator: "neq", Value: "value2"},
		{AttributeID: "3", Comparator: "contains", Value: "value3"},
		{AttributeID: "4", Comparator: "ncontains", Value: "value4"},
		{AttributeID: "5", Comparator: "lt", Value: "5"},
		{AttributeID: "6", Comparator: "gt", Value: "5"},
		{AttributeID: "7", Comparator: "contains", Value: "foo"},
		{AttributeID: "7", Comparator: "contains", Value: "bar"},
		{AttributeID: "7", Comparator: "contains", Value: "*"},
		{AttributeID: "7", Comparator: "ncontains", Value: "test2"},
		{AttributeID: "8", Comparator: "ncontains", Value: "*"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
		{AttributeID: "2", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value2.2"},
		{AttributeID: "3", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value3-test"},
		{AttributeID: "4", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-valuefour-test"},
		{AttributeID: "5", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
		{AttributeID: "6", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "7"},
		{AttributeID: "7", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: `["foo", "bar", "test"]`},
		{AttributeID: "8", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: `[]`},
	}
	CheckTestBool(t, true, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesSuccess2(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
		{AttributeID: "2", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value2.2"},
		{AttributeID: "3", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value3-test"},
		{AttributeID: "4", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-valuefour-test"},
		{AttributeID: "5", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
		{AttributeID: "6", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "7"},
	}
	CheckTestBool(t, true, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesMultipleEntities(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "2", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1111"},
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	CheckTestBool(t, true, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesMissingAttribute(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
		{AttributeID: "2", Comparator: "neq", Value: "value2"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesEqWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value11"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesNeqWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "neq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesContainsWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "contains", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value2-test"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesNcontainsWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "ncontains", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value1-test"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesLtWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "lt", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "5"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesGtWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "gt", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "5"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesGteWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "gte", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesLteWrong(t *testing.T) {
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "lte", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "6"},
	}
	CheckTestBool(t, false, MatchesSearchAttributes("1", &searchAttributes, attributeValues))
}
