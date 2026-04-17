import { PageSection } from "../components/PageSection"
import { motion } from "framer-motion"

export function OtherFeatures() {
  return (
     <PageSection className="other-features-section">
      <motion.div 
        className="other-features-banner"
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ type: "spring", duration: 1, bounce: 0 }}
      >
        <header className="other-features-header">
          <div className="other-features-left">
            <h2 className="other-features-headline">
              Open standards.<br />
              Solid foundations.<br />
              Solid foundations.
            </h2>
          </div>
          <div className="other-features-right">
            <p className="other-features-description">
              This wallet is designed for a world where money moves across borders, systems, and contexts.
            </p>
            <p className="other-features-description">
              It works wherever the web works, without assuming who you are, where you live, or how you participate.
            </p>
          </div>
        </header>

        <div className="other-features-cta-block">
          <a href="#" className="other-features-btn">
            Get the Interledger Wallet
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
          </a>

          <div className="other-features-badges">
            <a href="#" className="app-badge app-badge--ios">
              <span className="app-badge__icon"></span>
              <div className="app-badge__text">
                <span className="app-badge__sub">Download on the</span>
                <span className="app-badge__main">App Store</span>
              </div>
            </a>
            <a href="#" className="app-badge app-badge--android">
              <span className="app-badge__icon">▶</span>
              <div className="app-badge__text">
                <span className="app-badge__sub">GET IT ON</span>
                <span className="app-badge__main">Google Play</span>
              </div>
            </a>
          </div>

          <p className="browser-availability">
            Also available in browsers
          </p>
        </div>
      </motion.div>
    </PageSection>
  )
}
