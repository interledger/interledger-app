import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

/**
 * Feature 2
 * Active when carousel screen = 3.
 * TODO: Dark mode variant note: tablet dark mode uses Group 1010106674 (418:11479).
 * Requires designer confirmation whether to use structural variant.
 */
export function Feature2() {
  return (
    <FeatureSection
      screen={3}
      heading="Broad in reach. Local in feel."
      body="Settle transactions instantly with minimal fees, across any border."
      columnOrder="text-right"
      widget={
        <FeatureWidget
          avatar="JW"
          avatarColor="#7ab8e8"
          name="Jane W"
          amount="+ $120"
          note="dinner last night"
          timestamp="29.10.2026 12:30"
        />
      }
    />
  )
}
