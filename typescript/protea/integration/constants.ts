import type { BrowserContext } from '@playwright/test'

type PlaywrightCookie = Parameters<BrowserContext['addCookies']>['0'][0]

export const USER_COOKIE: PlaywrightCookie = {
  name: 'ory_kratos_session',
  value:
    'MTY5NjkzMTY4OHxnbmRvR3I3d2s3Q3NnR0U0NXZBaFdtMTVjOXh3VnI2NURpa0YxVnBmQTRvei0xOXVUM194ZlhrbkdoMGVrMDkyRUwxdm4wcEE3NW52Uk84b1NPb3FmSzZRS29mRjRNVHc4TVB3UVZUQndPcEhMQmwwRFN2MDNDTzRYMWE2amp2eFI1ekVjWFlvWkIyNWhWVFc2NllLTWFJUXEwQmtKVTJhalNxVHJ1bjhqLUUtZklsRlR1MzBYMFlwUTdWYktrLUhIcWdNcXFvcmdhVWNLeTY0dTJJU0c5dlRlTGxrRTFoUkVvSl8wdTM5UDV5SFpGc01xUXlWWFBNV3BXNnlhNHBFcmtLUlBkMlBTUFVVaXpQSndUejBYZz09fO9In1Qj5wWk8Yp2BBdAwm_ngVeUXijQH_68179WPBTG',
  domain: 'fynbos.test',
  path: '/',
  expires: 4853480178,
  httpOnly: true,
  secure: false,
  sameSite: 'Strict'
}
