import { describe, it } from "node:test";
import assert from "node:assert";

import { determineVersion } from "../determine-version.js";

describe("determine-version/determineVersion", () => {
  it("uses the tag verbatim and pushes images on a version tag push", () => {
    const { version, dockerPush } = determineVersion(
      "push",
      "refs/tags/v1.2.3",
    );

    assert.strictEqual(version, "v1.2.3");
    assert.strictEqual(dockerPush, true);
  });

  it("uses release tag_name and pushes images on release events", () => {
    const { version, dockerPush } = determineVersion(
      "release",
      "refs/tags/v1.2.3",
      { release: { tag_name: "v1.2.3" } },
    );

    assert.strictEqual(version, "v1.2.3");
    assert.strictEqual(dockerPush, true);
  });

  it("uses workflow_call release_tag input when provided", () => {
    const { version, dockerPush } = determineVersion(
      "workflow_call",
      "refs/heads/main",
      { inputs: { release_tag: "v1.2.3" } },
    );

    assert.strictEqual(version, "v1.2.3");
    assert.strictEqual(dockerPush, true);
  });

  it("requires release_tag input on workflow_call", () => {
    assert.throws(
      () => determineVersion("workflow_call", "refs/heads/main"),
      /workflow_call publish runs require inputs\.release_tag/,
    );
  });

  it("does not push and uses the sanitised PR ref on pull_request", () => {
    const { version, dockerPush } = determineVersion(
      "pull_request",
      "refs/pull/42/merge",
    );

    assert.strictEqual(version, "42_merge");
    assert.strictEqual(dockerPush, false);
  });

  it("prefixes ref_ when a non-dispatch sanitised name starts with a non-alphanumeric character", () => {
    const { version, dockerPush } = determineVersion(
      "push",
      "refs/heads/.weird",
    );

    assert.strictEqual(version, "ref_.weird");
    assert.strictEqual(dockerPush, false);
  });

  it("does not push on a regular branch push (semantic-release flow runs in release.yml)", () => {
    const { version, dockerPush } = determineVersion("push", "refs/heads/main");

    assert.strictEqual(version, "main");
    assert.strictEqual(dockerPush, false);
  });
});
