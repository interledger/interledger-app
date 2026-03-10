/**
 * E2E Runner Scaler — pure logic module.
 *
 * Determines scaling actions (create / delete spot VMs, cleanup offline runners)
 * based on the current state of the GitHub Actions queue and GCP instances.
 *
 * All side-effects are delegated to injected adapters so the logic is fully testable.
 *
 * @module e2e-runner-scaler
 */

/** @import { ScalerConfig, QueueState, SpotState, RunnerInfo, ScaleDecision, GcpAdapter, GitHubAdapter } from './types' */

/** @type {ScalerConfig} */
export const DEFAULT_CONFIG = {
  maxSpots: 5,
  queueWaitThresholdSecs: 30,
  spotPrefix: "e2e-spot",
  runnerLabel: "e2e-tester-dynamic",
  machineImage: "e2e-tester-20260310",
  machineType: "n2-standard-4",
  minSpots: 0,
};

/**
 * Decide how many spot instances to create.
 *
 * @param {QueueState} queue
 * @param {SpotState}  spots
 * @param {ScalerConfig} config
 * @returns {number} Number of instances to create (0 = no scale-up needed).
 */
export function decideScaleUp(queue, spots, config) {
  if (queue.pendingCount < 1) return 0;
  if (queue.oldestWaitSecs <= config.queueWaitThresholdSecs) return 0;

  const available = config.maxSpots - spots.total;
  if (available <= 0) return 0;

  return Math.min(queue.pendingCount, available);
}

/**
 * Select spot VM names to delete during scale-down.
 * Keeps {@link config.minSpots} idle runners alive (default 0).
 *
 * Only considers runners that are online but NOT busy.
 *
 * @param {QueueState}    queue
 * @param {RunnerInfo[]}  runners  - Current GitHub-registered runners matching our prefix.
 * @param {ScalerConfig}  config
 * @returns {string[]} VM names to delete.
 */
export function decideScaleDown(queue, runners, config) {
  if (queue.pendingCount > 0) return [];

  const idleRunners = runners
    .filter(
      (r) =>
        r.name.startsWith(config.spotPrefix + "-") &&
        r.status === "online" &&
        !r.busy,
    )
    .map((r) => r.name);

  if (idleRunners.length <= config.minSpots) return [];

  // Keep minSpots alive, delete the rest
  return idleRunners.slice(config.minSpots);
}

/**
 * Select offline runners to remove from GitHub.
 *
 * @param {RunnerInfo[]}  runners
 * @param {ScalerConfig}  config
 * @returns {RunnerInfo[]} Runners to deregister.
 */
export function decideCleanup(runners, config) {
  return runners.filter(
    (r) =>
      r.name.startsWith(config.spotPrefix + "-") && r.status === "offline",
  );
}

/**
 * Determine whether we need to create instances to meet the minimum pool size.
 *
 * @param {SpotState}     spots
 * @param {number}        pendingDeletions - Number of VMs about to be deleted.
 * @param {ScalerConfig}  config
 * @returns {number} Number of instances to create to reach minSpots.
 */
export function decideEnsureMinimum(spots, pendingDeletions, config) {
  const afterDeletion = spots.total - pendingDeletions;
  if (afterDeletion >= config.minSpots) return 0;
  return config.minSpots - afterDeletion;
}

/**
 * Build the full scaling decision from current state.
 *
 * @param {QueueState}    queue
 * @param {SpotState}     spots
 * @param {RunnerInfo[]}  runners
 * @param {ScalerConfig}  config
 * @returns {ScaleDecision}
 */
export function buildDecision(queue, spots, runners, config) {
  const toCreate = decideScaleUp(queue, spots, config);
  const toDelete = decideScaleDown(queue, runners, config);
  const toCleanup = decideCleanup(runners, config);

  // Account for scale-up creations when checking minimum:
  // if we're already creating instances for demand, they count toward the pool.
  const effectiveTotal = spots.total + toCreate;
  const afterDeletion = effectiveTotal - toDelete.length;
  const toEnsureMin = afterDeletion >= config.minSpots
    ? 0
    : config.minSpots - afterDeletion;

  return {
    instancesToCreate: toCreate + toEnsureMin,
    instancesToDelete: toDelete,
    runnersToCleanup: toCleanup,
  };
}

// ---------------------------------------------------------------------------
// Side-effectful execution (used by the workflow, delegates to adapters)
// ---------------------------------------------------------------------------

/**
 * Generate a unique spot instance name.
 *
 * @param {string} prefix
 * @returns {string}
 */
export function generateSpotName(prefix) {
  const ts = Math.floor(Date.now() / 1000);
  const rand = Math.floor(Math.random() * 100000);
  return `${prefix}-${ts}-${rand}`;
}

