package api

import "github.com/gorilla/mux"

type Route interface {
	SetupRoutes(s *mux.Router)
}
