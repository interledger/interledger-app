// organize-imports-ignore
// disables prettier-plugin-organize-imports for the whole file
// The CSS side-effect imports below are ordered intentionally for cascade
// precedence (tokens/variables first, then base, then components).
// Do NOT auto-sort them by running format — alphabetizing breaks the landing-page styling.
import { MotionConfig } from 'framer-motion'
import type { LinksFunction, LoaderFunctionArgs } from 'react-router'
import { redirect } from 'react-router'
import './css/colors.css'
import './css/tokens/typography.css'
import './css/typography.css'
import './css/tokens/spacing.css'
import './css/layout.css'
import './css/tokens/animations.css'
import './css/animations.css'
import './css/nav.css'
import './css/hero.css'
import './css/feature.css'
import './css/cards.css'
import './css/other-features.css'
import './css/send-receive.css'
import './css/ecosystem.css'
import './css/built-with-change.css'
import './css/our-ecosystem.css'
import './css/footer.css'
import { Layouts } from '~/components/Scaffold'
import { PhoneCarouselProvider } from './context/PhoneCarouselContext'
import { Nav } from './components/Nav'
import { HeroSection } from './components/HeroSection'
import { PhysicalCards } from './sections/PhysicalCards'
import { OtherFeatures } from './sections/OtherFeatures'
import { SendReceive } from './sections/SendReceive'
import { Ecosystem } from './sections/Ecosystem'
import { BuiltWithChange } from './sections/BuiltWithChange'
import { OurEcosystem } from './sections/OurEcosystem'
import { Footer } from './components/Footer'

export const handle = {
  layout: Layouts.LandingPage,
  scaffold: {
    header: {}
  }
}

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const url = new URL(request.url)
  const password = url.searchParams.get('landingPagePassword')

  if (password !== 'interledger2026') {
    return redirect('/')
  }
  return null
}

export const links: LinksFunction = () => [
  { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
  {
    rel: 'preconnect',
    href: 'https://fonts.gstatic.com',
    crossOrigin: 'anonymous'
  },
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Poppins:wght@400;600&family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap'
  }
]

export default function LandingPage() {
  return (
    // Honor OS Reduce Motion: transforms snap to target, opacity fades kept.
    <MotionConfig reducedMotion='user'>
      <PhoneCarouselProvider>
        <div className='landing-shell'>
          <Nav />
          <HeroSection />
          <PhysicalCards />
          <OtherFeatures />
          <SendReceive />
          <Ecosystem />
          <BuiltWithChange />
          <OurEcosystem />
          <Footer />
        </div>
      </PhoneCarouselProvider>
    </MotionConfig>
  )
}
