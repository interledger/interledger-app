import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

/**
 * Feature 4
 * Active when carousel screen = 5.
 */
export function Feature4() {
  return (
    <FeatureSection
      screen={5}
      heading="Designed to adapt. Easy to adopt."
      body="An API-first approach making integration seamless for anyone."
      columnOrder="text-right"
      widget={
        <FeatureWidget
          avatar="SK"
          avatarColor="#b87ae8"
          name="Sam K"
          amount="+ $850"
          note="freelance work"
          timestamp="01.11.2026 14:00"
        />
      }
    />
  )
}
