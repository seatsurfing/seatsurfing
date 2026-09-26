package router

import (
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
)

// SearchAttribute filters locations or spaces by attribute value in REST
// requests. See service.SearchAttribute.
type SearchAttribute = service.SearchAttribute

// MatchesSearchAttributes reports whether entityID matches all attributes.
// See service.MatchesSearchAttributes.
func MatchesSearchAttributes(entityID string, m *[]SearchAttribute, attributeValues []*SpaceAttributeValue) bool {
	return service.MatchesSearchAttributes(entityID, m, attributeValues)
}
