import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

/**
 * Feature 1 — "Global by design. Inclusive by default."
 * Active when carousel screen = 2.
 */
export function Feature1() {
  return (
    <FeatureSection
      screen={2}
      heading="Global by design. Inclusive by default."
      body="Designed to meet people where they are, how they are."
      visual={
        <FeatureWidget
          avatar="MH"
          avatarColor="#e87a7a"
          name="Mike H"
          amount="+ $348"
          note="tnx for the adventure"
          timestamp="28.10.2026 21:57"
        />
      }
    />
  )
}
