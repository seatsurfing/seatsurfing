package router

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type KioskRouter struct {
}

type KioskBookingResponse struct {
	ID           string    `json:"id"`
	Subject      string    `json:"subject"`
	Owner        string    `json:"owner"`
	OwnerVisible bool      `json:"ownerVisible"`
	Enter        time.Time `json:"enter"`
	Leave        time.Time `json:"leave"`
}

type KioskResponse struct {
	SpaceID        string                `json:"spaceId"`
	SpaceName      string                `json:"spaceName"`
	LocationID     string                `json:"locationId"`
	LocationName   string                `json:"locationName"`
	Timezone       string                `json:"timezone"`
	Status         string                `json:"status"`
	CurrentBooking *KioskBookingResponse `json:"currentBooking"`
	NextBooking    *KioskBookingResponse `json:"nextBooking"`
	RefreshedAt    time.Time             `json:"refreshedAt"`
}

func (router *KioskRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}/status", router.getKiosk).Methods("GET")
}

func (router *KioskRouter) getKiosk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceID := vars["id"]

	// Extract kiosk secret from Authorization: Bearer <secret> header
	secret := ""
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		secret = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Look up space
	space, err := GetSpaceRepository().GetOne(spaceID)
	if err != nil || space == nil {
		SendNotFound(w)
		return
	}

	// Kiosk mode must be enabled for this space
	if !space.KioskEnabled {
		SendNotFound(w)
		return
	}

	// Get location and organization context
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil || location == nil {
		SendNotFound(w)
		return
	}

	// Validate kiosk secret
	if secret == "" {
		SendUnauthorized(w)
		return
	}
	storedHash, err := GetSettingsRepository().Get(location.OrganizationID, SettingKioskSecret.Name)
	if err != nil || storedHash == "" {
		SendUnauthorized(w)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(secret)) != nil {
		SendUnauthorized(w)
		return
	}

	// Determine timezone
	tz := GetLocationRepository().GetTimezone(location)
	tzLocation, err := time.LoadLocation(tz)
	if err != nil || tzLocation == nil {
		log.Println("kiosk: error loading timezone:", tz, err)
		tzLocation = time.UTC
	}
	now := time.Now().In(tzLocation)

	// Determine name visibility
	showNames, _ := GetSettingsRepository().GetBool(location.OrganizationID, SettingShowNames.Name)

	// Fetch current and next bookings
	current, next, _ := GetBookingRepository().GetCurrentAndNextBySpaceID(spaceID, now)

	res := &KioskResponse{
		SpaceID:        space.ID,
		SpaceName:      space.Name,
		LocationID:     location.ID,
		LocationName:   location.Name,
		Timezone:       tz,
		CurrentBooking: nil,
		NextBooking:    nil,
		RefreshedAt:    now,
	}

	if current != nil {
		res.Status = "occupied"
		res.CurrentBooking = router.toKioskBooking(current, showNames, tzLocation)
	} else {
		res.Status = "available"
	}
	if next != nil {
		res.NextBooking = router.toKioskBooking(next, showNames, tzLocation)
	}

	SendJSON(w, res)
}

func (router *KioskRouter) toKioskBooking(b *KioskBookingEntry, showNames bool, loc *time.Location) *KioskBookingResponse {
	owner := ""
	ownerVisible := false
	if showNames {
		owner = b.UserEmail
		ownerVisible = true
	}
	return &KioskBookingResponse{
		ID:           b.ID,
		Subject:      b.Subject,
		Owner:        owner,
		OwnerVisible: ownerVisible,
		Enter:        b.Enter.In(loc),
		Leave:        b.Leave.In(loc),
	}
}
