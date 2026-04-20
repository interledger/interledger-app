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
      heading="Designed to adopt. Easy to adopt."
      body="We are building the wallet in the open and want to participate from the start."
      columnOrder="text-left"
      visual={
        <div style={{ marginTop: "16px", display: "flex", justifyContent: "center" }}>
          <a href="/signup" className="nav-cta" style={{ display: "inline-flex" }}>
            Sign up now <span style={{ marginLeft: "4px" }}>&rarr;</span>
          </a>
        </div>
      }
    />
  )
}
