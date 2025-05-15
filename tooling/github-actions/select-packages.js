// @ts-check

import { PLATFORMS, GO_PACKAGES, TS_PACKAGES } from "./constants.js";

/**
 * @returns {Builds}
 */
export function generateBuilds() {
  /** @type {Builds} */
  const builds = {
    package: [...GO_PACKAGES, ...TS_PACKAGES],
    platform: PLATFORMS,
  };
  return builds;
}

/** @param {import('github-script').AsyncFunctionArguments} AsyncFunctionArguments */
export default async ({ core }) => {
  const builds = generateBuilds();

  /* TODO: Remove - debug only */ console.log(JSON.stringify(builds, null, 2));
  core.setOutput("builds", JSON.stringify(builds));
};
