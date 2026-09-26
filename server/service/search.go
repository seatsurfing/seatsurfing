package service

import (
	"encoding/json"
	"errors"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// Synthetic location attributes computed at search time.
const (
	SearchAttributeNumSpaces     string = "numSpaces"
	SearchAttributeNumFreeSpaces string = "numFreeSpaces"
	SearchAttributeBuddyOnSite   string = "buddyOnSite"
)

// SearchComparators are the comparators SearchAttribute supports.
var SearchComparators = []string{"eq", "neq", "contains", "ncontains", "gt", "gte", "lt", "lte"}

// SearchAttribute filters locations or spaces by attribute value. The
// validate tags are applied to REST request bodies; ValidateSearchAttributes
// applies the same rules for callers without a request body.
type SearchAttribute struct {
	AttributeID string `json:"attributeId" validate:"omitempty,uuid|oneof=numSpaces numFreeSpaces buddyOnSite"`
	Comparator  string `json:"comparator" validate:"omitempty,oneof=eq neq contains ncontains gt gte lt lte"`
	Value       string `json:"value" validate:"max=256"`
}

func MatchesSearchAttributes(entityID string, m *[]SearchAttribute, attributeValues []*SpaceAttributeValue) bool {
	var matchString = func(a, b, comparator string) bool {
		if comparator == "eq" {
			return a == b
		} else if comparator == "neq" {
			return a != b
		} else if comparator == "contains" {
			return strings.Contains(a, b)
		} else if comparator == "ncontains" {
			return !strings.Contains(a, b)
		} else if comparator == "gt" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt > attrValInt
		} else if comparator == "lt" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt < attrValInt
		} else if comparator == "gte" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt >= attrValInt
		} else if comparator == "lte" {
			searchAttrInt, err := strconv.Atoi(a)
			if err != nil {
				return false
			}
			attrValInt, err := strconv.Atoi(b)
			if err != nil {
				return false
			}
			return searchAttrInt <= attrValInt
		}
		return false
	}

	var matchArray = func(a []string, b, comparator string) bool {
		if comparator == "contains" {
			if b == "*" {
				return len(a) > 0
			}
			return slices.Contains(a, b)
		} else if comparator == "ncontains" {
			if b == "*" {
				return len(a) == 0
			}
			return !slices.Contains(a, b)
		}
		return false
	}

	for _, searchAttr := range *m {
		found := false
		for _, attrVal := range attributeValues {
			if (attrVal.AttributeID == searchAttr.AttributeID) && (attrVal.EntityID == entityID) {
				if strings.Index(attrVal.Value, "[") == 0 && strings.Index(attrVal.Value, "]") == len(attrVal.Value)-1 {
					var arr []string
					if err := json.Unmarshal([]byte(attrVal.Value), &arr); err != nil {
						log.Println(err)
						return false
					}
					found = matchArray(arr, searchAttr.Value, searchAttr.Comparator)
				} else {
					found = matchString(attrVal.Value, searchAttr.Value, searchAttr.Comparator)
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ValidateSearchAttributes checks attributes with the same rules as the
// validate tags of SearchAttribute.
func ValidateSearchAttributes(attributes []SearchAttribute) error {
	for _, a := range attributes {
		if a.AttributeID != "" && a.AttributeID != SearchAttributeNumSpaces && a.AttributeID != SearchAttributeNumFreeSpaces && a.AttributeID != SearchAttributeBuddyOnSite {
			if _, err := uuid.Parse(a.AttributeID); err != nil {
				return errors.New("invalid attribute ID " + a.AttributeID)
			}
		}
		if a.Comparator != "" && !slices.Contains(SearchComparators, a.Comparator) {
			return errors.New("invalid comparator " + a.Comparator)
		}
		if len(a.Value) > 256 {
			return errors.New("attribute value too long")
		}
	}
	return nil
}
