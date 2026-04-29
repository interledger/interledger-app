import { PageSection } from "../components/PageSection"

interface CardProps {
  text: string
  className: string
}

function OpenWorldCard({ text, className }: CardProps) {
  return (
    <article className={`open-world-card ${className}`}>
      <p className="text-body-lg-emphasis">{text}</p>
    </article>
  )
}

export function Ecosystem() {
  return (
    <PageSection className="ecosystem-section">
      <div className="ecosystem-container">
        <h2 className="ecosystem-title text-h2">Made for a connected open world</h2>
        
        <div className="ecosystem-grid">
          <OpenWorldCard 
            text="Builders exploring Interledger" 
            className="card-lavender" 
          />
          <OpenWorldCard 
            text="Developers testing open payments" 
            className="card-aqua" 
          />
          <OpenWorldCard 
            text="Organizations experimenting with interoperable value transfer" 
            className="card-pistachio" 
          />
          <OpenWorldCard 
            text="Anyone who believes payments should be open by default" 
            className="card-blush" 
          />
        </div>
      </div>
    </PageSection>
  )
}
