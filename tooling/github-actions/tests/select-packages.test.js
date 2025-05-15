// @ts-check
import { describe, it } from "node:test";
import assert from "node:assert";

import { generateBuilds } from "../select-packages.js";

describe("select-packages/generateBuilds", () => {
  it("should return the correct builds struncture", () => {
    const builds = generateBuilds();

    assert.equal(
      Object.keys(builds).length,
      2,
      'The builds object has more than 2 properties. It should only contain: "package" and "platform"',
    );
    assert.ok(
      builds.package.length > 0,
      'The builds objects "package" property is empty',
    );
    assert.ok(
      builds.platform.length > 0,
      'The builds objects "package" property is empty',
    );
  });
});
