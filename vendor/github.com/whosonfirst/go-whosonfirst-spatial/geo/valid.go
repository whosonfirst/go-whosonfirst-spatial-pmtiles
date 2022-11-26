package geo

func IsValidMinLatitude(lat float64) bool {
	return lat >= -90.00
}

func IsValidMaxLatitude(lat float64) bool {
	return lat <= 90.00
}

func IsValidLatitude(lat float64) bool {
	return IsValidMinLatitude(lat) && IsValidMaxLatitude(lat)
}

func IsValidMinLongitude(lon float64) bool {
	return lon >= -180.00
}

func IsValidMaxLongitude(lon float64) bool {
	return lon <= 180.00
}

func IsValidLongitude(lon float64) bool {
	return IsValidMinLongitude(lon) && IsValidMaxLongitude(lon)
}
