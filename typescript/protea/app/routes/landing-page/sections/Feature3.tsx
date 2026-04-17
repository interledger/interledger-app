import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

import orbitSvg from "../orbit.svg"

/**
 * Feature 3
 * Active when carousel screen = 4.
 */
export function Feature3() {
  return (
    <FeatureSection
      screen={4}
      heading="One system. Many contexts."
      body="Ready to diverse environments and needs."
      columnOrder="text-left"
      widget={<img src={orbitSvg} alt="Orbit graphic" style={{ width: "80px", marginTop: "16px", display: "block", marginInline: "auto" }} />}
    />
  )
}
