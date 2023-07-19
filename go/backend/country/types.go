package country

type Country string

func (c Country) Numeric() (string, error) {
	details, ok := Details[c]
	if !ok {
		return "", ErrNotFound
	}

	return details.Numeric, nil
}

func (c Country) Valid() bool {
	_, exists := Details[c]
	return exists
}

func (c Country) String() string {
	return string(c)
}

func (c Country) IsSupported() bool {
	details, ok := Details[c]
	if !ok {
		return false
	}

	return details.Supported
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

var Details = map[Country]Detail{
	US: {Name: "United States of America", Numeric: "840", Supported: true},
	GB: {Name: "United Kingdom of Great Britain and Northern Ireland", Numeric: "826", Supported: true},
	AX: {Name: "Åland Islands", Numeric: "248", Supported: false},
	AF: {Name: "Afghanistan", Numeric: "004", Supported: false},
	AL: {Name: "Albania", Numeric: "008", Supported: false},
	DZ: {Name: "Algeria", Numeric: "012", Supported: false},
	AS: {Name: "American Samoa", Numeric: "016", Supported: false},
	AD: {Name: "Andorra", Numeric: "020", Supported: false},
	AO: {Name: "Angola", Numeric: "024", Supported: false},
	AI: {Name: "Anguilla", Numeric: "660", Supported: false},
	AQ: {Name: "Antarctica", Numeric: "010", Supported: false},
	AG: {Name: "Antigua and Barbuda", Numeric: "028", Supported: false},
	AR: {Name: "Argentina", Numeric: "032", Supported: false},
	AM: {Name: "Armenia", Numeric: "051", Supported: false},
	AW: {Name: "Aruba", Numeric: "533", Supported: false},
	AU: {Name: "Australia", Numeric: "036", Supported: false},
	AT: {Name: "Austria", Numeric: "040", Supported: false},
	AZ: {Name: "Azerbaijan", Numeric: "031", Supported: false},
	BS: {Name: "Bahamas", Numeric: "044", Supported: false},
	BH: {Name: "Bahrain", Numeric: "048", Supported: false},
	BD: {Name: "Bangladesh", Numeric: "050", Supported: false},
	BB: {Name: "Barbados", Numeric: "052", Supported: false},
	BY: {Name: "Belarus", Numeric: "112", Supported: false},
	BE: {Name: "Belgium", Numeric: "056", Supported: false},
	BZ: {Name: "Belize", Numeric: "084", Supported: false},
	BJ: {Name: "Benin", Numeric: "204", Supported: false},
	BM: {Name: "Bermuda", Numeric: "060", Supported: false},
	BT: {Name: "Bhutan", Numeric: "064", Supported: false},
	BO: {Name: "Bolivia, Plurinational State of", Numeric: "068", Supported: false},
	BQ: {Name: "Bonaire, Sint Eustatius and Saba", Numeric: "535", Supported: false},
	BA: {Name: "Bosnia and Herzegovina", Numeric: "070", Supported: false},
	BW: {Name: "Botswana", Numeric: "072", Supported: false},
	BV: {Name: "Bouvet Island", Numeric: "074", Supported: false},
	BR: {Name: "Brazil", Numeric: "076", Supported: false},
	IO: {Name: "British Indian Ocean Territory", Numeric: "086", Supported: false},
	BN: {Name: "Brunei Darussalam", Numeric: "096", Supported: false},
	BG: {Name: "Bulgaria", Numeric: "100", Supported: false},
	BF: {Name: "Burkina Faso", Numeric: "854", Supported: false},
	BI: {Name: "Burundi", Numeric: "108", Supported: false},
	CV: {Name: "Cabo Verde", Numeric: "132", Supported: false},
	KH: {Name: "Cambodia", Numeric: "116", Supported: false},
	CM: {Name: "Cameroon", Numeric: "120", Supported: false},
	CA: {Name: "Canada", Numeric: "124", Supported: false},
	KY: {Name: "Cayman Islands", Numeric: "136", Supported: false},
	CF: {Name: "Central African Republic", Numeric: "140", Supported: false},
	TD: {Name: "Chad", Numeric: "148", Supported: false},
	CL: {Name: "Chile", Numeric: "152", Supported: false},
	CN: {Name: "China", Numeric: "156", Supported: false},
	CX: {Name: "Christmas Island", Numeric: "162", Supported: false},
	CC: {Name: "Cocos (Keeling) Islands", Numeric: "166", Supported: false},
	CO: {Name: "Colombia", Numeric: "170", Supported: false},
	KM: {Name: "Comoros", Numeric: "174", Supported: false},
	CG: {Name: "Congo", Numeric: "178", Supported: false},
	CD: {Name: "Congo, the Democratic Republic of the", Numeric: "180", Supported: false},
	CK: {Name: "Cook Islands", Numeric: "184", Supported: false},
	CR: {Name: "Costa Rica", Numeric: "188", Supported: false},
	HR: {Name: "Croatia", Numeric: "191", Supported: false},
	CU: {Name: "Cuba", Numeric: "192", Supported: false},
	CW: {Name: "Curaçao", Numeric: "531", Supported: false},
	CY: {Name: "Cyprus", Numeric: "196", Supported: false},
	CZ: {Name: "Czech Republic", Numeric: "203", Supported: false},
	CI: {Name: "Côte d'Ivoire", Numeric: "384", Supported: false},
	DK: {Name: "Denmark", Numeric: "208", Supported: false},
	DJ: {Name: "Djibouti", Numeric: "262", Supported: false},
	DM: {Name: "Dominica", Numeric: "212", Supported: false},
	DO: {Name: "Dominican Republic", Numeric: "214", Supported: false},
	EC: {Name: "Ecuador", Numeric: "218", Supported: false},
	EG: {Name: "Egypt", Numeric: "818", Supported: false},
	SV: {Name: "El Salvador", Numeric: "222", Supported: false},
	GQ: {Name: "Equatorial Guinea", Numeric: "226", Supported: false},
	ER: {Name: "Eritrea", Numeric: "232", Supported: false},
	EE: {Name: "Estonia", Numeric: "233", Supported: false},
	SZ: {Name: "Eswatini", Numeric: "748", Supported: false},
	ET: {Name: "Ethiopia", Numeric: "231", Supported: false},
	FK: {Name: "Falkland Islands (Malvinas)", Numeric: "238", Supported: false},
	FO: {Name: "Faroe Islands", Numeric: "234", Supported: false},
	FJ: {Name: "Fiji", Numeric: "242", Supported: false},
	FI: {Name: "Finland", Numeric: "246", Supported: false},
	FR: {Name: "France", Numeric: "250", Supported: false},
	GF: {Name: "French Guiana", Numeric: "254", Supported: false},
	PF: {Name: "French Polynesia", Numeric: "258", Supported: false},
	TF: {Name: "French Southern Territories", Numeric: "260", Supported: false},
	GA: {Name: "Gabon", Numeric: "266", Supported: false},
	GM: {Name: "Gambia", Numeric: "270", Supported: false},
	GE: {Name: "Georgia", Numeric: "268", Supported: false},
	DE: {Name: "Germany", Numeric: "276", Supported: false},
	GH: {Name: "Ghana", Numeric: "288", Supported: false},
	GI: {Name: "Gibraltar", Numeric: "292", Supported: false},
	GR: {Name: "Greece", Numeric: "300", Supported: false},
	GL: {Name: "Greenland", Numeric: "304", Supported: false},
	GD: {Name: "Grenada", Numeric: "308", Supported: false},
	GP: {Name: "Guadeloupe", Numeric: "312", Supported: false},
	GU: {Name: "Guam", Numeric: "316", Supported: false},
	GT: {Name: "Guatemala", Numeric: "320", Supported: false},
	GG: {Name: "Guernsey", Numeric: "831", Supported: false},
	GN: {Name: "Guinea", Numeric: "324", Supported: false},
	GW: {Name: "Guinea-Bissau", Numeric: "624", Supported: false},
	GY: {Name: "Guyana", Numeric: "328", Supported: false},
	HT: {Name: "Haiti", Numeric: "332", Supported: false},
	HM: {Name: "Heard Island and McDonald Islands", Numeric: "334", Supported: false},
	VA: {Name: "Holy See", Numeric: "336", Supported: false},
	HN: {Name: "Honduras", Numeric: "340", Supported: false},
	HK: {Name: "Hong Kong", Numeric: "344", Supported: false},
	HU: {Name: "Hungary", Numeric: "348", Supported: false},
	IS: {Name: "Iceland", Numeric: "352", Supported: false},
	IN: {Name: "India", Numeric: "356", Supported: false},
	ID: {Name: "Indonesia", Numeric: "360", Supported: false},
	IR: {Name: "Iran, Islamic Republic of", Numeric: "364", Supported: false},
	IQ: {Name: "Iraq", Numeric: "368", Supported: false},
	IE: {Name: "Ireland", Numeric: "372", Supported: false},
	IM: {Name: "Isle of Man", Numeric: "833", Supported: false},
	IL: {Name: "Israel", Numeric: "376", Supported: false},
	IT: {Name: "Italy", Numeric: "380", Supported: false},
	JM: {Name: "Jamaica", Numeric: "388", Supported: false},
	JP: {Name: "Japan", Numeric: "392", Supported: false},
	JE: {Name: "Jersey", Numeric: "832", Supported: false},
	JO: {Name: "Jordan", Numeric: "400", Supported: false},
	KZ: {Name: "Kazakhstan", Numeric: "398", Supported: false},
	KE: {Name: "Kenya", Numeric: "404", Supported: false},
	KI: {Name: "Kiribati", Numeric: "296", Supported: false},
	KP: {Name: "Korea, Democratic People's Republic of", Numeric: "408", Supported: false},
	KR: {Name: "Korea, Republic of", Numeric: "410", Supported: false},
	XK: {Name: "Kosovo", Numeric: "383", Supported: false},
	KW: {Name: "Kuwait", Numeric: "414", Supported: false},
	KG: {Name: "Kyrgyzstan", Numeric: "417", Supported: false},
	LA: {Name: "Lao People's Democratic Republic", Numeric: "418", Supported: false},
	LV: {Name: "Latvia", Numeric: "428", Supported: false},
	LB: {Name: "Lebanon", Numeric: "422", Supported: false},
	LS: {Name: "Lesotho", Numeric: "426", Supported: false},
	LR: {Name: "Liberia", Numeric: "430", Supported: false},
	LY: {Name: "Libya", Numeric: "434", Supported: false},
	LI: {Name: "Liechtenstein", Numeric: "438", Supported: false},
	LT: {Name: "Lithuania", Numeric: "440", Supported: false},
	LU: {Name: "Luxembourg", Numeric: "442", Supported: false},
	MO: {Name: "Macao", Numeric: "446", Supported: false},
	MG: {Name: "Madagascar", Numeric: "450", Supported: false},
	MW: {Name: "Malawi", Numeric: "454", Supported: false},
	MY: {Name: "Malaysia", Numeric: "458", Supported: false},
	MV: {Name: "Maldives", Numeric: "462", Supported: false},
	ML: {Name: "Mali", Numeric: "466", Supported: false},
	MT: {Name: "Malta", Numeric: "470", Supported: false},
	MH: {Name: "Marshall Islands", Numeric: "584", Supported: false},
	MQ: {Name: "Martinique", Numeric: "474", Supported: false},
	MR: {Name: "Mauritania", Numeric: "478", Supported: false},
	MU: {Name: "Mauritius", Numeric: "480", Supported: false},
	YT: {Name: "Mayotte", Numeric: "175", Supported: false},
	MX: {Name: "Mexico", Numeric: "484", Supported: false},
	FM: {Name: "Micronesia, Federated States of", Numeric: "583", Supported: false},
	MD: {Name: "Moldova, Republic of", Numeric: "498", Supported: false},
	MC: {Name: "Monaco", Numeric: "492", Supported: false},
	MN: {Name: "Mongolia", Numeric: "496", Supported: false},
	ME: {Name: "Montenegro", Numeric: "499", Supported: false},
	MS: {Name: "Montserrat", Numeric: "500", Supported: false},
	MA: {Name: "Morocco", Numeric: "504", Supported: false},
	MZ: {Name: "Mozambique", Numeric: "508", Supported: false},
	MM: {Name: "Myanmar", Numeric: "104", Supported: false},
	NA: {Name: "Namibia", Numeric: "516", Supported: false},
	NR: {Name: "Nauru", Numeric: "520", Supported: false},
	NP: {Name: "Nepal", Numeric: "524", Supported: false},
	NL: {Name: "Netherlands", Numeric: "528", Supported: false},
	NC: {Name: "New Caledonia", Numeric: "540", Supported: false},
	NZ: {Name: "New Zealand", Numeric: "554", Supported: false},
	NI: {Name: "Nicaragua", Numeric: "558", Supported: false},
	NE: {Name: "Niger", Numeric: "562", Supported: false},
	NG: {Name: "Nigeria", Numeric: "566", Supported: false},
	NU: {Name: "Niue", Numeric: "570", Supported: false},
	NF: {Name: "Norfolk Island", Numeric: "574", Supported: false},
	MK: {Name: "North Macedonia", Numeric: "807", Supported: false},
	MP: {Name: "Northern Mariana Islands", Numeric: "580", Supported: false},
	NO: {Name: "Norway", Numeric: "578", Supported: false},
	OM: {Name: "Oman", Numeric: "512", Supported: false},
	PK: {Name: "Pakistan", Numeric: "586", Supported: false},
	PW: {Name: "Palau", Numeric: "585", Supported: false},
	PS: {Name: "Palestine, State of", Numeric: "275", Supported: false},
	PA: {Name: "Panama", Numeric: "591", Supported: false},
	PG: {Name: "Papua New Guinea", Numeric: "598", Supported: false},
	PY: {Name: "Paraguay", Numeric: "600", Supported: false},
	PE: {Name: "Peru", Numeric: "604", Supported: false},
	PH: {Name: "Philippines", Numeric: "608", Supported: false},
	PN: {Name: "Pitcairn", Numeric: "612", Supported: false},
	PL: {Name: "Poland", Numeric: "616", Supported: false},
	PT: {Name: "Portugal", Numeric: "620", Supported: false},
	PR: {Name: "Puerto Rico", Numeric: "630", Supported: false},
	QA: {Name: "Qatar", Numeric: "634", Supported: false},
	RO: {Name: "Romania", Numeric: "642", Supported: false},
	RU: {Name: "Russian Federation", Numeric: "643", Supported: false},
	RW: {Name: "Rwanda", Numeric: "646", Supported: false},
	RE: {Name: "Réunion", Numeric: "638", Supported: false},
	BL: {Name: "Saint Barthélemy", Numeric: "652", Supported: false},
	SH: {Name: "Saint Helena, Ascension and Tristan da Cunha", Numeric: "654", Supported: false},
	KN: {Name: "Saint Kitts and Nevis", Numeric: "659", Supported: false},
	LC: {Name: "Saint Lucia", Numeric: "662", Supported: false},
	MF: {Name: "Saint Martin (French part)", Numeric: "663", Supported: false},
	PM: {Name: "Saint Pierre and Miquelon", Numeric: "666", Supported: false},
	VC: {Name: "Saint Vincent and the Grenadines", Numeric: "670", Supported: false},
	WS: {Name: "Samoa", Numeric: "882", Supported: false},
	SM: {Name: "San Marino", Numeric: "674", Supported: false},
	ST: {Name: "Sao Tome and Principe", Numeric: "678", Supported: false},
	SA: {Name: "Saudi Arabia", Numeric: "682", Supported: false},
	SN: {Name: "Senegal", Numeric: "686", Supported: false},
	RS: {Name: "Serbia", Numeric: "688", Supported: false},
	SC: {Name: "Seychelles", Numeric: "690", Supported: false},
	SL: {Name: "Sierra Leone", Numeric: "694", Supported: false},
	SG: {Name: "Singapore", Numeric: "702", Supported: false},
	SX: {Name: "Sint Maarten (Dutch part)", Numeric: "534", Supported: false},
	SK: {Name: "Slovakia", Numeric: "703", Supported: false},
	SI: {Name: "Slovenia", Numeric: "705", Supported: false},
	SB: {Name: "Solomon Islands", Numeric: "090", Supported: false},
	SO: {Name: "Somalia", Numeric: "706", Supported: false},
	ZA: {Name: "South Africa", Numeric: "710", Supported: false},
	GS: {Name: "South Georgia and the South Sandwich Islands", Numeric: "239", Supported: false},
	SS: {Name: "South Sudan", Numeric: "728", Supported: false},
	ES: {Name: "Spain", Numeric: "724", Supported: false},
	LK: {Name: "Sri Lanka", Numeric: "144", Supported: false},
	SD: {Name: "Sudan", Numeric: "729", Supported: false},
	SR: {Name: "Suriname", Numeric: "740", Supported: false},
	SJ: {Name: "Svalbard and Jan Mayen", Numeric: "744", Supported: false},
	SE: {Name: "Sweden", Numeric: "752", Supported: false},
	CH: {Name: "Switzerland", Numeric: "756", Supported: false},
	SY: {Name: "Syrian Arab Republic", Numeric: "760", Supported: false},
	TW: {Name: "Taiwan, Province of China", Numeric: "158", Supported: false},
	TJ: {Name: "Tajikistan", Numeric: "762", Supported: false},
	TZ: {Name: "Tanzania, United Republic of", Numeric: "834", Supported: false},
	TH: {Name: "Thailand", Numeric: "764", Supported: false},
	TL: {Name: "Timor-Leste", Numeric: "626", Supported: false},
	TG: {Name: "Togo", Numeric: "768", Supported: false},
	TK: {Name: "Tokelau", Numeric: "772", Supported: false},
	TO: {Name: "Tonga", Numeric: "776", Supported: false},
	TT: {Name: "Trinidad and Tobago", Numeric: "780", Supported: false},
	TN: {Name: "Tunisia", Numeric: "788", Supported: false},
	TR: {Name: "Turkey", Numeric: "792", Supported: false},
	TM: {Name: "Turkmenistan", Numeric: "795", Supported: false},
	TC: {Name: "Turks and Caicos Islands", Numeric: "796", Supported: false},
	TV: {Name: "Tuvalu", Numeric: "798", Supported: false},
	UG: {Name: "Uganda", Numeric: "800", Supported: false},
	UA: {Name: "Ukraine", Numeric: "804", Supported: false},
	AE: {Name: "United Arab Emirates", Numeric: "784", Supported: false},
	UM: {Name: "United States Minor Outlying Islands", Numeric: "581", Supported: false},
	UY: {Name: "Uruguay", Numeric: "858", Supported: false},
	UZ: {Name: "Uzbekistan", Numeric: "860", Supported: false},
	VU: {Name: "Vanuatu", Numeric: "548", Supported: false},
	VE: {Name: "Venezuela (Bolivarian Republic of)", Numeric: "862", Supported: false},
	VN: {Name: "Viet Nam", Numeric: "704", Supported: false},
	VG: {Name: "Virgin Islands (British)", Numeric: "092", Supported: false},
	VI: {Name: "Virgin Islands (U.S.)", Numeric: "850", Supported: false},
	WF: {Name: "Wallis and Futuna", Numeric: "876", Supported: false},
	EH: {Name: "Western Sahara", Numeric: "732", Supported: false},
	YE: {Name: "Yemen", Numeric: "887", Supported: false},
	ZM: {Name: "Zambia", Numeric: "894", Supported: false},
	ZW: {Name: "Zimbabwe", Numeric: "716", Supported: false},
}

var States = map[Country]map[string]string{
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
		"NIR": "Norhtern Ireland",
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
	},
}
