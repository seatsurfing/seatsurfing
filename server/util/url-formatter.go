package util

import (
	"strconv"
	"strings"

	"github.com/seatsurfing/seatsurfing/server/config"
)

func FormatURL(domain string) string {
	scheme := config.GetConfig().PublicScheme
	port := config.GetConfig().PublicPort
	if scheme == "https" && port == 443 {
		port = 0
	}
	if scheme == "http" && port == 80 {
		port = 0
	}
	var sb strings.Builder
	sb.WriteString(scheme)
	sb.WriteString("://")
	sb.WriteString(domain)
	if port != 0 {
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(port))
	}
	return sb.String()
}
