import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

/**
 * Feature 3
 * Active when carousel screen = 4.
 * TODO: Dark mode variant note: tablet dark mode uses Group 1010106675 (418:11526).
 * Requires designer confirmation whether to use structural variant.
 */
export function Feature3() {
  return (
    <FeatureSection
      screen={4}
      heading="One system. Endless possibilities."
      body="Your keys, your money. Manage everything securely from one place."
      columnOrder="text-left"
      widget={
        <FeatureWidget
          avatar="AR"
          avatarColor="#e8b87a"
          name="Alex R"
          amount="- $45"
          note="coffee beans"
          timestamp="30.10.2026 09:15"
        />
      }
    />
  )
}
