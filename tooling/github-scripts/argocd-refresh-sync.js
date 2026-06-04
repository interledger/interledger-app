import process from "node:process";
import { setTimeout as sleep } from "node:timers/promises";
import { pathToFileURL } from "node:url";

/**
 * @typedef {{
 *   name: string,
 *   labels: Record<string, string>
 * }} Application
 */

/**
 * @typedef {{
 *   healthStatus: string,
 *   syncStatus: string,
 *   operationPhase: string,
 * }} ApplicationStatus
 */

/**
 * @typedef {{
 *   listApplications: () => Promise<Application[]>,
 *   refreshApplication: (appName: string) => Promise<void>,
 *   syncApplication: (appName: string) => Promise<void>,
 *   getApplicationStatus: (appName: string) => Promise<ApplicationStatus>,
 * }} ArgocdApi
 */

/**
 * @param {string} selector
 * @returns {string[]}
 */
export function parseSelector(selector) {
  return selector
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
}

/**
 * @param {Record<string, string>} labels
 * @param {string} selector
 * @returns {boolean}
 */
export function matchesSelector(labels, selector) {
  const parts = parseSelector(selector);
  return parts.every((part) => {
    const equals = part.match(/^([^!=]+)=([^=]+)$/);
    if (equals) {
      const [, key, value] = equals;
      return (labels[key] ?? "") === value;
    }

    const notEquals = part.match(/^([^!=]+)!=(.+)$/);
    if (notEquals) {
      const [, key, value] = notEquals;
      return (labels[key] ?? "") !== value;
    }

    return false;
  });
}

/**
 * @param {Application[]} applications
 * @param {string} selector
 * @returns {string[]}
 */
export function selectApplications(applications, selector) {
  return applications
    .filter((app) => matchesSelector(app.labels, selector))
    .map((app) => app.name);
}

/**
 * @param {{
 *   selector: string,
 *   timeoutSeconds: number,
 *   pollIntervalSeconds: number,
 * }} options
 * @param {ArgocdApi} api
 * @param {(ms: number) => Promise<void>} [sleepFn]
 */
export async function runRefreshSync(options, api, sleepFn = sleep) {
  const applications = await api.listApplications();
  const selected = selectApplications(applications, options.selector);

  if (selected.length === 0) {
    throw new Error(`No Argo CD applications matched selector '${options.selector}'.`);
  }

  console.log(`Selected Argo CD applications: ${selected.join(" ")}`);

  // Phase 1: refresh all.
  for (const appName of selected) {
    console.log(`Refreshing Argo CD application '${appName}'...`);
    await api.refreshApplication(appName);
  }

  // Phase 2: sync all (single pass; do not spam sync).
  for (const appName of selected) {
    const status = await api.getApplicationStatus(appName);
    if (status.operationPhase === "Running") {
      console.log(`${appName}: operation already running; skipping explicit sync request.`);
      continue;
    }

    console.log(`Syncing Argo CD application '${appName}'...`);
    await api.syncApplication(appName);
  }

  // Phase 3: check status one app at a time with per-app timeout.
  for (const appName of selected) {
    console.log(`Waiting for '${appName}' to become Healthy and Synced...`);
    const startedAt = Date.now();

    while (Date.now() - startedAt < options.timeoutSeconds * 1000) {
      const status = await api.getApplicationStatus(appName);
      console.log(
        `${appName}: health=${status.healthStatus}, sync=${status.syncStatus}, operation=${status.operationPhase || "none"}`,
      );

      if (status.healthStatus === "Healthy" && status.syncStatus === "Synced") {
        console.log(`${appName} is Healthy and Synced.`);
        break;
      }

      await sleepFn(options.pollIntervalSeconds * 1000);
    }

    const finalStatus = await api.getApplicationStatus(appName);
    if (finalStatus.healthStatus !== "Healthy" || finalStatus.syncStatus !== "Synced") {
      throw new Error(`Timed out waiting for '${appName}' to become Healthy and Synced.`);
    }
  }

  return selected;
}

/**
 * @param {string} endpoint
 * @param {string} token
 * @param {string} clientId
 * @param {string} clientSecret
 * @returns {ArgocdApi}
 */
