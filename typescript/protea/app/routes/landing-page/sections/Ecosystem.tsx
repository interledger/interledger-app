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
  const cards = [
    {
      text: "Builders exploring Interledger",
      className: "card-lavender"
    },
    {
      text: "Developers testing open payments",
      className: "card-aqua"
    },
    {
      text: "Organizations experimenting with interoperable value transfer",
      className: "card-pistachio"
    },
    {
      text: "Anyone who believes payments should be open by default",
      className: "card-blush"
    }
  ]

  return (
    <PageSection className="ecosystem-section">
      <div className="ecosystem-container">
        <h2 className="ecosystem-title text-h2">Made for a connected open world</h2>
        
        <div className="ecosystem-grid">
          {cards.map((card, index) => (
            <OpenWorldCard 
              key={index} 
              text={card.text} 
              className={card.className} 
            />
          ))}
        </div>
      </div>
    </PageSection>
  )
}