/**
 * Execute the scaling decision using real adapters.
 *
 * @param {ScaleDecision} decision
 * @param {GcpAdapter}    gcp
 * @param {GitHubAdapter} gh
 * @param {ScalerConfig}  config
 * @param {(msg: string) => void} log
 */
export async function executeDecision(decision, gcp, gh, config, log) {
  // 1. Scale up
  if (decision.instancesToCreate > 0) {
    log(
      `⬆️  Creating ${decision.instancesToCreate} spot instance(s)...`,
    );
    const token = await gh.getRegistrationToken();

    for (let i = 0; i < decision.instancesToCreate; i++) {
      const name = generateSpotName(config.spotPrefix);
      log(`  🚀 Creating: ${name}`);
      await gcp.createSpotInstance(name, {
        machineImage: config.machineImage,
        machineType: config.machineType,
        metadata: {
          "runner-token": token,
          "runner-name": name,
          "runner-labels": config.runnerLabel,
        },
      });
    }
  }

  // 2. Scale down
  for (const name of decision.instancesToDelete) {
    log(`  🗑️  Deleting instance: ${name}`);
    await gcp.deleteInstance(name);
  }

  // 3. Cleanup offline runners
  for (const runner of decision.runnersToCleanup) {
    log(
      `  🧹 Removing offline runner: ${runner.name} (ID: ${runner.id})`,
    );
    await gh.removeRunner(runner.id);
  }
}

// ---------------------------------------------------------------------------
// Default export for `actions/github-script`
// ---------------------------------------------------------------------------

/** @param {import('github-script').AsyncFunctionArguments} AsyncFunctionArguments */
export default async ({ core, exec }) => {
  const config = {
    ...DEFAULT_CONFIG,
    machineImage:
      process.env.MACHINE_IMAGE ?? DEFAULT_CONFIG.machineImage,
    machineType:
      process.env.MACHINE_TYPE ?? DEFAULT_CONFIG.machineType,
    maxSpots: Number(process.env.MAX_SPOTS ?? DEFAULT_CONFIG.maxSpots),
    minSpots: Number(process.env.MIN_SPOTS ?? DEFAULT_CONFIG.minSpots),
    spotPrefix: process.env.SPOT_PREFIX ?? DEFAULT_CONFIG.spotPrefix,
    runnerLabel: process.env.RUNNER_LABEL ?? DEFAULT_CONFIG.runnerLabel,
  };

  const projectId = requireEnv("PROJECT_ID");
  const zone = requireEnv("ZONE");
  const runnerAdminToken = requireEnv("RUNNER_ADMIN_TOKEN");
  const repo = requireEnv("GITHUB_REPOSITORY");
  const ghToken = requireEnv("GH_TOKEN");

  const log = /** @param {string} msg */ (msg) => core.info(msg);

  // -- Collect queue state --------------------------------------------------
  log("📊 Collecting queue state...");
  const queue = await collectQueueState(exec, repo);
  log(`   Pending: ${queue.pendingCount}, oldest wait: ${queue.oldestWaitSecs}s`);

  // -- Collect GCP instance state -------------------------------------------
  log("📊 Collecting spot instance state...");
  const spots = await collectSpotState(exec, projectId, config.spotPrefix);
  log(`   Running: ${spots.running}, booting: ${spots.booting}, total: ${spots.total}`);

  // -- Collect GitHub runner state ------------------------------------------
  log("📊 Collecting runner state...");
  const runners = await collectRunnerState(runnerAdminToken, repo);
  const ourRunners = runners.filter((r) => r.name.startsWith(config.spotPrefix + "-"));
  log(`   Our runners: ${ourRunners.length} (online: ${ourRunners.filter((r) => r.status === "online").length})`);

  // -- Build and execute decision -------------------------------------------
  const decision = buildDecision(queue, spots, ourRunners, config);
  log(`📋 Decision: create=${decision.instancesToCreate}, delete=${decision.instancesToDelete.length}, cleanup=${decision.runnersToCleanup.length}`);

  /** @type {GcpAdapter} */
  const gcp = {
    createSpotInstance: async (name, opts) => {
      const metadataStr = Object.entries(opts.metadata)
        .map(([k, v]) => `${k}=${v}`)
        .join(",");
      await execCommand(exec, [
        "gcloud", "compute", "instances", "create", name,
        "--source-machine-image", opts.machineImage,
        "--zone", zone,
        "--project", projectId,
        "--machine-type", opts.machineType,
        "--provisioning-model", "SPOT",
        "--instance-termination-action", "DELETE",
        "--maintenance-policy", "TERMINATE",
        "--no-restart-on-failure",
        "--metadata", metadataStr,
      ]);
    },
    deleteInstance: async (name) => {
      await execCommand(exec, [
        "gcloud", "compute", "instances", "delete", name,
        "--zone", zone,
        "--project", projectId,
        "--quiet",
      ]);
    },
  };

  /** @type {GitHubAdapter} */
  const gh = {
    getRegistrationToken: async () => {
      const resp = await fetch(
        `https://api.github.com/repos/${repo}/actions/runners/registration-token`,
        {
          method: "POST",
          headers: {
            Authorization: `token ${runnerAdminToken}`,
            Accept: "application/vnd.github+json",
          },
        },
      );
      if (!resp.ok) throw new Error(`Failed to get registration token: ${resp.status}`);
      const data = /** @type {{ token: string }} */ (await resp.json());
      return data.token;
    },
    removeRunner: async (id) => {
      await fetch(
        `https://api.github.com/repos/${repo}/actions/runners/${id}`,
        {
          method: "DELETE",
          headers: {
            Authorization: `token ${runnerAdminToken}`,
            Accept: "application/vnd.github+json",
          },
        },
      );
    },
  };

  await executeDecision(decision, gcp, gh, config, log);
  log("✅ Scaler run complete.");
};

