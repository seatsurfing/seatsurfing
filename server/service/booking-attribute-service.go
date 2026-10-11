package service

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// BookingAttributeService manages the organization-defined attributes that
// are captured with a booking. They currently apply to public bookings only.
type BookingAttributeService struct {
}

var bookingAttributeService *BookingAttributeService
var bookingAttributeServiceOnce sync.Once

func GetBookingAttributeService() *BookingAttributeService {
	bookingAttributeServiceOnce.Do(func() {
		bookingAttributeService = &BookingAttributeService{}
	})
	return bookingAttributeService
}

const (
	MaxBookingAttributes                   = 3
	MaxBookingAttributeLabelLength         = 256
	MaxBookingAttributeTextLength          = 256
	MaxBookingAttributeMultilineTextLength = 4096
	MaxBookingAttributeNumberLength        = 32
)

// Booking attribute error codes, sent by the REST API as X-Error-Code.
const (
	BookingCodeInvalidAttributes     = 1014
	BookingCodeAttributeLimitReached = 1015
)

var bookingAttributeNumberPattern = regexp.MustCompile(`^0*[1-9][0-9]*$`)

const (
	BookingAttributeCheckboxChecked   = "1"
	BookingAttributeCheckboxUnchecked = "0"
)

type BookingAttributeInput struct {
	Label    string
	Type     BookingAttributeType
	Required bool
}

func (s *BookingAttributeService) canManage(user *User) bool {
	return HasPermission(user, user.OrganizationID, PermissionOrgSettings, PermissionLevelAdmin)
}

