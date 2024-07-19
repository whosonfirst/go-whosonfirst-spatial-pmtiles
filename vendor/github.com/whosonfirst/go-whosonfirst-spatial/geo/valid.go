package geo

// Returns a boolean value indicating where 'lat' is greater than or equal to -90
func IsValidMinLatitude(lat float64) bool {
	return lat >= -90.00
}

// Returns a boolean value indicating where 'lat' is less than or equal to 90
func IsValidMaxLatitude(lat float64) bool {
	return lat <= 90.00
}

// Returns a boolean value indicating where 'lat' is greater than or equal to -90 and less than or equal to 90
func IsValidLatitude(lat float64) bool {
	return IsValidMinLatitude(lat) && IsValidMaxLatitude(lat)
}

// Returns a boolean value indicating where 'lon' is greater than or equal to -180
func IsValidMinLongitude(lon float64) bool {
	return lon >= -180.00
}

// Returns a boolean value indicating where 'lon' is less than or equal to 180
func IsValidMaxLongitude(lon float64) bool {
	return lon <= 180.00
}

// Returns a boolean value indicating where 'lat' is greater than or equal to -180 and less than or equal to 180
func IsValidLongitude(lon float64) bool {
	return IsValidMinLongitude(lon) && IsValidMaxLongitude(lon)
}
