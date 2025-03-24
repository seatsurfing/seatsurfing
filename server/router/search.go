package router

import (
	"encoding/json"
	"log"
	"slices"
	"strconv"
	"strings"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SearchAttribute struct {
	AttributeID string `json:"attributeId"`
	Comparator  string `json:"comparator"`
	Value       string `json:"value"`
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
