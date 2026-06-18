package util

type InstallStats struct {
	NumLocations int `json:"numLocations" validate:"min=0"`
	NumUsers     int `json:"numUsers" validate:"min=0"`
	NumBookings  int `json:"numBookings" validate:"min=0"`
	NumSpaces    int `json:"numSpaces" validate:"min=0"`
}

var installStats *InstallStats = &InstallStats{}

func GetInstallStats() *InstallStats {
	return installStats
}

func SetInstallStats(stats *InstallStats) {
	installStats = stats
}