// ---------------------------------------------------------------------------
// Helpers for the default export (data collection)
// ---------------------------------------------------------------------------

/**
 * @param {string} name
 * @returns {string}
 */
function requireEnv(name) {
  const val = process.env[name];
  if (!val) throw new Error(`Missing required environment variable: ${name}`);
  return val;
}

/**
 * Run a command and return stdout. Throws on non-zero exit code.
 *
 * @param {import('github-script').AsyncFunctionArguments['exec']} exec
 * @param {string[]} args
 * @returns {Promise<string>}
 */
async function execCommand(exec, args) {
  let stdout = "";
  let stderr = "";
  const exitCode = await exec.exec(args[0], args.slice(1), {
    listeners: {
      stdout: (/** @type {Buffer} */ data) => { stdout += data.toString(); },
      stderr: (/** @type {Buffer} */ data) => { stderr += data.toString(); },
    },
    silent: false,
    ignoreReturnCode: true,
  });
  if (exitCode !== 0) {
    throw new Error(`Command failed (exit ${exitCode}): ${args.join(" ")}\n${stderr}`);
  }
  return stdout.trim();
}

/**
 * Collect queue state from `gh` CLI.
 *
 * @param {import('github-script').AsyncFunctionArguments['exec']} exec
 * @param {string} repo
 * @returns {Promise<QueueState>}
 */
async function collectQueueState(exec, repo) {
  /** @param {string} status */
  const listRuns = async (status) => {
    const raw = await execCommand(exec, [
      "gh", "run", "list",
      "--repo", repo,
      "--workflow", "e2e-tests.yml",
      "--status", status,
      "--limit", "100",
      "--json", "createdAt",
    ]);
    return raw ? JSON.parse(raw) : [];
  };

  const [queued, waiting] = await Promise.all([
    listRuns("queued"),
    listRuns("waiting"),
  ]);

  const all = [...queued, ...waiting];
  const pendingCount = all.length;
  let oldestWaitSecs = 0;

  if (pendingCount > 0) {
    const oldest = all
      .map((r) => new Date(r.createdAt).getTime())
      .sort((a, b) => a - b)[0];
    oldestWaitSecs = Math.floor((Date.now() - oldest) / 1000);
  }

  return { pendingCount, oldestWaitSecs };
}

/**
 * Collect spot instance state from GCP.
 *
 * @param {import('github-script').AsyncFunctionArguments['exec']} exec
 * @param {string} projectId
 * @param {string} prefix
 * @returns {Promise<SpotState>}
 */
async function collectSpotState(exec, projectId, prefix) {
  /** @param {string} filter */
  const count = async (filter) => {
    const raw = await execCommand(exec, [
      "gcloud", "compute", "instances", "list",
      "--project", projectId,
      "--filter", filter,
      "--format", "value(name)",
    ]);
    return raw ? raw.split("\n").filter(Boolean).length : 0;
  };

  const [running, booting] = await Promise.all([
    count(`name~'^${prefix}-' AND status=RUNNING`),
    count(`name~'^${prefix}-' AND (status=STAGING OR status=PROVISIONING)`),
  ]);

  return { running, booting, total: running + booting };
}

/**
 * Collect runner state from GitHub API.
 *
 * @param {string} token
 * @param {string} repo
 * @returns {Promise<RunnerInfo[]>}
 */
async function collectRunnerState(token, repo) {
  const resp = await fetch(
    `https://api.github.com/repos/${repo}/actions/runners?per_page=100`,
    {
      headers: {
        Authorization: `token ${token}`,
        Accept: "application/vnd.github+json",
      },
    },
  );
  if (!resp.ok) return [];
  const data = /** @type {{ runners: RunnerInfo[] }} */ (await resp.json());
  return data.runners ?? [];
}
