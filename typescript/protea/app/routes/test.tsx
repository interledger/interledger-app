import React, { useEffect, useState } from 'react'
import { Form } from 'remix'
import { Button } from '~/components'
import { Autocomplete } from '~/components/Autocomplete'

type Country = {
  id: string
  name: string
}

export default function SignupPage() {
  const [country, setCountry] = useState<Country>()
  const [query, setSetQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState(countryListAlpha2)

  useEffect(() => {
    if (query === '') setFilteredCountries(countryListAlpha2)
    else {
      setFilteredCountries(
        countryListAlpha2.filter((country) => {
          return (
            country.name
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, '')) ||
            country.id
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, ''))
          )
        })
      )
    }
  }, [query])
  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      {/* Header */}
      <Form
        action={`/signup`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        {/* TODO: Replace with autocomplete */}
        <Autocomplete
          id='country'
          value={country}
          onChange={setCountry}
          onQuery={(event) => setSetQuery(event.target.value)}
          options={filteredCountries}
          label='Country'
        />
        <input value={String(country?.id)} name='country' type='hidden' />

        <Button disabled={country?.id == 'Blank'} type='submit'>
          Create account
        </Button>
      </Form>
      <div className='col-span-full flex h-10 flex-col bg-primary sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      <div className='col-span-full flex h-10 flex-col bg-container sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      <div className='col-span-full flex h-10 flex-col bg-container-primary sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      {/* Form */}
      <div className='col-span-full flex h-60 flex-col bg-yellow-300 sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      
    </main>
  )
}

