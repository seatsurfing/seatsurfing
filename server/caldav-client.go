package main

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/google/uuid"
)

type CalDAVClient struct {
	url        string
	httpClient webdav.HTTPClient
	client     *caldav.Client
	principal  string
	homeSet    string
}

type CalDAVCalendar struct {
	Path string
	Name string
}

type CalDAVEvent struct {
	ID       string
	Title    string
	Start    time.Time
	End      time.Time
	Location string
}

func (c *CalDAVClient) Connect(url, username, password string) error {
	httpClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, username, password)
	caldavClient, err := caldav.NewClient(httpClient, url)
	if err != nil {
		return err
	}
	principal, err := caldavClient.FindCurrentUserPrincipal(context.Background())
	if err != nil {
		return err
	}
	homeSet, err := caldavClient.FindCalendarHomeSet(context.Background(), principal)
	if err != nil {
		return err
	}
	c.url = url
	c.client = caldavClient
	c.httpClient = httpClient
	c.principal = principal
	c.homeSet = homeSet
	return nil
}

func (c *CalDAVClient) ListCalendars() ([]*CalDAVCalendar, error) {
	calendars, err := c.client.FindCalendars(context.Background(), c.homeSet)
	if err != nil {
		return nil, err
	}
	res := make([]*CalDAVCalendar, 0)
	for _, calendar := range calendars {
		res = append(res, &CalDAVCalendar{
			Path: calendar.Path,
			Name: calendar.Name,
		})
	}
	return res, nil
}

func (c *CalDAVClient) CreateEvent(calendarPath string, e *CalDAVEvent) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	cal := c.getCaldavEvent(e)

	_, err := c.client.PutCalendarObject(context.Background(), path.Join(calendarPath, e.ID+".ics"), cal)
	return err
}

func (c *CalDAVClient) DeleteEvent(calendarPath string, e *CalDAVEvent) error {
	cal := c.getCaldavEvent(e)

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		return err
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return err
	}
	pathNew := url.URL{
		Scheme: u.Scheme,
		User:   u.User,
		Host:   u.Host,
		Path:   path.Join(calendarPath, e.ID+".ics"),
	}

	req, err := http.NewRequest(http.MethodDelete, pathNew.String(), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", ical.MIMEType)

	resp, err := c.httpClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (c *CalDAVClient) getCaldavEvent(e *CalDAVEvent) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//seatsurfing.app//seatsurfing//EN")
	cal.Props.SetText(ical.PropVersion, "2.0")

	event := ical.NewEvent()
	event.Props.SetText(ical.PropSummary, e.Title)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now())
	event.Props.SetDateTime(ical.PropDateTimeStart, e.Start)
	event.Props.SetDateTime(ical.PropDateTimeEnd, e.End)
	event.Props.SetText(ical.PropLocation, e.Location)
	event.Props.Del(ical.PropDuration)
	event.Props.SetText(ical.PropUID, e.ID)
	cal.Children = append(cal.Children, event.Component)

	return cal
}
