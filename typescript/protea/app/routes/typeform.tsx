import type { LoaderFunctionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const form = url.searchParams.get('form') || ''

  if (form == 'risk') return riskLoader(url.searchParams)

  throw json({}, 404)
}

// Example url:
// https://fynbos.app/typeform?form=risk&age=F&income=F&cash=E&tfsa=Y&contributions=D&maxed=Y&intention=C&reaction=D
export async function riskLoader(searchParams: URLSearchParams) {
  const redirectParams = new URLSearchParams()
  const ageParam = searchParams.get('age') || ''
  const incomeParam = searchParams.get('income') || ''
  const cashParam = searchParams.get('cash') || ''
  const intentionParam = searchParams.get('intention') || ''
  const reactionParam = searchParams.get('reaction') || ''

  // The searchParams relating to tfsa
  const tfsa = searchParams.get('tfsa') || ''
  const contributions = searchParams.get('contributions') || ''
  const maxed = searchParams.get('maxed') || ''

  let age = '',
    income = '',
    cash = '',
    intention = '',
    reaction = ''

  // A: 18-24
  // B: 24-30
  // C: 30-34
  // D: 35-40
  // E: 41-40
  // F: 50+
  switch (ageParam) {
    case 'A':
      age = '21'
      break
    case 'B':
      age = '27'
      break
    case 'C':
      age = '32'
      break
    case 'D':
      age = '37'
      break
    case 'E':
      age = '45'
      break
    case 'F':
      age = '50'
      break
  }

  // A: Up to 15k
  // B: 15k to 25k
  // C: 25k to 35k
  // D: 35k to 50k
  // E: 50k to 80k
  // F: 80k to 150k
  // G: 150+
  switch (incomeParam) {
    case 'A':
      income = '12500'
      break
    case 'B':
      income = '37500'
      break
    case 'C':
      income = '62500'
      break
    case 'D':
      income = '87500'
      break
    case 'E':
      income = '125000'
      break
    case 'F':
      income = '175000'
      break
    case 'G':
      income = '200000'
      break
  }

  // A: Up to 30k
  // B: 30k to 50k
  // C: 50k to 100k
  // D: 100k to 500k
  // E: 500k+
  switch (cashParam) {
    case 'A':
      cash = '25000'
      break
    case 'B':
      cash = '75000'
      break
    case 'C':
      cash = '175000'
      break
    case 'D':
      cash = '375000'
      break
    case 'E':
      cash = '500000'
      break
  }

  // A: Grow without risk
  // B: Grow as much as possible
  // C: The middle
  switch (intentionParam) {
    case 'A':
      intention = 'MIN_LOSS'
      break
    case 'B':
      intention = 'MAX_GAIN'
      break
    case 'C':
      intention = 'BOTH'
      break
  }

  // A: Sell everything
  // B: Do nothing
  // C: Sell some
  // D: Buy more
  switch (reactionParam) {
    case 'A':
      reaction = 'SELL_ALL'
      break
    case 'B':
      reaction = 'KEEP_ALL'
      break
    case 'C':
      reaction = 'SELL_SOME'
      break
    case 'D':
      reaction = 'BUY_MORE'
      break
  }

  const profile = dictionary.find(
    (val) =>
      val.age == age &&
      val.cash == cash &&
      val.intention == intention &&
      val.reaction == reaction &&
      val.income == income
  )

  if (!profile) throw json({}, 404)

  const score = Math.floor(Number(profile.score) / 2)

  redirectParams.append('profile', String(score))

  if (tfsa == 'Y') {
    if (contributions == 'D' || maxed == 'Y') {
      redirectParams.append('tfsa', 'maxed')
    } else {
      redirectParams.append('tfsa', 'some')
    }
  } else {
    redirectParams.append('tfsa', 'none')
  }

  if (2 * Number(income) > Number(cash))
    redirectParams.append('emergency', 'true')

  return redirect(`/wealth/risk?${redirectParams.toString()}`)
}

type Dictionary = {
  age: string
  income: string
  cash: string
  intention: string
  reaction: string
  score: string
}

const dictionary: Dictionary[] = [
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '21',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '27',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '32',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '37',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '45',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '10'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '45',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '45',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '4.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '2.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '5.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '5'
  },
  {
    age: '50',
    income: '12500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '5.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '5.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '12500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '12500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '12500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '5.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '5.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '2.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '5.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '37500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '37500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '37500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '37500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '9'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '5.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '5.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '62500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '9'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '62500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '5.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4.5'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '8'
  },
  {
    age: '50',
    income: '87500',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '125000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '8'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7.5'
  },
  {
    age: '50',
    income: '125000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '8'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '175000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '25000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '5.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '75000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '175000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '375000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'BUY_MORE',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'SELL_SOME',
    score: '6.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'BOTH',
    reaction: 'KEEP_ALL',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'BUY_MORE',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_ALL',
    score: '2.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'SELL_SOME',
    score: '3'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MIN_LOSS',
    reaction: 'KEEP_ALL',
    score: '3.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'BUY_MORE',
    score: '7.5'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_ALL',
    score: '4'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'SELL_SOME',
    score: '7'
  },
  {
    age: '50',
    income: '200000',
    cash: '500000',
    intention: 'MAX_GAIN',
    reaction: 'KEEP_ALL',
    score: '7.5'
  }
]
