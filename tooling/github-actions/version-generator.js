import { execSync } from "node:child_process";

let shouldPushTag = /** @type {boolean} */ (false);
let pushDockerImage = /** @type {boolean} */ (false);
let generateRelease = /** @type {boolean} */ (false);

/**
 * @param {string} event
 * @param {string} refName
 */
export function generateVersion(event, refName) {
  /** @type {string | undefined} */
  let version = undefined;

  if (event === "schedule") {
    pushDockerImage = true;
    version = "nightly";
  } else if (event === "workflow_dispatch") {
    version = `manual_${refName}`;
    pushDockerImage = true;
  } else if (refName.startsWith("release/v")) {
    pushDockerImage = true;
    shouldPushTag = true;
    generateRelease = true;

    const versionPrefix = refName.replace("release/", "");
    // Remove 'v' and split on '.' or '-' to extract version components
    const versionParts = versionPrefix.match(
      /^v?(\d+)\.(\d+)(?:\.(\d+))?(?:-([\w.-]+))?/,
    );

    let [, major, minor, patch = "0", preRelease] = versionParts || [];
    let versionSearch = `v${major}.${minor}.*`;
    if (preRelease) {
      versionSearch += `-${preRelease}`;
    }

    console.log(`VERSION_SEARCH: ${versionSearch}`);

    let latestTag = "";
    try {
      latestTag = execSync(
        `git tag -l "${versionSearch}" --sort=-taggerdate | head -n 1`,
      )
        .toString()
        .trim();
    } catch (error) {
      console.warn(
        "Failed to fetch latest tag:",
        /** @type {Error} */ (error).message,
      );
    }

    if (latestTag) {
      const parts = latestTag.match(/^v?(\d+)\.(\d+)\.(\d+)(?:-([\w.-]+))?/);
      if (parts) {
        patch = (Number(parts[3]) + 1).toString();
        preRelease = parts[4] || preRelease;
      }
    }

    version = `v${major}.${minor}.${patch}`;
    if (preRelease) {
      version += `-${preRelease}`;
    }
  } else {
    version = refName.replace(/[^a-zA-Z0-9_.-]/g, "_");

    if (!/^[a-zA-Z0-9]/.test(version)) {
      version = `tag_${version}`;
    }
  }

  return version;
}

/** @param {import('github-script').AsyncFunctionArguments} AsyncFunctionArguments */
export default async ({ core, context }) => {
  const eventName = context.eventName;
  const refName = context.ref.replace(/^refs\/(?:heads|tags|pull)\//, "");

  console.log(JSON.stringify({ eventName, refName }, null, 2));

  const version = generateVersion(eventName, refName);

  console.log(`New version will be: ${version}`);
  core.setOutput("NEW_VERSION", version);

  console.log(`Will tag be pushed? ${shouldPushTag}`);
  core.setOutput("SHOULD_PUSH_TAG", shouldPushTag);

  console.log(`Will docker image be pushed? ${pushDockerImage}`);
  core.setOutput("PUSH_DOCKER_IMAGE", pushDockerImage);

  console.log(`Will generate release noted? ${generateRelease}`);
  core.setOutput("GENERATE_RELEASE", generateRelease);

  if (shouldPushTag) {
    try {
      execSync(`git tag -fa ${version} -m "${version}"`, {
        stdio: "inherit",
      });
      execSync(`git push -f origin ${version}`, { stdio: "inherit" });
    } catch (error) {
      core.setFailed(
        `Failed to create or push tag: ${/** @type {Error} */ (error).message}`,
      );
    }
  }
};
