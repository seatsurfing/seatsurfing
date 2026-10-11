package router

import (
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
)

type BookingAttributeRouter struct {
}

type CreateBookingAttributeRequest struct {
	Label    string `json:"label" validate:"required,max=256"`
	Type     int    `json:"type" validate:"oneof=1 2 3 4"`
	Required bool   `json:"required"`
}

type MoveBookingAttributeRequest struct {
	Direction string `json:"direction" validate:"required,oneof=up down"`
}

type GetBookingAttributeResponse struct {
	ID string `json:"id"`
	CreateBookingAttributeRequest
}

type GetBookingAttributeValueResponse struct {
	AttributeID string `json:"attributeId"`
	Label       string `json:"label"`
	Type        int    `json:"type"`
	Value       string `json:"value"`
}

func (router *BookingAttributeRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}/move", router.move).Methods("PUT")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *BookingAttributeRouter) getAll(w http.ResponseWriter, r *http.Request) {
	list, bErr := service.GetBookingAttributeService().List(GetRequestUser(r))
	if bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	res := []*GetBookingAttributeResponse{}
	for _, e := range list {
		res = append(res, router.copyToRestModel(e))
	}
	SendJSON(w, res)
}

func (router *BookingAttributeRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateBookingAttributeRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	e, bErr := service.GetBookingAttributeService().Create(GetRequestUser(r), router.copyFromRestModel(&m))
	if bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	SendCreated(w, e.ID)
}

func (router *BookingAttributeRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateBookingAttributeRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if bErr := service.GetBookingAttributeService().Update(GetRequestUser(r), mux.Vars(r)["id"], router.copyFromRestModel(&m)); bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	SendUpdated(w)
}

func (router *BookingAttributeRouter) move(w http.ResponseWriter, r *http.Request) {
	var m MoveBookingAttributeRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if bErr := service.GetBookingAttributeService().Move(GetRequestUser(r), mux.Vars(r)["id"], m.Direction == "up"); bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	SendUpdated(w)
}

func (router *BookingAttributeRouter) delete(w http.ResponseWriter, r *http.Request) {
	if bErr := service.GetBookingAttributeService().Delete(GetRequestUser(r), mux.Vars(r)["id"]); bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	SendUpdated(w)
}

func (router *BookingAttributeRouter) copyFromRestModel(m *CreateBookingAttributeRequest) *service.BookingAttributeInput {
	return &service.BookingAttributeInput{
		Label:    m.Label,
		Type:     BookingAttributeType(m.Type),
		Required: m.Required,
	}
}

func (router *BookingAttributeRouter) copyToRestModel(e *BookingAttribute) *GetBookingAttributeResponse {
	m := &GetBookingAttributeResponse{}
	m.ID = e.ID
	m.Label = e.Label
	m.Type = int(e.Type)
	m.Required = e.Required
	return m
}
