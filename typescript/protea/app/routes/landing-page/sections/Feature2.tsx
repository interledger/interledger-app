import { FeatureSection } from "../components/FeatureSection"
import { FeatureWidget } from "../components/FeatureWidget"

import orbitSvg from "../orbit.svg"

/**
 * Feature 2
 * Active when carousel screen = 3.
 */
export function Feature2() {
  return (
    <FeatureSection
      screen={3}
      heading="Broad in reach. Borderless by design."
      body="Built broad, built borderless. Designed to work everywhere."
      columnOrder="text-left"
      widget={<img src={orbitSvg} alt="Orbit graphic" style={{ width: "80px", marginTop: "16px", display: "block", marginInline: "auto" }} />}
    />
  )
}
