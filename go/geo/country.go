package geo

import "strings"

type Country string

func ParseCountry(c string) Country {
	formattedInput := strings.TrimSpace(strings.ToUpper(c))
	_, ok := details[Country(formattedInput)]
	if ok {
		return Country(formattedInput)
	}

	for ctry, d := range details {
		if strings.EqualFold(d.Numeric, formattedInput) {
			return ctry
		}
	}

	// default to US
	return US
}

func (c Country) Numeric() (string, error) {
	d, ok := details[c]
	if !ok {
		return "", ErrCountryNotFound
	}

	return d.Numeric, nil
}

func (c Country) Valid() bool {
	_, exists := details[c]
	return exists
}

func (c Country) String() string {
	return string(c)
}

func (c Country) IsSupported() bool {
	d, ok := details[c]
	if !ok {
		return false
	}

	return d.Supported
}

var euCountries = map[Country]bool{
	AT: true, // Austria
	BE: true, // Belgium
	BG: true, // Bulgaria
	HR: true, // Croatia
	CY: true, // Cyprus
	CZ: true, // Czech Republic
	DK: true, // Denmark
	EE: true, // Estonia
	FI: true, // Finland
	FR: true, // France
	DE: true, // Germany
	GR: true, // Greece
	HU: true, // Hungary
	IE: true, // Ireland
	IT: true, // Italy
	LV: true, // Latvia
	LT: true, // Lithuania
	LU: true, // Luxembourg
	MT: true, // Malta
	NL: true, // Netherlands
	PL: true, // Poland
	PT: true, // Portugal
	RO: true, // Romania
	SK: true, // Slovakia
	SI: true, // Slovenia
	ES: true, // Spain
	SE: true, // Sweden
}

// IsEUCountry returns true if the country is a member of the European Union.
func IsEUCountry(c Country) bool {
	return euCountries[c]
}

// EUCountries returns a list of all EU member countries.
func EUCountries() []Country {
	result := make([]Country, 0, len(euCountries))
	for c := range euCountries {
		result = append(result, c)
	}
	return result
}