func (s *BookingAttributeService) validateInput(in *BookingAttributeInput) (*BookingAttributeInput, *BookingError) {
	label := strings.TrimSpace(in.Label)
	if label == "" || utf8.RuneCountInString(label) > MaxBookingAttributeLabelLength {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	if in.Type != BookingAttributeTypeText && in.Type != BookingAttributeTypeTextMultiline && in.Type != BookingAttributeTypeNumber && in.Type != BookingAttributeTypeCheckbox {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	return &BookingAttributeInput{Label: label, Type: in.Type, Required: in.Required}, nil
}

func (s *BookingAttributeService) getManageable(user *User, id string) (*BookingAttribute, *BookingError) {
	if !s.canManage(user) {
		return nil, &BookingError{Kind: BookingErrorForbidden}
	}
	e, err := GetBookingAttributeRepository().GetOne(id)
	if err != nil || !CanAccessOrg(user, e.OrganizationID) {
		return nil, &BookingError{Kind: BookingErrorNotFound}
	}
	return e, nil
}

func (s *BookingAttributeService) List(user *User) ([]*BookingAttribute, *BookingError) {
	if !s.canManage(user) {
		return nil, &BookingError{Kind: BookingErrorForbidden}
	}
	list, err := GetBookingAttributeRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		return nil, &BookingError{Kind: BookingErrorInternal}
	}
	return list, nil
}

func (s *BookingAttributeService) Create(user *User, in *BookingAttributeInput) (*BookingAttribute, *BookingError) {
	if !s.canManage(user) {
		return nil, &BookingError{Kind: BookingErrorForbidden}
	}
	in, bErr := s.validateInput(in)
	if bErr != nil {
		return nil, bErr
	}
	e := &BookingAttribute{
		OrganizationID: user.OrganizationID,
		Label:          in.Label,
		Type:           in.Type,
		Required:       in.Required,
	}
	created, err := GetBookingAttributeRepository().CreateIfBelowLimit(e, MaxBookingAttributes)
	if err != nil {
		log.Println(err)
		return nil, &BookingError{Kind: BookingErrorInternal}
	}
	if !created {
		return nil, &BookingError{Kind: BookingErrorInvalid, Code: BookingCodeAttributeLimitReached}
	}
	return e, nil
}

func (s *BookingAttributeService) Update(user *User, id string, in *BookingAttributeInput) *BookingError {
	e, bErr := s.getManageable(user, id)
	if bErr != nil {
		return bErr
	}
	in, bErr = s.validateInput(in)
	if bErr != nil {
		return bErr
	}
	e.Label = in.Label
	e.Type = in.Type
	e.Required = in.Required
	if err := GetBookingAttributeRepository().Update(e); err != nil {
		log.Println(err)
		return &BookingError{Kind: BookingErrorInternal}
	}
	return nil
}

func (s *BookingAttributeService) Delete(user *User, id string) *BookingError {
	e, bErr := s.getManageable(user, id)
	if bErr != nil {
		return bErr
	}
	if err := GetBookingAttributeRepository().Delete(e); err != nil {
		log.Println(err)
		return &BookingError{Kind: BookingErrorInternal}
	}
	return nil
}

// GetForPublicBooking returns the attributes a public booker has to fill in.
func (s *BookingAttributeService) GetForPublicBooking(organizationID string) ([]*BookingAttribute, error) {
	return GetBookingAttributeRepository().GetAll(organizationID)
}

// ValidateValues checks the submitted values (attribute ID to value) against
// the organization's attributes and returns them normalized. Empty optional
// values are dropped, unchecked checkboxes are stored as "0".
func (s *BookingAttributeService) ValidateValues(organizationID string, values map[string]string) (map[string]string, *BookingError) {
	attributes, err := GetBookingAttributeRepository().GetAll(organizationID)
	if err != nil {
		log.Println(err)
		return nil, &BookingError{Kind: BookingErrorInternal}
	}
	invalid := &BookingError{Kind: BookingErrorInvalid, Code: BookingCodeInvalidAttributes}
	known := make(map[string]*BookingAttribute, len(attributes))
	for _, a := range attributes {
		known[a.ID] = a
	}
	for id := range values {
		if _, ok := known[id]; !ok {
			return nil, invalid
		}
	}
	res := map[string]string{}
	for _, a := range attributes {
		value, ok := normalizeBookingAttributeValue(a.Type, values[a.ID])
		if !ok {
			return nil, invalid
		}
		missing := value == "" || (a.Type == BookingAttributeTypeCheckbox && value == BookingAttributeCheckboxUnchecked)
		if missing && a.Required {
			return nil, invalid
		}
		if value == "" {
			continue
		}
		res[a.ID] = value
	}
	return res, nil
}

func normalizeBookingAttributeValue(t BookingAttributeType, value string) (string, bool) {
	if !utf8.ValidString(value) {
		return "", false
	}
	switch t {
	case BookingAttributeTypeText:
		value = strings.TrimSpace(value)
		if strings.ContainsAny(value, "\r\n") || utf8.RuneCountInString(value) > MaxBookingAttributeTextLength {
			return "", false
		}
	case BookingAttributeTypeTextMultiline:
		value = strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
		if utf8.RuneCountInString(value) > MaxBookingAttributeMultilineTextLength {
			return "", false
		}
	case BookingAttributeTypeNumber:
		value = strings.TrimSpace(value)
		if value == "" {
			return "", true
		}
		if len(value) > MaxBookingAttributeNumberLength {
			return "", false
		}
		if !bookingAttributeNumberPattern.MatchString(value) {
			return "", false
		}
	case BookingAttributeTypeCheckbox:
		switch strings.TrimSpace(value) {
		case BookingAttributeCheckboxChecked, "true":
			value = BookingAttributeCheckboxChecked
		case "", BookingAttributeCheckboxUnchecked, "false":
			value = BookingAttributeCheckboxUnchecked
		default:
			return "", false
		}
	default:
		return "", false
	}
	return value, true
}

// SetValues stores already validated values for a booking.
func (s *BookingAttributeService) SetValues(bookingID string, values map[string]string) error {
	for attributeID, value := range values {
		if err := GetBookingAttributeValueRepository().Set(bookingID, attributeID, value); err != nil {
			return err
		}
	}
	return nil
}

// GetValuesForBookings returns the stored values grouped by booking ID.
func (s *BookingAttributeService) GetValuesForBookings(bookingIDs []string) (map[string][]*BookingAttributeValue, error) {
	list, err := GetBookingAttributeValueRepository().GetAllForBookingList(bookingIDs)
	if err != nil {
		return nil, err
	}
	res := map[string][]*BookingAttributeValue{}
	for _, v := range list {
		res[v.BookingID] = append(res[v.BookingID], v)
	}
	return res, nil
}
