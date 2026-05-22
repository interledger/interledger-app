import { PageSection } from "../components/PageSection"
import buildingWithChangeInMind from "../assets/building-with-change-in-mind.png"

export function BuiltWithChange() {
  return (
    <PageSection className="built-with-change">
      <div className="built-with-change-header">
        <h2 className="text-h2">Build with change in mind</h2>
        <p className="text-h4">
          We believe anyone should be able to send and receive money freely, 
          across borders and currencies, without unnecessary barriers or extractive fees. 
          This wallet is our commitment to building that future in the open and over time.
        </p>
      </div>
      
      <div className="built-with-change-image-wrapper">
        <img 
          src={buildingWithChangeInMind}
          alt="Building the future of open payments"
          loading="lazy"
        />
      </div>
    </PageSection>
  )
}