const (
	AD Country = "AD"
	AE Country = "AE"
	AF Country = "AF"
	AG Country = "AG"
	AI Country = "AI"
	AL Country = "AL"
	AM Country = "AM"
	AO Country = "AO"
	AQ Country = "AQ"
	AR Country = "AR"
	AS Country = "AS"
	AT Country = "AT"
	AU Country = "AU"
	AW Country = "AW"
	AX Country = "AX"
	AZ Country = "AZ"
	BA Country = "BA"
	BB Country = "BB"
	BD Country = "BD"
	BE Country = "BE"
	BF Country = "BF"
	BG Country = "BG"
	BH Country = "BH"
	BI Country = "BI"
	BJ Country = "BJ"
	BL Country = "BL"
	BM Country = "BM"
	BN Country = "BN"
	BO Country = "BO"
	BQ Country = "BQ"
	BR Country = "BR"
	BS Country = "BS"
	BT Country = "BT"
	BV Country = "BV"
	BW Country = "BW"
	BY Country = "BY"
	BZ Country = "BZ"
	CA Country = "CA"
	CC Country = "CC"
	CD Country = "CD"
	CF Country = "CF"
	CG Country = "CG"
	CH Country = "CH"
	CI Country = "CI"
	CK Country = "CK"
	CL Country = "CL"
	CM Country = "CM"
	CN Country = "CN"
	CO Country = "CO"
	CR Country = "CR"
	CU Country = "CU"
	CV Country = "CV"
	CW Country = "CW"
	CX Country = "CX"
	CY Country = "CY"
	CZ Country = "CZ"
	DE Country = "DE"
	DJ Country = "DJ"
	DK Country = "DK"
	DM Country = "DM"
	DO Country = "DO"
	DZ Country = "DZ"
	EC Country = "EC"
	EE Country = "EE"
	EG Country = "EG"
	EH Country = "EH"
	ER Country = "ER"
	ES Country = "ES"
	ET Country = "ET"
	FI Country = "FI"
	FJ Country = "FJ"
	FK Country = "FK"
	FM Country = "FM"
	FO Country = "FO"
	FR Country = "FR"
	GA Country = "GA"
	GB Country = "GB"
	GD Country = "GD"
	GE Country = "GE"
	GF Country = "GF"
	GG Country = "GG"
	GH Country = "GH"
	GI Country = "GI"
	GL Country = "GL"
	GM Country = "GM"
	GN Country = "GN"
	GP Country = "GP"
	GQ Country = "GQ"
	GR Country = "GR"
	GS Country = "GS"
	GT Country = "GT"
	GU Country = "GU"
	GW Country = "GW"
	GY Country = "GY"
	HK Country = "HK"
	HM Country = "HM"
	HN Country = "HN"
	HR Country = "HR"
	HT Country = "HT"
	HU Country = "HU"
	ID Country = "ID"
	IE Country = "IE"
	IL Country = "IL"
	IM Country = "IM"
	IN Country = "IN"
	IO Country = "IO"
	IQ Country = "IQ"
	IR Country = "IR"
	IS Country = "IS"
	IT Country = "IT"
	JE Country = "JE"
	JM Country = "JM"
	JO Country = "JO"
	JP Country = "JP"
	KE Country = "KE"
	KG Country = "KG"
	KH Country = "KH"
	KI Country = "KI"
	KM Country = "KM"
	KN Country = "KN"
	KP Country = "KP"
	KR Country = "KR"
	KW Country = "KW"
	KY Country = "KY"
	KZ Country = "KZ"
	LA Country = "LA"
	LB Country = "LB"
	LC Country = "LC"
	LI Country = "LI"
	LK Country = "LK"
	LR Country = "LR"
	LS Country = "LS"
	LT Country = "LT"
	LU Country = "LU"
	LV Country = "LV"
	LY Country = "LY"
	MA Country = "MA"
	MC Country = "MC"
	MD Country = "MD"
	ME Country = "ME"
	MF Country = "MF"
	MG Country = "MG"
	MH Country = "MH"
	MK Country = "MK"
	ML Country = "ML"
	MM Country = "MM"
	MN Country = "MN"
	MO Country = "MO"
	MP Country = "MP"
	MQ Country = "MQ"
	MR Country = "MR"
	MS Country = "MS"
	MT Country = "MT"
	MU Country = "MU"
	MV Country = "MV"
	MW Country = "MW"
	MX Country = "MX"
	MY Country = "MY"
	MZ Country = "MZ"
	NA Country = "NA"
	NC Country = "NC"
	NE Country = "NE"
	NF Country = "NF"
	NG Country = "NG"
	NI Country = "NI"
	NL Country = "NL"
	NO Country = "NO"
	NP Country = "NP"
	NR Country = "NR"
	NU Country = "NU"
	NZ Country = "NZ"
	OM Country = "OM"
	PA Country = "PA"
	PE Country = "PE"
	PF Country = "PF"
	PG Country = "PG"
	PH Country = "PH"
	PK Country = "PK"
	PL Country = "PL"
	PM Country = "PM"
	PN Country = "PN"
	PR Country = "PR"
	PS Country = "PS"
	PT Country = "PT"
	PW Country = "PW"
	PY Country = "PY"
	QA Country = "QA"
	RE Country = "RE"
	RO Country = "RO"
	RS Country = "RS"
	RU Country = "RU"
	RW Country = "RW"
	SA Country = "SA"
	SB Country = "SB"
	SC Country = "SC"
	SD Country = "SD"
	SE Country = "SE"
	SG Country = "SG"
	SH Country = "SH"
	SI Country = "SI"
	SJ Country = "SJ"
	SK Country = "SK"
	SL Country = "SL"
	SM Country = "SM"
	SN Country = "SN"
	SO Country = "SO"
	SR Country = "SR"
	SS Country = "SS"
	ST Country = "ST"
	SV Country = "SV"
	SX Country = "SX"
	SY Country = "SY"
	SZ Country = "SZ"
	TC Country = "TC"
	TD Country = "TD"
	TF Country = "TF"
	TG Country = "TG"
	TH Country = "TH"
	TJ Country = "TJ"
	TK Country = "TK"
	TL Country = "TL"
	TM Country = "TM"
	TN Country = "TN"
	TO Country = "TO"
	TR Country = "TR"
	TT Country = "TT"
	TV Country = "TV"
	TW Country = "TW"
	TZ Country = "TZ"
	UA Country = "UA"
	UG Country = "UG"
	UM Country = "UM"
	US Country = "US"
	UY Country = "UY"
	UZ Country = "UZ"
	VA Country = "VA"
	VC Country = "VC"
	VE Country = "VE"
	VG Country = "VG"
	VI Country = "VI"
	VN Country = "VN"
	VU Country = "VU"
	WF Country = "WF"
	WS Country = "WS"
	XK Country = "XK"
	YE Country = "YE"
	YT Country = "YT"
	ZA Country = "ZA"
	ZM Country = "ZM"
	ZW Country = "ZW"
)

type Detail struct {
	Name      string
	Numeric   string
	Supported bool
}

