import { describe, it } from "node:test";
import assert from "node:assert";

import {
  getRequiredConfig,
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
  it("refreshes all, avoids sync spam when operation is already running, then checks status per app", async () => {
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

    // Refresh phase must run for all selected applications.
    assert.equal(calls.filter((x) => x === "refresh:a").length, 1);
    assert.equal(calls.filter((x) => x === "refresh:b").length, 1);

    // With operationPhase=Running on the pre-sync status check,
    // explicit sync calls are intentionally skipped to avoid spamming sync.
    assert.equal(calls.filter((x) => x === "sync:a").length, 0);
    assert.equal(calls.filter((x) => x === "sync:b").length, 0);
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

describe("argocd-refresh-sync/config", () => {
  it("uses defaults when timeout/poll vars are absent", () => {
    const config = getRequiredConfig({
      ARGOCD_ENDPOINT: "https://sandbox-argo.interledger.tech",
      ARGOCD_AUTH_TOKEN: "token",
      CF_ACCESS_CLIENT_ID: "id",
      CF_ACCESS_CLIENT_SECRET: "secret",
      APPLICATION_SELECTOR: "environment=wallet-dev1",
    });

    assert.equal(config.timeoutSeconds, 300);
    assert.equal(config.pollIntervalSeconds, 20);
  });

  it("fails fast for invalid timeout/poll values", () => {
    assert.throws(
      () =>
        getRequiredConfig({
          ARGOCD_ENDPOINT: "https://sandbox-argo.interledger.tech",
          ARGOCD_AUTH_TOKEN: "token",
          CF_ACCESS_CLIENT_ID: "id",
          CF_ACCESS_CLIENT_SECRET: "secret",
          APPLICATION_SELECTOR: "environment=wallet-dev1",
          TIMEOUT_SECONDS: "not-a-number",
          POLL_INTERVAL_SECONDS: "20",
        }),
      /TIMEOUT_SECONDS must be a finite number greater than 0/,
    );

    assert.throws(
      () =>
        getRequiredConfig({
          ARGOCD_ENDPOINT: "https://sandbox-argo.interledger.tech",
          ARGOCD_AUTH_TOKEN: "token",
          CF_ACCESS_CLIENT_ID: "id",
          CF_ACCESS_CLIENT_SECRET: "secret",
          APPLICATION_SELECTOR: "environment=wallet-dev1",
          TIMEOUT_SECONDS: "300",
          POLL_INTERVAL_SECONDS: "0",
        }),
      /POLL_INTERVAL_SECONDS must be a finite number greater than 0/,
    );

    assert.throws(
      () =>
        getRequiredConfig({
          ARGOCD_ENDPOINT: "https://sandbox-argo.interledger.tech",
          ARGOCD_AUTH_TOKEN: "token",
          CF_ACCESS_CLIENT_ID: "id",
          CF_ACCESS_CLIENT_SECRET: "secret",
          APPLICATION_SELECTOR: "environment=wallet-dev1",
          TIMEOUT_SECONDS: "30",
          POLL_INTERVAL_SECONDS: "60",
        }),
      /POLL_INTERVAL_SECONDS must be less than or equal to TIMEOUT_SECONDS/,
    );
  });
});
