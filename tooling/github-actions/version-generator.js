let shouldPushTag = false;
let pushDockerImage = false;
let generateRelease = false;

/**
 * @param {string} event
 * @param {string} refName
 */
export function generateVersion(event, refName) {
  /** @type {string | undefined} }*/
  let newVersion = undefined;

  if (event === "schedule") {
    pushDockerImage = true;
    newVersion = "nightly";
  } else if (event === "workflow_dispatch") {
    newVersion = `manual_${refName}`;
    pushDockerImage = true;
  } else if (refName.startsWith("release/v")) {
  } else {
    newVersion = refName.replace(/[^a-zA-Z0-9_.-]/g, "_");

    if (!/^[a-zA-Z0-9]/.test(newVersion)) {
      newVersion = `tag_${newVersion}`;
    }
  }

  return {
    newVersion,
  };
}

/** @param {import('github-script').AsyncFunctionArguments} AsyncFunctionArguments */
export default async ({ core, context }) => {
  const eventName = context.eventName;
  const refName = context.ref.replace(/^refs\/(?:heads|tags|pull)\//, "");

  console.log(JSON.stringify({ eventName, refName }, null, 2));

  const { newVersion } = generateVersion(eventName, refName);

  console.log("newVersion", newVersion);

  core.setOutput("NEW_VERSION", newVersion);
  core.setOutput("SHOULD_PUSH_TAG", shouldPushTag);
  core.setOutput("PUSH_DOCKER_IMAGE", pushDockerImage);
  core.setOutput("GENERATE_RELEASE", generateRelease);
};