var details = map[Country]Detail{
	AD: {Name: "Andorra", Numeric: "020"},
	AE: {Name: "United Arab Emirates", Numeric: "784"},
	AF: {Name: "Afghanistan", Numeric: "004"},
	AG: {Name: "Antigua and Barbuda", Numeric: "028"},
	AI: {Name: "Anguilla", Numeric: "660"},
	AL: {Name: "Albania", Numeric: "008"},
	AM: {Name: "Armenia", Numeric: "051"},
	AO: {Name: "Angola", Numeric: "024"},
	AQ: {Name: "Antarctica", Numeric: "010"},
	AR: {Name: "Argentina", Numeric: "032"},
	AS: {Name: "American Samoa", Numeric: "016"},
	AT: {Name: "Austria", Numeric: "040", Supported: true},
	AU: {Name: "Australia", Numeric: "036"},
	AW: {Name: "Aruba", Numeric: "533"},
	AX: {Name: "Åland Islands", Numeric: "248"},
	AZ: {Name: "Azerbaijan", Numeric: "031"},
	BA: {Name: "Bosnia and Herzegovina", Numeric: "070"},
	BB: {Name: "Barbados", Numeric: "052"},
	BD: {Name: "Bangladesh", Numeric: "050"},
	BE: {Name: "Belgium", Numeric: "056", Supported: true},
	BF: {Name: "Burkina Faso", Numeric: "854"},
	BG: {Name: "Bulgaria", Numeric: "100", Supported: true},
	BH: {Name: "Bahrain", Numeric: "048"},
	BI: {Name: "Burundi", Numeric: "108"},
	BJ: {Name: "Benin", Numeric: "204"},
	BL: {Name: "Saint Barthélemy", Numeric: "652"},
	BM: {Name: "Bermuda", Numeric: "060"},
	BN: {Name: "Brunei Darussalam", Numeric: "096"},
	BO: {Name: "Bolivia, Plurinational State of", Numeric: "068"},
	BQ: {Name: "Bonaire, Sint Eustatius and Saba", Numeric: "535"},
	BR: {Name: "Brazil", Numeric: "076"},
	BS: {Name: "Bahamas", Numeric: "044"},
	BT: {Name: "Bhutan", Numeric: "064"},
	BV: {Name: "Bouvet Island", Numeric: "074"},
	BW: {Name: "Botswana", Numeric: "072"},
	BY: {Name: "Belarus", Numeric: "112"},
	BZ: {Name: "Belize", Numeric: "084"},
	CA: {Name: "Canada", Numeric: "124", Supported: true},
	CC: {Name: "Cocos (Keeling) Islands", Numeric: "166"},
	CD: {Name: "Congo, the Democratic Republic of the", Numeric: "180"},
	CF: {Name: "Central African Republic", Numeric: "140"},
	CG: {Name: "Congo", Numeric: "178"},
	CH: {Name: "Switzerland", Numeric: "756"},
	CI: {Name: "Côte d'Ivoire", Numeric: "384"},
	CK: {Name: "Cook Islands", Numeric: "184"},
	CL: {Name: "Chile", Numeric: "152"},
	CM: {Name: "Cameroon", Numeric: "120"},
	CN: {Name: "China", Numeric: "156"},
	CO: {Name: "Colombia", Numeric: "170"},
	CR: {Name: "Costa Rica", Numeric: "188"},
	CU: {Name: "Cuba", Numeric: "192"},
	CV: {Name: "Cabo Verde", Numeric: "132"},
	CW: {Name: "Curaçao", Numeric: "531"},
	CX: {Name: "Christmas Island", Numeric: "162"},
	CY: {Name: "Cyprus", Numeric: "196"},
	CZ: {Name: "Czech Republic", Numeric: "203", Supported: true},
	DE: {Name: "Germany", Numeric: "276", Supported: true},
	DJ: {Name: "Djibouti", Numeric: "262"},
	DK: {Name: "Denmark", Numeric: "208", Supported: true},
	DM: {Name: "Dominica", Numeric: "212"},
	DO: {Name: "Dominican Republic", Numeric: "214"},
	DZ: {Name: "Algeria", Numeric: "012"},
	EC: {Name: "Ecuador", Numeric: "218"},
	EE: {Name: "Estonia", Numeric: "233", Supported: true},
	EG: {Name: "Egypt", Numeric: "818"},
	EH: {Name: "Western Sahara", Numeric: "732"},
	ER: {Name: "Eritrea", Numeric: "232"},
	ES: {Name: "Spain", Numeric: "724", Supported: true},
	ET: {Name: "Ethiopia", Numeric: "231"},
	FI: {Name: "Finland", Numeric: "246", Supported: true},
	FJ: {Name: "Fiji", Numeric: "242"},
	FK: {Name: "Falkland Islands (Malvinas)", Numeric: "238"},
	FM: {Name: "Micronesia, Federated States of", Numeric: "583"},
	FO: {Name: "Faroe Islands", Numeric: "234"},
	FR: {Name: "France", Numeric: "250", Supported: true},
	GA: {Name: "Gabon", Numeric: "266"},
	GB: {Name: "United Kingdom of Great Britain and Northern Ireland", Numeric: "826", Supported: true},
	GD: {Name: "Grenada", Numeric: "308"},
	GE: {Name: "Georgia", Numeric: "268"},
	GF: {Name: "French Guiana", Numeric: "254"},
	GG: {Name: "Guernsey", Numeric: "831"},
	GH: {Name: "Ghana", Numeric: "288"},
	GI: {Name: "Gibraltar", Numeric: "292"},
	GL: {Name: "Greenland", Numeric: "304"},
	GM: {Name: "Gambia", Numeric: "270"},
	GN: {Name: "Guinea", Numeric: "324"},
	GP: {Name: "Guadeloupe", Numeric: "312"},
	GQ: {Name: "Equatorial Guinea", Numeric: "226"},
	GR: {Name: "Greece", Numeric: "300", Supported: true},
	GS: {Name: "South Georgia and the South Sandwich Islands", Numeric: "239"},
	GT: {Name: "Guatemala", Numeric: "320"},
	GU: {Name: "Guam", Numeric: "316"},
	GW: {Name: "Guinea-Bissau", Numeric: "624"},
	GY: {Name: "Guyana", Numeric: "328"},
	HK: {Name: "Hong Kong", Numeric: "344"},
	HM: {Name: "Heard Island and McDonald Islands", Numeric: "334"},
	HN: {Name: "Honduras", Numeric: "340"},
	HR: {Name: "Croatia", Numeric: "191", Supported: true},
	HT: {Name: "Haiti", Numeric: "332"},
	HU: {Name: "Hungary", Numeric: "348", Supported: true},
	ID: {Name: "Indonesia", Numeric: "360"},
	IE: {Name: "Ireland", Numeric: "372", Supported: true},
	IL: {Name: "Israel", Numeric: "376"},
	IM: {Name: "Isle of Man", Numeric: "833"},
	IN: {Name: "India", Numeric: "356", Supported: true},
	IO: {Name: "British Indian Ocean Territory", Numeric: "086"},
	IQ: {Name: "Iraq", Numeric: "368"},
	IR: {Name: "Iran, Islamic Republic of", Numeric: "364"},
	IS: {Name: "Iceland", Numeric: "352"},
	IT: {Name: "Italy", Numeric: "380", Supported: true},
	JE: {Name: "Jersey", Numeric: "832"},
	JM: {Name: "Jamaica", Numeric: "388"},
	JO: {Name: "Jordan", Numeric: "400"},
	JP: {Name: "Japan", Numeric: "392", Supported: true},
	KE: {Name: "Kenya", Numeric: "404"},
	KG: {Name: "Kyrgyzstan", Numeric: "417"},
	KH: {Name: "Cambodia", Numeric: "116"},
	KI: {Name: "Kiribati", Numeric: "296"},
	KM: {Name: "Comoros", Numeric: "174"},
	KN: {Name: "Saint Kitts and Nevis", Numeric: "659"},
	KP: {Name: "Korea, Democratic People's Republic of", Numeric: "408"},
	KR: {Name: "Korea, Republic of", Numeric: "410"},
	KW: {Name: "Kuwait", Numeric: "414"},
	KY: {Name: "Cayman Islands", Numeric: "136"},
	KZ: {Name: "Kazakhstan", Numeric: "398"},
	LA: {Name: "Lao People's Democratic Republic", Numeric: "418"},
	LB: {Name: "Lebanon", Numeric: "422"},
	LC: {Name: "Saint Lucia", Numeric: "662"},
	LI: {Name: "Liechtenstein", Numeric: "438"},
	LK: {Name: "Sri Lanka", Numeric: "144"},
	LR: {Name: "Liberia", Numeric: "430"},
	LS: {Name: "Lesotho", Numeric: "426"},
	LT: {Name: "Lithuania", Numeric: "440", Supported: true},
	LU: {Name: "Luxembourg", Numeric: "442", Supported: true},
	LV: {Name: "Latvia", Numeric: "428", Supported: true},
	LY: {Name: "Libya", Numeric: "434"},
	MA: {Name: "Morocco", Numeric: "504"},
	MC: {Name: "Monaco", Numeric: "492"},
	MD: {Name: "Moldova, Republic of", Numeric: "498"},
	ME: {Name: "Montenegro", Numeric: "499"},
	MF: {Name: "Saint Martin (French part)", Numeric: "663"},
	MG: {Name: "Madagascar", Numeric: "450"},
	MH: {Name: "Marshall Islands", Numeric: "584"},
	MK: {Name: "North Macedonia", Numeric: "807"},
	ML: {Name: "Mali", Numeric: "466"},
	MM: {Name: "Myanmar", Numeric: "104"},
	MN: {Name: "Mongolia", Numeric: "496"},
	MO: {Name: "Macao", Numeric: "446"},
	MP: {Name: "Northern Mariana Islands", Numeric: "580"},
	MQ: {Name: "Martinique", Numeric: "474"},
	MR: {Name: "Mauritania", Numeric: "478"},
	MS: {Name: "Montserrat", Numeric: "500"},
	MT: {Name: "Malta", Numeric: "470"},
	MU: {Name: "Mauritius", Numeric: "480"},
	MV: {Name: "Maldives", Numeric: "462"},
	MW: {Name: "Malawi", Numeric: "454"},
	MX: {Name: "Mexico", Numeric: "484"},
	MY: {Name: "Malaysia", Numeric: "458"},
	MZ: {Name: "Mozambique", Numeric: "508"},
	NA: {Name: "Namibia", Numeric: "516"},
	NC: {Name: "New Caledonia", Numeric: "540"},
	NE: {Name: "Niger", Numeric: "562"},
	NF: {Name: "Norfolk Island", Numeric: "574"},
	NG: {Name: "Nigeria", Numeric: "566"},
	NI: {Name: "Nicaragua", Numeric: "558"},
	NL: {Name: "Netherlands", Numeric: "528", Supported: true},
	NO: {Name: "Norway", Numeric: "578"},
	NP: {Name: "Nepal", Numeric: "524"},
	NR: {Name: "Nauru", Numeric: "520"},
	NU: {Name: "Niue", Numeric: "570"},
	NZ: {Name: "New Zealand", Numeric: "554"},
	OM: {Name: "Oman", Numeric: "512"},
	PA: {Name: "Panama", Numeric: "591"},
	PE: {Name: "Peru", Numeric: "604"},
	PF: {Name: "French Polynesia", Numeric: "258"},
	PG: {Name: "Papua New Guinea", Numeric: "598"},
	PH: {Name: "Philippines", Numeric: "608"},
	PK: {Name: "Pakistan", Numeric: "586"},
	PL: {Name: "Poland", Numeric: "616", Supported: true},
	PM: {Name: "Saint Pierre and Miquelon", Numeric: "666"},
	PN: {Name: "Pitcairn", Numeric: "612"},
	PR: {Name: "Puerto Rico", Numeric: "630"},
	PS: {Name: "Palestine, State of", Numeric: "275"},
	PT: {Name: "Portugal", Numeric: "620"},
	PW: {Name: "Palau", Numeric: "585"},
	PY: {Name: "Paraguay", Numeric: "600"},
	QA: {Name: "Qatar", Numeric: "634"},
	RE: {Name: "Réunion", Numeric: "638"},
	RO: {Name: "Romania", Numeric: "642", Supported: true},
	RS: {Name: "Serbia", Numeric: "688"},
	RU: {Name: "Russian Federation", Numeric: "643"},
	RW: {Name: "Rwanda", Numeric: "646"},
	SA: {Name: "Saudi Arabia", Numeric: "682"},
	SB: {Name: "Solomon Islands", Numeric: "090"},
	SC: {Name: "Seychelles", Numeric: "690"},
	SD: {Name: "Sudan", Numeric: "729"},
	SE: {Name: "Sweden", Numeric: "752", Supported: true},
	SG: {Name: "Singapore", Numeric: "702"},
	SH: {Name: "Saint Helena, Ascension and Tristan da Cunha", Numeric: "654"},
	SI: {Name: "Slovenia", Numeric: "705", Supported: true},
	SJ: {Name: "Svalbard and Jan Mayen", Numeric: "744"},
	SK: {Name: "Slovakia", Numeric: "703", Supported: true},
	SL: {Name: "Sierra Leone", Numeric: "694"},
	SM: {Name: "San Marino", Numeric: "674"},
	SN: {Name: "Senegal", Numeric: "686"},
	SO: {Name: "Somalia", Numeric: "706"},
	SR: {Name: "Suriname", Numeric: "740"},
	SS: {Name: "South Sudan", Numeric: "728"},
	ST: {Name: "Sao Tome and Principe", Numeric: "678"},
	SV: {Name: "El Salvador", Numeric: "222"},
	SX: {Name: "Sint Maarten (Dutch part)", Numeric: "534"},
	SY: {Name: "Syrian Arab Republic", Numeric: "760"},
	SZ: {Name: "Eswatini", Numeric: "748"},
	TC: {Name: "Turks and Caicos Islands", Numeric: "796"},
	TD: {Name: "Chad", Numeric: "148"},
	TF: {Name: "French Southern Territories", Numeric: "260"},
	TG: {Name: "Togo", Numeric: "768"},
	TH: {Name: "Thailand", Numeric: "764"},
	TJ: {Name: "Tajikistan", Numeric: "762"},
	TK: {Name: "Tokelau", Numeric: "772"},
	TL: {Name: "Timor-Leste", Numeric: "626"},
	TM: {Name: "Turkmenistan", Numeric: "795"},
	TN: {Name: "Tunisia", Numeric: "788"},
	TO: {Name: "Tonga", Numeric: "776"},
	TR: {Name: "Turkey", Numeric: "792"},
	TT: {Name: "Trinidad and Tobago", Numeric: "780"},
	TV: {Name: "Tuvalu", Numeric: "798"},
	TW: {Name: "Taiwan, Province of China", Numeric: "158"},
	TZ: {Name: "Tanzania, United Republic of", Numeric: "834"},
	UA: {Name: "Ukraine", Numeric: "804"},
	UG: {Name: "Uganda", Numeric: "800"},
	UM: {Name: "United States Minor Outlying Islands", Numeric: "581"},
	US: {Name: "United States of America", Numeric: "840", Supported: true},
	UY: {Name: "Uruguay", Numeric: "858"},
	UZ: {Name: "Uzbekistan", Numeric: "860"},
	VA: {Name: "Holy See", Numeric: "336"},
	VC: {Name: "Saint Vincent and the Grenadines", Numeric: "670"},
	VE: {Name: "Venezuela (Bolivarian Republic of)", Numeric: "862"},
	VG: {Name: "Virgin Islands (British)", Numeric: "092"},
	VI: {Name: "Virgin Islands (U.S.)", Numeric: "850"},
	VN: {Name: "Viet Nam", Numeric: "704"},
	VU: {Name: "Vanuatu", Numeric: "548"},
	WF: {Name: "Wallis and Futuna", Numeric: "876"},
	WS: {Name: "Samoa", Numeric: "882"},
	XK: {Name: "Kosovo", Numeric: "383"},
	YE: {Name: "Yemen", Numeric: "887"},
	YT: {Name: "Mayotte", Numeric: "175"},
	ZA: {Name: "South Africa", Numeric: "710", Supported: true},
	ZM: {Name: "Zambia", Numeric: "894"},
	ZW: {Name: "Zimbabwe", Numeric: "716"},
}