export function createArgocdApi(endpoint, token, clientId, clientSecret) {
  const argocdBase = endpoint.replace(/\/$/, "").startsWith("http")
    ? endpoint.replace(/\/$/, "")
    : `https://${endpoint.replace(/\/$/, "")}`;

  /**
   * @param {string} path
   * @param {RequestInit} [init]
   */
  const request = async (path, init = {}) => {
    const response = await fetch(`${argocdBase}${path}`, {
      redirect: "follow",
      ...init,
      headers: {
        "CF-Access-Client-Id": clientId,
        "CF-Access-Client-Secret": clientSecret,
        Authorization: `Bearer ${token}`,
        ...(init.headers || {}),
      },
    });

    const text = await response.text();
    if (!response.ok) {
      throw new Error(
        `Argo CD API request failed (${response.status} ${response.statusText}) at ${path}: ${text.slice(0, 400)}`,
      );
    }

    return text;
  };

  /**
   * @param {string} text
   * @param {string} path
   */
  const parseJson = (text, path) => {
    try {
      return JSON.parse(text);
    } catch {
      throw new Error(
        `Argo CD API returned non-JSON at ${path}: ${text.slice(0, 400)}`,
      );
    }
  };

  return {
    async listApplications() {
      const text = await request("/api/v1/applications", { method: "GET" });
      const data = parseJson(text, "/api/v1/applications");
      const items = Array.isArray(data.items) ? data.items : [];
      return items.map((item) => ({
        name: String(item?.metadata?.name || ""),
        labels:
          item?.metadata?.labels && typeof item.metadata.labels === "object"
            ? item.metadata.labels
            : {},
      }));
    },
    async refreshApplication(appName) {
      await request(`/api/v1/applications/${appName}?refresh=normal`, {
        method: "GET",
      });
    },
    async syncApplication(appName) {
      await request(`/api/v1/applications/${appName}/sync`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ prune: false, dryRun: false }),
      });
    },
    async getApplicationStatus(appName) {
      const text = await request(`/api/v1/applications/${appName}`, {
        method: "GET",
      });
      const data = parseJson(text, `/api/v1/applications/${appName}`);
      return {
        healthStatus: String(data?.status?.health?.status || "Unknown"),
        syncStatus: String(data?.status?.sync?.status || "Unknown"),
        operationPhase: String(data?.status?.operationState?.phase || ""),
      };
    },
  };
}

/**
 * @param {NodeJS.ProcessEnv} env
 */
function getRequiredConfig(env) {
  const required = [
    "ARGOCD_ENDPOINT",
    "ARGOCD_AUTH_TOKEN",
    "CF_ACCESS_CLIENT_ID",
    "CF_ACCESS_CLIENT_SECRET",
    "APPLICATION_SELECTOR",
  ];

  for (const key of required) {
    if (!env[key]) {
      throw new Error(`${key} is required. Check workflow environment/secrets.`);
    }
  }

  return {
    endpoint: String(env.ARGOCD_ENDPOINT),
    token: String(env.ARGOCD_AUTH_TOKEN),
    clientId: String(env.CF_ACCESS_CLIENT_ID),
    clientSecret: String(env.CF_ACCESS_CLIENT_SECRET),
    selector: String(env.APPLICATION_SELECTOR).replace(/\s+/g, ""),
    timeoutSeconds: Number(env.TIMEOUT_SECONDS || "300"),
    pollIntervalSeconds: Number(env.POLL_INTERVAL_SECONDS || "20"),
  };
}

async function main() {
  const config = getRequiredConfig(process.env);

  if (config.token) {
    console.log(`::add-mask::${config.token}`);
  }
  if (config.clientSecret) {
    console.log(`::add-mask::${config.clientSecret}`);
  }

  const api = createArgocdApi(
    config.endpoint,
    config.token,
    config.clientId,
    config.clientSecret,
  );

  const selected = await runRefreshSync(
    {
      selector: config.selector,
      timeoutSeconds: config.timeoutSeconds,
      pollIntervalSeconds: config.pollIntervalSeconds,
    },
    api,
  );

  if (process.env.GITHUB_STEP_SUMMARY) {
    const summary = [
      "### Argo CD deployment",
      "",
      `- Refreshed and synced applications matching: \`${config.selector}\``,
      `- Matched applications: \`${selected.join(" ")}\``,
      `- Endpoint: \`${config.endpoint}\``,
    ].join("\n");

    await import("node:fs/promises").then(({ appendFile }) =>
      appendFile(process.env.GITHUB_STEP_SUMMARY, `${summary}\n`),
    );
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch((error) => {
    console.error(`::error::${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  });
}
