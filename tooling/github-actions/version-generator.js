let shouldPushTag = false;
let pushDockerImage = false;
let generateRelease = false;

/**
 * @param {string} event
 * @param {string} ref
 */
export function generateVersion(event, ref) {
  /** @type {string | undefined} }*/
  let newVersion = undefined;

  if (event === "schedule") {
    pushDockerImage = true;
    newVersion = "nightly";
  } else if (event === "workflow_dispatch") {
    newVersion = `manual_{ref}`;
    pushDockerImage = true;
  } else if (ref.startsWith("release/v")) {
  } else {
    let newVersion = ref.replace(/[^a-zA-Z0-9_.-]/g, "_");

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
  const ref = context.ref;

  const { newVersion } = generateVersion(eventName, ref);

  core.setOutput("NEW_VERSION", newVersion);
};