// GetCountryDetail returns the detail for a country and a boolean indicating if it exists.
func GetCountryDetail(c Country) (Detail, bool) {
	d, ok := details[c]
	return d, ok
}

// AllCountries returns a list of all known countries.
func AllCountries() []Country {
	result := make([]Country, 0, len(details))
	for c := range details {
		result = append(result, c)
	}
	return result
}

var states = map[Country]map[string]string{
	GB: {
		"CAM": "Cambridgeshire",
		"CMA": "Cumbria",
		"DBY": "Derbyshire",
		"DEV": "Devon",
		"DOR": "Dorset",
		"ESX": "East Sussex",
		"ESS": "Essex",
		"GLS": "Gloucestershire",
		"HAM": "Hampshire",
		"HRT": "Hertfordshire",
		"KEN": "Kent",
		"LAN": "Lancashire",
		"LEC": "Leicestershire",
		"LIN": "Lincolnshire",
		"NFK": "Norfolk",
		"NYK": "North Yorkshire",
		"NTT": "Nottinghamshire",
		"OXF": "Oxfordshire",
		"SOM": "Somerset",
		"STS": "Staffordshire",
		"SFK": "Suffolk",
		"SRY": "Surrey",
		"WAR": "Warwickshire",
		"WSX": "West Sussex",
		"WOR": "Worcestershire",
		"LND": "London, City of",
		"BDG": "Barking and Dagenham",
		"BNE": "Barnet",
		"BEX": "Bexley",
		"BEN": "Brent",
		"BRY": "Bromley",
		"CMD": "Camden",
		"CRY": "Croydon",
		"EAL": "Ealing",
		"ENF": "Enfield",
		"GRE": "Greenwich",
		"HCK": "Hackney",
		"HMF": "Hammersmith and Fulham",
		"HRY": "Haringey",
		"HRW": "Harrow",
		"HAV": "Havering",
		"HIL": "Hillingdon",
		"HNS": "Hounslow",
		"ISL": "Islington",
		"KEC": "Kensington and Chelsea",
		"KTT": "Kingston upon Thames",
		"LBH": "Lambeth",
		"LEW": "Lewisham",
		"MRT": "Merton",
		"NWM": "Newham",
		"RDB": "Redbridge",
		"RIC": "Richmond upon Thames",
		"SWK": "Southwark",
		"STN": "Sutton",
		"TWH": "Tower Hamlets",
		"WFT": "Waltham Forest",
		"WND": "Wandsworth",
		"WSM": "Westminster",
		"BNS": "Barnsley",
		"BIR": "Birmingham",
		"BOL": "Bolton",
		"BRD": "Bradford",
		"BUR": "Bury",
		"CLD": "Calderdale",
		"COV": "Coventry",
		"DNC": "Doncaster",
		"DUD": "Dudley",
		"GAT": "Gateshead",
		"KIR": "Kirklees",
		"KWL": "Knowsley",
		"LDS": "Leeds",
		"LIV": "Liverpool",
		"MAN": "Manchester",
		"NET": "Newcastle upon Tyne",
		"NTY": "North Tyneside",
		"OLD": "Oldham",
		"RCH": "Rochdale",
		"ROT": "Rotherham",
		"SHN": "St. Helens",
		"SLF": "Salford",
		"SAW": "Sandwell",
		"SFT": "Sefton",
		"SHF": "Sheffield",
		"SOL": "Solihull",
		"STY": "South Tyneside",
		"SKP": "Stockport",
		"SND": "Sunderland",
		"TAM": "Tameside",
		"TRF": "Trafford",
		"WKF": "Wakefield",
		"WLL": "Walsall",
		"WGN": "Wigan",
		"WRL": "Wirral",
		"WLV": "Wolverhampton",
		"BAS": "Bath and North East Somerset",
		"BDF": "Bedford",
		"BBD": "Blackburn with Darwen",
		"BPL": "Blackpool",
		"BCP": "Bournemouth, Christchurch and Poole",
		"BRC": "Bracknell Forest",
		"BNH": "Brighton and Hove",
		"BST": "Bristol, City of",
		"BKM": "Buckinghamshire",
		"CBF": "Central Bedfordshire",
		"CHE": "Cheshire East",
		"CHW": "Cheshire West and Chester",
		"CON": "Cornwall",
		"DAL": "Darlington",
		"DER": "Derby",
		"DUR": "Durham, County",
		"ERY": "East Riding of Yorkshire",
		"HAL": "Halton",
		"HPL": "Hartlepool",
		"HEF": "Herefordshire",
		"IOW": "Isle of Wight",
		"IOS": "Isles of Scilly",
		"KHL": "Kingston upon Hull",
		"LCE": "Leicester",
		"LUT": "Luton",
		"MDW": "Medway",
		"MDB": "Middlesbrough",
		"MIK": "Milton Keynes",
		"NEL": "North East Lincolnshire",
		"NLN": "North Lincolnshire",
		"NNH": "North Northamptonshire",
		"NSM": "North Somerset",
		"NBL": "Northumberland",
		"NGM": "Nottingham",
		"PTE": "Peterborough",
		"PLY": "Plymouth",
		"POR": "Portsmouth",
		"RDG": "Reading",
		"RCC": "Redcar and Cleveland",
		"RUT": "Rutland",
		"SHR": "Shropshire",
		"SLG": "Slough",
		"SGC": "South Gloucestershire",
		"STH": "Southampton",
		"SOS": "Southend-on-Sea",
		"STT": "Stockton-on-Tees",
		"STE": "Stoke-on-Trent",
		"SWD": "Swindon",
		"TFW": "Telford and Wrekin",
		"THR": "Thurrock",
		"TOB": "Torbay",
		"WRT": "Warrington",
		"WBK": "West Berkshire",
		"WNH": "West Northamptonshire",
		"WIL": "Wiltshire",
		"WNM": "Windsor and Maidenhead",
		"WOK": "Wokingham",
		"YOR": "York",
		"ANN": "Antrim and Newtownabbey",
		"AND": "Ards and North Down",
		"ABC": "Armagh City, Banbridge and Craigavon",
		"BFS": "Belfast City",
		"CCG": "Causeway Coast and Glens",
		"DRS": "Derry and Strabane",
		"FMO": "Fermanagh and Omagh",
		"LBC": "Lisburn and Castlereagh",
		"MEA": "Mid and East Antrim",
		"MUL": "Mid-Ulster",
		"NMD": "Newry, Mourne and Down",
		"ABE": "Aberdeen City",
		"ABD": "Aberdeenshire",
		"ANS": "Angus",
		"AGB": "Argyll and Bute",
		"CLK": "Clackmannanshire",
		"DGY": "Dumfries and Galloway",
		"DND": "Dundee City",
		"EAY": "East Ayrshire",
		"EDU": "East Dunbartonshire",
		"ELN": "East Lothian",
		"ERW": "East Renfrewshire",
		"EDH": "Edinburgh, City of",
		"ELS": "Eilean Siar",
		"FAL": "Falkirk",
		"FIF": "Fife",
		"GLG": "Glasgow City",
		"HLD": "Highland",
		"IVC": "Inverclyde",
		"MLN": "Midlothian",
		"MRY": "Moray",
		"NAY": "North Ayrshire",
		"NLK": "North Lanarkshire",
		"ORK": "Orkney Islands",
		"PKN": "Perth and Kinross",
		"RFW": "Renfrewshire",
		"SCB": "Scottish Borders",
		"ZET": "Shetland Islands",
		"SAY": "South Ayrshire",
		"SLK": "South Lanarkshire",
		"STG": "Stirling",
		"WDU": "West Dunbartonshire",
		"WLN": "West Lothian",
		"BGW": "Blaenau Gwent",
		"BGE": "Bridgend",
		"CAY": "Caerphilly",
		"CRF": "Cardiff",
		"CMN": "Carmarthenshire",
		"CGN": "Ceredigion [Sir Ceredigion]",
		"CWY": "Conwy",
		"DEN": "Denbighshire",
		"FLN": "Flintshire",
		"GWN": "Gwynedd",
		"AGY": "Isle of Anglesey",
		"MTY": "Merthyr Tydfil",
		"MON": "Monmouthshire",
		"NTL": "Neath Port Talbot",
		"NWP": "Newport",
		"PEM": "Pembrokeshire",
		"POW": "Powys",
		"RCT": "Rhondda Cynon Taff",
		"SWA": "Swansea",
		"TOF": "Torfaen",
		"VGL": "Vale of Glamorgan",
		"WRX": "Wrexham",
		"ENG": "England",
		"NIR": "Northern Ireland",
		"SCT": "Scotland",
		"WLS": "Wales",
		"EAW": "England and Wales",
		"GBN": "Great Britain",
		"UKM": "United Kingdom",
	},
	US: {"AL": "ALABAMA",
		"AK": "ALASKA",
		"AZ": "ARIZONA",
		"AR": "ARKANSAS",
		"CA": "CALIFORNIA",
		"CO": "COLORADO",
		"CT": "CONNECTICUT",
		"DE": "DELAWARE",
		"FL": "FLORIDA",
		"GA": "GEORGIA",
		"HI": "HAWAII",
		"ID": "IDAHO",
		"IL": "ILLINOIS",
		"IN": "INDIANA",
		"IA": "IOWA",
		"KS": "KANSAS",
		"KY": "KENTUCKY",
		"LA": "LOUISIANA",
		"ME": "MAINE",
		"MD": "MARYLAND",
		"MA": "MASSACHUSETTS",
		"MI": "MICHIGAN",
		"MN": "MINNESOTA",
		"MS": "MISSISSIPPI",
		"MO": "MISSOURI",
		"MT": "MONTANA",
		"NE": "NEBRASKA",
		"NV": "NEVADA",
		"NH": "NEW HAMPSHIRE",
		"NJ": "NEW JERSEY",
		"NM": "NEW MEXICO",
		"NY": "NEW YORK",
		"NC": "NORTH CAROLINA",
		"ND": "NORTH DAKOTA",
		"OH": "OHIO",
		"OK": "OKLAHOMA",
		"OR": "OREGON",
		"PA": "PENNSYLVANIA",
		"RI": "RHODE ISLAND",
		"SC": "SOUTH CAROLINA",
		"SD": "SOUTH DAKOTA",
		"TN": "TENNESSEE",
		"TX": "TEXAS",
		"UT": "UTAH",
		"VT": "VERMONT",
		"VA": "VIRGINIA",
		"WA": "WASHINGTON",
		"WV": "WEST VIRGINIA",
		"WI": "WISCONSIN",
		"WY": "WYOMING",
		"DC": "DISTRICT OF COLUMBIA",
	},
}

