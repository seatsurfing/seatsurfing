package router

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type EsignRouter struct{}

func (r *EsignRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/esign/space/{id}/status", r.getSpaceStatus).Methods("GET")
}

func (r *EsignRouter) getSpaceStatus(w http.ResponseWriter, req *http.Request) {
	// log.Println("EsignRouter: Received GET /space/{id}/status")

	username, password, ok := req.BasicAuth()
	// log.Printf("Auth attempt: ok=%v user=%s\n", ok, username)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := GetUserRepository().GetUsersWithEmail(username)
	if err != nil || len(users) == 0 {
		log.Printf("Login failed: user not found for %s\n", username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user := users[0]

	pwValue, err := user.HashedPassword.Value()
	if err != nil {
		log.Printf("Login failed: invalid stored password for %s\n", username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pwStr, ok := pwValue.(string)
	if !ok || pwStr == "" {
		log.Printf("Login failed: stored password not a string for %s\n", username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !GetUserRepository().CheckPassword(pwStr, password) {
		log.Printf("Login failed: bad password for %s\n", username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// log.Printf("Login successful for %s\n", username)

	vars := mux.Vars(req)
	spaceID := vars["id"]
	// log.Printf("Returning status for spaceID: %s", spaceID)

	space, err := GetSpaceRepository().GetOne(spaceID)
	if err != nil {
		http.Error(w, "Space not found", http.StatusNotFound)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		http.Error(w, "Location not found", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	start := now.Add(-1 * time.Minute)
	end := now.Add(1 * time.Minute)

	list, err := GetSpaceRepository().GetAllInTime(location.ID, start, end)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	userEmail := ""

	for _, e := range list {
		if e.Space.ID == space.ID {
			for _, b := range e.Bookings {
				enter, _ := GetLocationRepository().AttachTimezoneInformation(b.Enter, location)
				leave, _ := GetLocationRepository().AttachTimezoneInformation(b.Leave, location)
				if now.After(enter) && now.Before(leave) {
					userEmail = b.UserEmail
					break
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"userEmail": userEmail,
	})
}
