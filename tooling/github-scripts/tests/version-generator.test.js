import { describe, it, mock } from "node:test";
import assert from "node:assert";
import childProcess from "node:child_process";

import { generateVersion } from "../version-generator.js";

describe("version-generator/generateVersion", () => {
  it("should not push docker image when falling back to default case", () => {
    const prNumber = 100;
    const expectedVersion = `${prNumber}_merge`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("merge", `refs/pull/${prNumber}/merge`);

    assert.strictEqual(
      version,
      expectedVersion,
      `Incorrect version. Version should be ${expectedVersion}.`,
    );
    assert.strictEqual(
      pushDockerImage,
      false,
      "Docker image should be puhsed.",
    );
    assert.strictEqual(shouldPushTag, false, "Tag should not be pushed.");
    assert.strictEqual(
      generateRelease,
      false,
      "Release should not be generated.",
    );
  });

  it("should push image version with version `nightly` when event is `schedule`", () => {
    const expectedVersion = "nightly";
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("schedule", "not_important");

    assert.strictEqual(
      version,
      expectedVersion,
      `Incorrect version. Version should be ${expectedVersion}.`,
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be pushed.");
    assert.strictEqual(shouldPushTag, false, "Tag should not be pushed.");
    assert.strictEqual(
      generateRelease,
      false,
      "Release should not be generated.",
    );
  });

  it("should push image with version matching `manual_<branch_name>` if event is `workflow_dispatcha", () => {
    const branch = "main";
    const expectedVersion = `manual_${branch}`;
    const ref = `refs/heads/${branch}`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("workflow_dispatch", ref);

    assert.strictEqual(
      version,
      expectedVersion,
      `Incorrect version. Version should be ${expectedVersion}.`,
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be puhsed.");
    assert.strictEqual(shouldPushTag, false, "Tag should not be pushed.");
    assert.strictEqual(
      generateRelease,
      false,
      "Release should not be generated.",
    );
  });

  it("should return version v1.0.0-pre when pushing to release/v1.0.0-pre, push image, push tag and generate release (pre-release) - no existing tag", () => {
    const branchVersion = "v1.0.0-pre";
    const expectedVersion = "v1.0.0-pre";

    mock.method(childProcess, "execSync", () => "");

    const branch = `release/${branchVersion}`;
    const ref = `refs/heads/${branch}`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("push", ref);

    assert.strictEqual(
      version,
      expectedVersion,
      "Incorrect version. Version should be nightly.",
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be puhsed.");
    assert.strictEqual(shouldPushTag, true, "Tag should be pushed.");
    assert.strictEqual(generateRelease, true, "Release should be generated.");
  });

  it("should return version v1.0.1-pre when pushing to branch `release/v1.0.0-pre`, push image, push new tag and generate release (pre-release) - existing tag", () => {
    const branchVersion = "v1.0.0-pre";
    const expectedVersion = "v1.0.1-pre";

    mock.method(childProcess, "execSync", () => "v1.0.0-pre");

    const branch = `release/${branchVersion}`;
    const ref = `refs/heads/${branch}`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("push", ref);

    assert.strictEqual(
      version,
      expectedVersion,
      "Incorrect version. Version should be nightly.",
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be puhsed.");
    assert.strictEqual(shouldPushTag, true, "Tag should be pushed.");
    assert.strictEqual(generateRelease, true, "Release should be generated.");
  });

  it("should return version v1.1.3 when pushing to branch the correct flags when pushing to release branch - existing tag", () => {
    const branchVersion = "v1.0.0";
    const expectedVersion = "v1.0.0";

    mock.method(childProcess, "execSync", () => "");

    const branch = `release/${branchVersion}`;
    const ref = `refs/heads/${branch}`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("push", ref);

    assert.strictEqual(
      version,
      expectedVersion,
      "Incorrect version. Version should be nightly.",
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be puhsed.");
    assert.strictEqual(shouldPushTag, true, "Tag should be pushed.");
    assert.strictEqual(generateRelease, true, "Release should be generated.");
  });

  it("should return version v1.1.2 and the correct flags when pushing to release branch - existing tag", () => {
    const branchVersion = "v1.1.0";
    const expectedVersion = "v1.1.3";

    mock.method(childProcess, "execSync", () => "v1.1.2\nv1.1.1\nv1.1.0");

    const branch = `release/${branchVersion}`;
    const ref = `refs/heads/${branch}`;
    const { version, shouldPushTag, pushDockerImage, generateRelease } =
      generateVersion("push", ref);

    assert.strictEqual(
      version,
      expectedVersion,
      "Incorrect version. Version should be nightly.",
    );
    assert.strictEqual(pushDockerImage, true, "Docker image should be puhsed.");
    assert.strictEqual(shouldPushTag, true, "Tag should be pushed.");
    assert.strictEqual(generateRelease, true, "Release should be generated.");
  });
});
