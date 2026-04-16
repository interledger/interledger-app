import { PageSection } from "../components/PageSection"

export function OtherFeatures() {
  return (
    <PageSection className="other-features-section">
      <div className="other-features-banner">
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
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
          </a>

          <div className="other-features-badges">
            {/* TODO: Replace with real badge assets later */}
            <div className="app-badge-placeholder" style={{ backgroundColor: 'black', color: 'white', borderRadius: '4px', padding: '8px 16px', fontSize: '12px', fontWeight: 'bold' }}>
              App Store
            </div>
            <div className="app-badge-placeholder" style={{ backgroundColor: 'black', color: 'white', borderRadius: '4px', padding: '8px 16px', fontSize: '12px', fontWeight: 'bold' }}>
              Google Play
            </div>
          </div>

          <p className="browser-availability">
            Also available in browsers
          </p>
        </div>
      </div>
    </PageSection>
  )
}
