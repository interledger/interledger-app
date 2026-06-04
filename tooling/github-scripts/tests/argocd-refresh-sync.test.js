import { describe, it } from "node:test";
import assert from "node:assert";

import {
  matchesSelector,
  runRefreshSync,
  selectApplications,
} from "../argocd-refresh-sync.js";

describe("argocd-refresh-sync/selectors", () => {
  it("matches equality and inequality selectors", () => {
    const labels = {
      environment: "wallet-dev1",
      team: "wallet",
    };

    assert.equal(matchesSelector(labels, "environment=wallet-dev1"), true);
    assert.equal(matchesSelector(labels, "environment!=wallet-sandbox"), true);
    assert.equal(matchesSelector(labels, "team=ops"), false);
    assert.equal(
      matchesSelector(labels, "environment=wallet-dev1,team=wallet"),
      true,
    );
  });

  it("selects application names from labels", () => {
    const selected = selectApplications(
      [
        {
          name: "dev1-a",
          labels: { environment: "wallet-dev1", team: "wallet" },
        },
        {
          name: "sandbox-a",
          labels: { environment: "wallet-sandbox", team: "wallet" },
        },
      ],
      "environment=wallet-dev1",
    );

    assert.deepEqual(selected, ["dev1-a"]);
  });
});

describe("argocd-refresh-sync/orchestration", () => {
  it("runs refresh-all, sync-all once, then checks status per app", async () => {
    /** @type {string[]} */
    const calls = [];

    /** @type {Record<string, number>} */
    const statusCalls = {
      a: 0,
      b: 0,
    };

    const api = {
      async listApplications() {
        calls.push("list");
        return [
          { name: "a", labels: { environment: "wallet-dev1" } },
          { name: "b", labels: { environment: "wallet-dev1" } },
        ];
      },
      async refreshApplication(appName) {
        calls.push(`refresh:${appName}`);
      },
      async syncApplication(appName) {
        calls.push(`sync:${appName}`);
      },
      async getApplicationStatus(appName) {
        calls.push(`status:${appName}`);
        statusCalls[appName] += 1;

        if (statusCalls[appName] <= 2) {
          return {
            healthStatus: "Progressing",
            syncStatus: "OutOfSync",
            operationPhase: "Running",
          };
        }

        return {
          healthStatus: "Healthy",
          syncStatus: "Synced",
          operationPhase: "Succeeded",
        };
      },
    };

    await runRefreshSync(
      {
        selector: "environment=wallet-dev1",
        timeoutSeconds: 300,
        pollIntervalSeconds: 20,
      },
      api,
      async () => {},
    );

    assert.equal(calls[0], "list");

    // Refresh phase must complete for all before sync starts.
    const lastRefreshIndex = Math.max(
      calls.indexOf("refresh:a"),
      calls.indexOf("refresh:b"),
    );
    const firstSyncIndex = Math.min(
      calls.indexOf("sync:a"),
      calls.indexOf("sync:b"),
    );
    assert.ok(lastRefreshIndex < firstSyncIndex);

    // Sync is single-pass, one call per app.
    assert.equal(calls.filter((x) => x === "sync:a").length, 1);
    assert.equal(calls.filter((x) => x === "sync:b").length, 1);
  });

  it("fails the run if an application does not converge in time", async () => {
    const api = {
      async listApplications() {
        return [{ name: "a", labels: { environment: "wallet-dev1" } }];
      },
      async refreshApplication() {},
      async syncApplication() {},
      async getApplicationStatus() {
        return {
          healthStatus: "Progressing",
          syncStatus: "OutOfSync",
          operationPhase: "Running",
        };
      },
    };

    await assert.rejects(
      () =>
        runRefreshSync(
          {
            selector: "environment=wallet-dev1",
            timeoutSeconds: 0,
            pollIntervalSeconds: 20,
          },
          api,
          async () => {},
        ),
      /Timed out waiting for 'a' to become Healthy and Synced/,
    );
  });
});