// All simple ISO 3166 country data
// including: 2-character, 3-character, full country name and numeric id.
// Sorted alphabetical by country name (special characters on bottom)
const countryListAlpha2 = [
  { id: 'AF', name: 'Afghanistan' },
  { id: 'AL', name: 'Albania' },
  { id: 'DZ', name: 'Algeria' },
  { id: 'AS', name: 'American Samoa' },
  { id: 'AD', name: 'Andorra' },
  { id: 'AO', name: 'Angola' },
  { id: 'AI', name: 'Anguilla' },
  { id: 'AQ', name: 'Antarctica' },
  { id: 'AG', name: 'Antigua and Barbuda' },
  { id: 'AR', name: 'Argentina' },
  { id: 'AM', name: 'Armenia' },
  { id: 'AW', name: 'Aruba' },
  { id: 'AU', name: 'Australia' },
  { id: 'AT', name: 'Austria' },
  { id: 'AZ', name: 'Azerbaijan' },
  { id: 'BS', name: 'Bahamas (the)' },
  { id: 'BH', name: 'Bahrain' },
  { id: 'BD', name: 'Bangladesh' },
  { id: 'BB', name: 'Barbados' },
  { id: 'BY', name: 'Belarus' },
  { id: 'BE', name: 'Belgium' },
  { id: 'BZ', name: 'Belize' },
  { id: 'BJ', name: 'Benin' },
  { id: 'BM', name: 'Bermuda' },
  { id: 'BT', name: 'Bhutan' },
  {
    id: 'BO',

    name: 'Bolivia (Plurinational State of)'
  },
  {
    id: 'BQ',

    name: 'Bonaire, Sint Eustatius and Saba'
  },
  { id: 'BA', name: 'Bosnia and Herzegovina' },
  { id: 'BW', name: 'Botswana' },
  { id: 'BV', name: 'Bouvet Island' },
  { id: 'BR', name: 'Brazil' },
  {
    id: 'IO',

    name: 'British Indian Ocean Territory (the)'
  },
  { id: 'BN', name: 'Brunei Darussalam' },
  { id: 'BG', name: 'Bulgaria' },
  { id: 'BF', name: 'Burkina Faso' },
  { id: 'BI', name: 'Burundi' },
  { id: 'CV', name: 'Cabo Verde' },
  { id: 'KH', name: 'Cambodia' },
  { id: 'CM', name: 'Cameroon' },
  { id: 'CA', name: 'Canada' },
  { id: 'KY', name: 'Cayman Islands (the)' },
  {
    id: 'CF',

    name: 'Central African Republic (the)'
  },
  { id: 'TD', name: 'Chad' },
  { id: 'CL', name: 'Chile' },
  { id: 'CN', name: 'China' },
  { id: 'CX', name: 'Christmas Island' },
  {
    id: 'CC',

    name: 'Cocos (Keeling) Islands (the)'
  },
  { id: 'CO', name: 'Colombia' },
  { id: 'KM', name: 'Comoros (the)' },
  {
    id: 'CD',

    name: 'Congo (the Democratic Republic of the)'
  },
  { id: 'CG', name: 'Congo (the)' },
  { id: 'CK', name: 'Cook Islands (the)' },
  { id: 'CR', name: 'Costa Rica' },
  { id: 'HR', name: 'Croatia' },
  { id: 'CU', name: 'Cuba' },
  { id: 'CW', name: 'Curaçao' },
  { id: 'CY', name: 'Cyprus' },
  { id: 'CZ', name: 'Czechia' },
  { id: 'CI', name: "Côte d'Ivoire" },
  { id: 'DK', name: 'Denmark' },
  { id: 'DJ', name: 'Djibouti' },
  { id: 'DM', name: 'Dominica' },
  { id: 'DO', name: 'Dominican Republic (the)' },
  { id: 'EC', name: 'Ecuador' },
  { id: 'EG', name: 'Egypt' },
  { id: 'SV', name: 'El Salvador' },
  { id: 'GQ', name: 'Equatorial Guinea' },
  { id: 'ER', name: 'Eritrea' },
  { id: 'EE', name: 'Estonia' },
  { id: 'SZ', name: 'Eswatini' },
  { id: 'ET', name: 'Ethiopia' },
  {
    id: 'FK',

    name: 'Falkland Islands (the) [Malvinas]'
  },
  { id: 'FO', name: 'Faroe Islands (the)' },
  { id: 'FJ', name: 'Fiji' },
  { id: 'FI', name: 'Finland' },
  { id: 'FR', name: 'France' },
  { id: 'GF', name: 'French Guiana' },
  { id: 'PF', name: 'French Polynesia' },
  {
    id: 'TF',

    name: 'French Southern Territories (the)'
  },
  { id: 'GA', name: 'Gabon' },
  { id: 'GM', name: 'Gambia (the)' },
  { id: 'GE', name: 'Georgia' },
  { id: 'DE', name: 'Germany' },
  { id: 'GH', name: 'Ghana' },
  { id: 'GI', name: 'Gibraltar' },
  { id: 'GR', name: 'Greece' },
  { id: 'GL', name: 'Greenland' },
  { id: 'GD', name: 'Grenada' },
  { id: 'GP', name: 'Guadeloupe' },
  { id: 'GU', name: 'Guam' },
  { id: 'GT', name: 'Guatemala' },
  { id: 'GG', name: 'Guernsey' },
  { id: 'GN', name: 'Guinea' },
  { id: 'GW', name: 'Guinea-Bissau' },
  { id: 'GY', name: 'Guyana' },
  { id: 'HT', name: 'Haiti' },
  {
    id: 'HM',

    name: 'Heard Island and McDonald Islands'
  },
  { id: 'VA', name: 'Holy See (the)' },
  { id: 'HN', name: 'Honduras' },
  { id: 'HK', name: 'Hong Kong' },
  { id: 'HU', name: 'Hungary' },
  { id: 'IS', name: 'Iceland' },
  { id: 'IN', name: 'India' },
  { id: 'ID', name: 'Indonesia' },
  {
    id: 'IR',

    name: 'Iran (Islamic Republic of)'
  },
  { id: 'IQ', name: 'Iraq' },
  { id: 'IE', name: 'Ireland' },
  { id: 'IM', name: 'Isle of Man' },
  { id: 'IL', name: 'Israel' },
  { id: 'IT', name: 'Italy' },
  { id: 'JM', name: 'Jamaica' },
  { id: 'JP', name: 'Japan' },
  { id: 'JE', name: 'Jersey' },
  { id: 'JO', name: 'Jordan' },
  { id: 'KZ', name: 'Kazakhstan' },
  { id: 'KE', name: 'Kenya' },
  { id: 'KI', name: 'Kiribati' },
  {
    id: 'KP',

    name: "Korea (the Democratic People's Republic of)"
  },
  { id: 'KR', name: 'Korea (the Republic of)' },
  { id: 'KW', name: 'Kuwait' },
  { id: 'KG', name: 'Kyrgyzstan' },
  {
    id: 'LA',

    name: "Lao People's Democratic Republic (the)"
  },
  { id: 'LV', name: 'Latvia' },
  { id: 'LB', name: 'Lebanon' },
  { id: 'LS', name: 'Lesotho' },
  { id: 'LR', name: 'Liberia' },
  { id: 'LY', name: 'Libya' },
  { id: 'LI', name: 'Liechtenstein' },
  { id: 'LT', name: 'Lithuania' },
  { id: 'LU', name: 'Luxembourg' },
  { id: 'MO', name: 'Macao' },
  { id: 'MG', name: 'Madagascar' },
  { id: 'MW', name: 'Malawi' },
  { id: 'MY', name: 'Malaysia' },
  { id: 'MV', name: 'Maldives' },
  { id: 'ML', name: 'Mali' },
  { id: 'MT', name: 'Malta' },
  { id: 'MH', name: 'Marshall Islands (the)' },
  { id: 'MQ', name: 'Martinique' },
  { id: 'MR', name: 'Mauritania' },
  { id: 'MU', name: 'Mauritius' },
  { id: 'YT', name: 'Mayotte' },
  { id: 'MX', name: 'Mexico' },
  {
    id: 'FM',

    name: 'Micronesia (Federated States of)'
  },
  {
    id: 'MD',

    name: 'Moldova (the Republic of)'
  },
  { id: 'MC', name: 'Monaco' },
  { id: 'MN', name: 'Mongolia' },
  { id: 'ME', name: 'Montenegro' },
  { id: 'MS', name: 'Montserrat' },
  { id: 'MA', name: 'Morocco' },
  { id: 'MZ', name: 'Mozambique' },
  { id: 'MM', name: 'Myanmar' },
  { id: 'NA', name: 'Namibia' },
  { id: 'NR', name: 'Nauru' },
  { id: 'NP', name: 'Nepal' },
  { id: 'NL', name: 'Netherlands (the)' },
  { id: 'NC', name: 'New Caledonia' },
  { id: 'NZ', name: 'New Zealand' },
  { id: 'NI', name: 'Nicaragua' },
  { id: 'NE', name: 'Niger (the)' },
  { id: 'NG', name: 'Nigeria' },
  { id: 'NU', name: 'Niue' },
  { id: 'NF', name: 'Norfolk Island' },
  {
    id: 'MP',

    name: 'Northern Mariana Islands (the)'
  },
  { id: 'NO', name: 'Norway' },
  { id: 'OM', name: 'Oman' },
  { id: 'PK', name: 'Pakistan' },
  { id: 'PW', name: 'Palau' },
  { id: 'PS', name: 'Palestine, State of' },
  { id: 'PA', name: 'Panama' },
  { id: 'PG', name: 'Papua New Guinea' },
  { id: 'PY', name: 'Paraguay' },
  { id: 'PE', name: 'Peru' },
  { id: 'PH', name: 'Philippines (the)' },
  { id: 'PN', name: 'Pitcairn' },
  { id: 'PL', name: 'Poland' },
  { id: 'PT', name: 'Portugal' },
  { id: 'PR', name: 'Puerto Rico' },
  { id: 'QA', name: 'Qatar' },
  {
    id: 'MK',

    name: 'Republic of North Macedonia'
  },
  { id: 'RO', name: 'Romania' },
  { id: 'RU', name: 'Russian Federation (the)' },
  { id: 'RW', name: 'Rwanda' },
  { id: 'RE', name: 'Réunion' },
  { id: 'BL', name: 'Saint Barthélemy' },
  {
    id: 'SH',

    name: 'Saint Helena, Ascension and Tristan da Cunha'
  },
  { id: 'KN', name: 'Saint Kitts and Nevis' },
  { id: 'LC', name: 'Saint Lucia' },
  {
    id: 'MF',

    name: 'Saint Martin (French part)'
  },
  {
    id: 'PM',

    name: 'Saint Pierre and Miquelon'
  },
  {
    id: 'VC',

    name: 'Saint Vincent and the Grenadines'
  },
  { id: 'WS', name: 'Samoa' },
  { id: 'SM', name: 'San Marino' },
  { id: 'ST', name: 'Sao Tome and Principe' },
  { id: 'SA', name: 'Saudi Arabia' },
  { id: 'SN', name: 'Senegal' },
  { id: 'RS', name: 'Serbia' },
  { id: 'SC', name: 'Seychelles' },
  { id: 'SL', name: 'Sierra Leone' },
  { id: 'SG', name: 'Singapore' },
  {
    id: 'SX',

    name: 'Sint Maarten (Dutch part)'
  },
  { id: 'SK', name: 'Slovakia' },
  { id: 'SI', name: 'Slovenia' },
  { id: 'SB', name: 'Solomon Islands' },
  { id: 'SO', name: 'Somalia' },
  { id: 'ZA', name: 'South Africa' },
  {
    id: 'GS',

    name: 'South Georgia and the South Sandwich Islands'
  },
  { id: 'SS', name: 'South Sudan' },
  { id: 'ES', name: 'Spain' },
  { id: 'LK', name: 'Sri Lanka' },
  { id: 'SD', name: 'Sudan (the)' },
  { id: 'SR', name: 'Suriname' },
  { id: 'SJ', name: 'Svalbard and Jan Mayen' },
  { id: 'SE', name: 'Sweden' },
  { id: 'CH', name: 'Switzerland' },
  { id: 'SY', name: 'Syrian Arab Republic' },
  { id: 'TW', name: 'Taiwan' },
  { id: 'TJ', name: 'Tajikistan' },
  {
    id: 'TZ',

    name: 'Tanzania, United Republic of'
  },
  { id: 'TH', name: 'Thailand' },
  { id: 'TL', name: 'Timor-Leste' },
  { id: 'TG', name: 'Togo' },
  { id: 'TK', name: 'Tokelau' },
  { id: 'TO', name: 'Tonga' },
  { id: 'TT', name: 'Trinidad and Tobago' },
  { id: 'TN', name: 'Tunisia' },
  { id: 'TR', name: 'Turkey' },
  { id: 'TM', name: 'Turkmenistan' },
  {
    id: 'TC',

    name: 'Turks and Caicos Islands (the)'
  },
  { id: 'TV', name: 'Tuvalu' },
  { id: 'UG', name: 'Uganda' },
  { id: 'UA', name: 'Ukraine' },
  {
    id: 'AE',

    name: 'United Arab Emirates (the)'
  },
  {
    id: 'GB',

    name: 'United Kingdom of Great Britain and Northern Ireland (the)'
  },
  {
    id: 'UM',

    name: 'United States Minor Outlying Islands (the)'
  },
  {
    id: 'US',

    name: 'United States of America (the)'
  },
  { id: 'UY', name: 'Uruguay' },
  { id: 'UZ', name: 'Uzbekistan' },
  { id: 'VU', name: 'Vanuatu' },
  {
    id: 'VE',

    name: 'Venezuela (Bolivarian Republic of)'
  },
  { id: 'VN', name: 'Viet Nam' },
  { id: 'VG', name: 'Virgin Islands (British)' },
  { id: 'VI', name: 'Virgin Islands (U.S.)' },
  { id: 'WF', name: 'Wallis and Futuna' },
  { id: 'EH', name: 'Western Sahara' },
  { id: 'YE', name: 'Yemen' },
  { id: 'ZM', name: 'Zambia' },
  { id: 'ZW', name: 'Zimbabwe' },
  { id: 'AX', name: 'Åland Islands' }
]
