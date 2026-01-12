package util

var CountriesEU = map[string]string{
	"AT": "Austria",
	"BE": "Belgium",
	"BG": "Bulgaria",
	"HR": "Croatia",
	"CY": "Cyprus",
	"CZ": "Czechia",
	"DK": "Denmark",
	"EE": "Estonia",
	"FI": "Finland",
	"FR": "France",
	"DE": "Germany",
	"GR": "Greece",
	"HU": "Hungary",
	"IE": "Ireland",
	"IT": "Italy",
	"LV": "Latvia",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"MT": "Malta",
	"NL": "Netherlands",
	"PL": "Poland",
	"PT": "Portugal",
	"RO": "Romania",
	"SK": "Slovakia",
	"SI": "Slovenia",
	"ES": "Spain",
	"SE": "Sweden",
}

var CountriesNonEUEurope = map[string]string{
	"AL": "Albania",
	"AD": "Andorra",
	"BY": "Belarus",
	"BA": "Bosnia and Herzegovina",
	"IS": "Iceland",
	"LI": "Liechtenstein",
	"MD": "Moldova",
	"MC": "Monaco",
	"ME": "Montenegro",
	"MK": "North Macedonia",
	"NO": "Norway",
	"RU": "Russia",
	"SM": "San Marino",
	"RS": "Serbia",
	"CH": "Switzerland",
	"TR": "Türkiye",
	"UA": "Ukraine",
	"GB": "United Kingdom",
	"VA": "Vatican City",
}

var CountriesAfrica = map[string]string{
	"DZ": "Algeria", "AO": "Angola", "BJ": "Benin", "BW": "Botswana", "BF": "Burkina Faso",
	"BI": "Burundi", "CV": "Cabo Verde", "CM": "Cameroon", "CF": "Central African Republic",
	"TD": "Chad", "KM": "Comoros", "CG": "Congo", "CD": "DR Congo", "CI": "Ivory Coast",
	"DJ": "Djibouti", "EG": "Egypt", "GQ": "Equatorial Guinea", "ER": "Eritrea",
	"SZ": "Eswatini", "ET": "Ethiopia", "GA": "Gabon", "GM": "Gambia", "GH": "Ghana",
	"GN": "Guinea", "GW": "Guinea-Bissau", "KE": "Kenya", "LS": "Lesotho", "LR": "Liberia",
	"LY": "Libya", "MG": "Madagascar", "MW": "Malawi", "ML": "Mali", "MR": "Mauritania",
	"MU": "Mauritius", "MA": "Morocco", "MZ": "Mozambique", "NA": "Namibia", "NE": "Niger",
	"NG": "Nigeria", "RW": "Rwanda", "ST": "Sao Tome and Principe", "SN": "Senegal",
	"SC": "Seychelles", "SL": "Sierra Leone", "SO": "Somalia", "ZA": "South Africa",
	"SS": "South Sudan", "SD": "Sudan", "TZ": "Tanzania", "TG": "Togo", "TN": "Tunisia",
	"UG": "Uganda", "ZM": "Zambia", "ZW": "Zimbabwe",
}

var CountriesAsia = map[string]string{
	"AF": "Afghanistan", "AM": "Armenia", "AZ": "Azerbaijan", "BH": "Bahrain", "BD": "Bangladesh",
	"BT": "Bhutan", "BN": "Brunei", "KH": "Cambodia", "CN": "China", "GE": "Georgia",
	"IN": "India", "ID": "Indonesia", "IR": "Iran", "IQ": "Iraq", "IL": "Israel",
	"JP": "Japan", "JO": "Jordan", "KZ": "Kazakhstan", "KP": "North Korea", "KR": "South Korea",
	"KW": "Kuwait", "KG": "Kyrgyzstan", "LA": "Laos", "LB": "Lebanon", "MY": "Malaysia",
	"MV": "Maldives", "MN": "Mongolia", "MM": "Myanmar", "NP": "Nepal", "OM": "Oman",
	"PK": "Pakistan", "PH": "Philippines", "QA": "Qatar", "SA": "Saudi Arabia", "SG": "Singapore",
	"LK": "Sri Lanka", "SY": "Syria", "TJ": "Tajikistan", "TH": "Thailand", "TL": "Timor-Leste",
	"TM": "Turkmenistan", "AE": "UAE", "UZ": "Uzbekistan", "VN": "Vietnam", "YE": "Yemen",
}

var CountriesNorthAmerica = map[string]string{
	"AG": "Antigua and Barbuda", "BS": "Bahamas", "BB": "Barbados", "BZ": "Belize",
	"CA": "Canada", "CR": "Costa Rica", "CU": "Cuba", "DM": "Dominica",
	"DO": "Dominican Republic", "SV": "El Salvador", "GD": "Grenada", "GT": "Guatemala",
	"HT": "Haiti", "HN": "Honduras", "JM": "Jamaica", "MX": "Mexico", "NI": "Nicaragua",
	"PA": "Panama", "KN": "Saint Kitts and Nevis", "LC": "Saint Lucia", "VC": "Saint Vincent",
	"TT": "Trinidad and Tobago", "US": "United States",
}

var CountriesSouthAmerica = map[string]string{
	"AR": "Argentina", "BO": "Bolivia", "BR": "Brazil", "CL": "Chile", "CO": "Colombia",
	"EC": "Ecuador", "GY": "Guyana", "PY": "Paraguay", "PE": "Peru", "SR": "Suriname",
	"UY": "Uruguay", "VE": "Venezuela",
}

var CountriesOceania = map[string]string{
	"AU": "Australia", "FJ": "Fiji", "KI": "Kiribati", "MH": "Marshall Islands",
	"FM": "Micronesia", "NR": "Nauru", "NZ": "New Zealand", "PW": "Palau",
	"PG": "Papua New Guinea", "WS": "Samoa", "SB": "Solomon Islands", "TO": "Tonga",
	"TV": "Tuvalu", "VU": "Vanuatu",
}

var CountriesByRegion = map[string]map[string]string{
	"EU":           CountriesEU,
	"EuropeNonEU":  CountriesNonEUEurope,
	"Africa":       CountriesAfrica,
	"Asia":         CountriesAsia,
	"NorthAmerica": CountriesNorthAmerica,
	"SouthAmerica": CountriesSouthAmerica,
	"Oceania":      CountriesOceania,
}

func IsValidCountry(code string) bool {
	for _, countries := range CountriesByRegion {
		if _, exists := countries[code]; exists {
			return true
		}
	}
	return false
}

func IsValidCountryInRegion(region, code string) bool {
	countries, exists := CountriesByRegion[region]
	if !exists {
		return false
	}
	_, exists = countries[code]
	return exists
}