// GetStates returns the states/subdivisions for a country and a boolean indicating if any exist.
func GetStates(c Country) (map[string]string, bool) {
	s, ok := states[c]
	if !ok {
		return nil, false
	}
	// Return a copy to prevent modification
	result := make(map[string]string, len(s))
	for k, v := range s {
		result[k] = v
	}
	return result, true
}

// GetStateName returns the name of a state/subdivision given its code and country.
func GetStateName(c Country, stateCode string) (string, bool) {
	s, ok := states[c]
	if !ok {
		return "", false
	}
	name, ok := s[stateCode]
	return name, ok
}

var alpha2ToAlpha3 = map[string]string{
	"AD": "AND",
	"AE": "ARE",
	"AF": "AFG",
	"AG": "ATG",
	"AI": "AIA",
	"AL": "ALB",
	"AM": "ARM",
	"AO": "AGO",
	"AQ": "ATA",
	"AR": "ARG",
	"AS": "ASM",
	"AT": "AUT",
	"AU": "AUS",
	"AW": "ABW",
	"AX": "ALA",
	"AZ": "AZE",
	"BA": "BIH",
	"BB": "BRB",
	"BD": "BGD",
	"BE": "BEL",
	"BF": "BFA",
	"BG": "BGR",
	"BH": "BHR",
	"BI": "BDI",
	"BJ": "BEN",
	"BL": "BLM",
	"BM": "BMU",
	"BN": "BRN",
	"BO": "BOL",
	"BQ": "BES",
	"BR": "BRA",
	"BS": "BHS",
	"BT": "BTN",
	"BV": "BVT",
	"BW": "BWA",
	"BY": "BLR",
	"BZ": "BLZ",
	"CA": "CAN",
	"CC": "CCK",
	"CD": "COD",
	"CF": "CAF",
	"CG": "COG",
	"CH": "CHE",
	"CI": "CIV",
	"CK": "COK",
	"CL": "CHL",
	"CM": "CMR",
	"CN": "CHN",
	"CO": "COL",
	"CR": "CRI",
	"CU": "CUB",
	"CV": "CPV",
	"CW": "CUW",
	"CX": "CXR",
	"CY": "CYP",
	"CZ": "CZE",
	"DE": "DEU",
	"DJ": "DJI",
	"DK": "DNK",
	"DM": "DMA",
	"DO": "DOM",
	"DZ": "DZA",
	"EC": "ECU",
	"EE": "EST",
	"EG": "EGY",
	"EH": "ESH",
	"ER": "ERI",
	"ES": "ESP",
	"ET": "ETH",
	"FI": "FIN",
	"FJ": "FJI",
	"FK": "FLK",
	"FM": "FSM",
	"FO": "FRO",
	"FR": "FRA",
	"GA": "GAB",
	"GB": "GBR",
	"GD": "GRD",
	"GE": "GEO",
	"GF": "GUF",
	"GG": "GGY",
	"GH": "GHA",
	"GI": "GIB",
	"GL": "GRL",
	"GM": "GMB",
	"GN": "GIN",
	"GP": "GLP",
	"GQ": "GNQ",
	"GR": "GRC",
	"GS": "SGS",
	"GT": "GTM",
	"GU": "GUM",
	"GW": "GNB",
	"GY": "GUY",
	"HK": "HKG",
	"HM": "HMD",
	"HN": "HND",
	"HR": "HRV",
	"HT": "HTI",
	"HU": "HUN",
	"ID": "IDN",
	"IE": "IRL",
	"IL": "ISR",
	"IM": "IMN",
	"IN": "IND",
	"IO": "IOT",
	"IQ": "IRQ",
	"IR": "IRN",
	"IS": "ISL",
	"IT": "ITA",
	"JE": "JEY",
	"JM": "JAM",
	"JO": "JOR",
	"JP": "JPN",
	"KE": "KEN",
	"KG": "KGZ",
	"KH": "KHM",
	"KI": "KIR",
	"KM": "COM",
	"KN": "KNA",
	"KP": "PRK",
	"KR": "KOR",
	"KW": "KWT",
	"KY": "CYM",
	"KZ": "KAZ",
	"LA": "LAO",
	"LB": "LBN",
	"LC": "LCA",
	"LI": "LIE",
	"LK": "LKA",
	"LR": "LBR",
	"LS": "LSO",
	"LT": "LTU",
	"LU": "LUX",
	"LV": "LVA",
	"LY": "LBY",
	"MA": "MAR",
	"MC": "MCO",
	"MD": "MDA",
	"ME": "MNE",
	"MF": "MAF",
	"MG": "MDG",
	"MH": "MHL",
	"MK": "MKD",
	"ML": "MLI",
	"MM": "MMR",
	"MN": "MNG",
	"MO": "MAC",
	"MP": "MNP",
	"MQ": "MTQ",
	"MR": "MRT",
	"MS": "MSR",
	"MT": "MLT",
	"MU": "MUS",
	"MV": "MDV",
	"MW": "MWI",
	"MX": "MEX",
	"MY": "MYS",
	"MZ": "MOZ",
	"NA": "NAM",
	"NC": "NCL",
	"NE": "NER",
	"NF": "NFK",
	"NG": "NGA",
	"NI": "NIC",
	"NL": "NLD",
	"NO": "NOR",
	"NP": "NPL",
	"NR": "NRU",
	"NU": "NIU",
	"NZ": "NZL",
	"OM": "OMN",
	"PA": "PAN",
	"PE": "PER",
	"PF": "PYF",
	"PG": "PNG",
	"PH": "PHL",
	"PK": "PAK",
	"PL": "POL",
	"PM": "SPM",
	"PN": "PCN",
	"PR": "PRI",
	"PS": "PSE",
	"PT": "PRT",
	"PW": "PLW",
	"PY": "PRY",
	"QA": "QAT",
	"RE": "REU",
	"RO": "ROU",
	"RS": "SRB",
	"RU": "RUS",
	"RW": "RWA",
	"SA": "SAU",
	"SB": "SLB",
	"SC": "SYC",
	"SD": "SDN",
	"SE": "SWE",
	"SG": "SGP",
	"SH": "SHN",
	"SI": "SVN",
	"SJ": "SJM",
	"SK": "SVK",
	"SL": "SLE",
	"SM": "SMR",
	"SN": "SEN",
	"SO": "SOM",
	"SR": "SUR",
	"SS": "SSD",
	"ST": "STP",
	"SV": "SLV",
	"SX": "SXM",
	"SY": "SYR",
	"SZ": "SWZ",
	"TC": "TCA",
	"TD": "TCD",
	"TF": "ATF",
	"TG": "TGO",
	"TH": "THA",
	"TJ": "TJK",
	"TK": "TKL",
	"TL": "TLS",
	"TM": "TKM",
	"TN": "TUN",
	"TO": "TON",
	"TR": "TUR",
	"TT": "TTO",
	"TV": "TUV",
	"TW": "TWN",
	"TZ": "TZA",
	"UA": "UKR",
	"UG": "UGA",
	"UM": "UMI",
	"US": "USA",
	"UY": "URY",
	"UZ": "UZB",
	"VA": "VAT",
	"VC": "VCT",
	"VE": "VEN",
	"VG": "VGB",
	"VI": "VIR",
	"VN": "VNM",
	"VU": "VUT",
	"WF": "WLF",
	"WS": "WSM",
	"XK": "XKX",
	"YE": "YEM",
	"YT": "MYT",
	"ZA": "ZAF",
	"ZM": "ZMB",
	"ZW": "ZWE",
}

func ToAlpha3(alpha2 string) string {
	if v, ok := alpha2ToAlpha3[strings.ToUpper(alpha2)]; ok {
		return v
	}
	return ""
}
