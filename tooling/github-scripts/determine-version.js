/**
 * Determines the Docker image version tag and whether to push it,
 * based on the GitHub Actions event that triggered the workflow.
 *
 * @param {string} eventName - github.event_name (e.g. "push", "workflow_dispatch", "pull_request")
 * @param {string} ref - github.ref (e.g. "refs/tags/v1.2.3", "refs/heads/main", "refs/pull/42/merge")
 * @returns {{ version: string, dockerPush: boolean }}
 */
export function determineVersion(eventName, ref) {
  const isTag = ref.startsWith("refs/tags/");
  const refName = ref.replace(/^refs\/(?:heads|tags|pull)\//, "");

  // Semantic-release pushes a vX.Y.Z tag to main after every release.
  // That tag push is the normal production path: use the tag as-is and push images.
  if (eventName === "push" && isTag) {
    return { version: refName, dockerPush: true };
  }

  // Manual dispatch lets engineers build and push an image from any branch
  // without going through a release (e.g. hotfix staging, one-off testing).
  if (eventName === "workflow_dispatch") {
    return { version: `manual_${sanitize(refName)}`, dockerPush: true };
  }

  // Pull-request and any other build: verify the images compile, never push.
  // The sanitisation + ref_ prefix ensures the result is a valid Docker tag segment.
  let version = sanitize(refName);
  if (!/^[a-zA-Z0-9]/.test(version)) {
    version = `ref_${version}`;
  }
  return { version, dockerPush: false };
}

/**
 * @param {string} s
 * @returns {string}
 */
function sanitize(s) {
  return s.replace(/[^a-zA-Z0-9_.-]/g, "_");
}

/** @param {import('github-script').AsyncFunctionArguments} AsyncFunctionArguments */
export default async ({ core, context }) => {
  const { version, dockerPush } = determineVersion(
    context.eventName,
    context.ref,
  );

  console.log(`Version: ${version}`);
  console.log(`Push docker image: ${dockerPush}`);

  core.setOutput("version", version);
  core.setOutput("docker_push", String(dockerPush));
};
