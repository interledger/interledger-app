import { PageSection } from "../components/PageSection"
import { Glow } from "../components/Glow"
import fiantLogo from "../assets/providers/fiant.png"
import gatehubLogo from "../assets/providers/gatehub.png"
import chimoneyLogo from "../assets/providers/chimoney.png"
import xagoLogo from "../assets/providers/xago.png"

interface PartnerProps {
  name: string
  region: string
  logo: string
}

function ShoutOutCard({ name, region, logo }: PartnerProps) {
  return (
    <div className="shout-out-card">
      <div className="partner-logo">
        <img src={logo} alt={`${name} logo`} />
      </div>
      <div className="partner-info">
        <p className="text-body-sm-standard">{region}</p>
      </div>
    </div>
  )
}

export function OurEcosystem() {
  const partners = [
    { name: "FIANT", region: "USA", logo: fiantLogo },
    { name: "GATEHUB", region: "EU", logo: gatehubLogo },
    { name: "CHIMONEY", region: "Canada", logo: chimoneyLogo },
    { name: "XAGO", region: "South Africa", logo: xagoLogo }
  ]

  return (
    <PageSection className="our-ecosystem-section">
      <div className="our-ecosystem-header">
        <h2 className="text-h2">Our ecosystem</h2>
        <p className="text-body-lg-standard">
          Unlock the potential of Open Payments and Web Monetization 
          through the Interledger Wallet and help drive the evolution 
          of digital finances.
        </p>
      </div>

      <div className="partners-grid">
        {partners.map((p, i) => (
          <ShoutOutCard key={i} name={p.name} logo={p.logo} region={p.region} />
        ))}
      </div>

      <div className="ecosystem-cta">
        <Glow scrollTransform={{ scale: 1, rotate: 0, y: 0, opacity: 0.3 }} x="50%" y="20px" className="cta-glow-bg" />
        <h3 className="text-h2">You know a potential partner?</h3>
        <button className="cta-button-primary">
          Contact us
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ marginLeft: '8px' }}>
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
            <polyline points="15 3 21 3 21 9" />
            <line x1="10" y1="14" x2="21" y2="3" />
          </svg>
        </button>
      </div>
    </PageSection>
  )
}
