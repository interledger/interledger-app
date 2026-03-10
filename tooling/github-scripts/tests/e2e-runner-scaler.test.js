import { describe, it } from "node:test";
import assert from "node:assert";

import {
  DEFAULT_CONFIG,
  decideScaleUp,
  decideScaleDown,
  decideCleanup,
  decideEnsureMinimum,
  buildDecision,
  executeDecision,
  generateSpotName,
} from "../e2e-runner-scaler.js";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** @returns {import('../e2e-runner-scaler.js').ScalerConfig} */
function cfg(overrides = {}) {
  return { ...DEFAULT_CONFIG, ...overrides };
}

/** @param {Partial<import('../types').QueueState>} [overrides] */
function queue(overrides = {}) {
  return { pendingCount: 0, oldestWaitSecs: 0, ...overrides };
}

/** @param {Partial<import('../types').SpotState>} [overrides] */
function spots(overrides = {}) {
  return { running: 0, booting: 0, total: 0, ...overrides };
}

/**
 * @param {string}  name
 * @param {string}  [status]
 * @param {boolean} [busy]
 * @returns {import('../types').RunnerInfo}
 */
function runner(name, status = "online", busy = false) {
  return { id: Math.floor(Math.random() * 10000), name, status, busy };
}

// ---------------------------------------------------------------------------
// decideScaleUp
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/decideScaleUp", () => {
  it("returns 0 when no pending jobs", () => {
    assert.strictEqual(
      decideScaleUp(queue({ pendingCount: 0 }), spots(), cfg()),
      0,
    );
  });

  it("returns 0 when pending but wait time below threshold", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 3, oldestWaitSecs: 10 }),
        spots(),
        cfg({ queueWaitThresholdSecs: 30 }),
      ),
      0,
    );
  });

  it("creates instances equal to pending count when capacity allows", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 3, oldestWaitSecs: 60 }),
        spots({ total: 0 }),
        cfg({ maxSpots: 5 }),
      ),
      3,
    );
  });

  it("caps at available capacity", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 10, oldestWaitSecs: 60 }),
        spots({ total: 3 }),
        cfg({ maxSpots: 5 }),
      ),
      2,
    );
  });

  it("returns 0 when already at max capacity", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 2, oldestWaitSecs: 60 }),
        spots({ total: 5 }),
        cfg({ maxSpots: 5 }),
      ),
      0,
    );
  });

  it("scales up exactly 1 when 1 pending and threshold exceeded", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 1, oldestWaitSecs: 31 }),
        spots({ total: 0 }),
        cfg({ maxSpots: 5, queueWaitThresholdSecs: 30 }),
      ),
      1,
    );
  });

  it("does not scale when wait time equals threshold exactly", () => {
    assert.strictEqual(
      decideScaleUp(
        queue({ pendingCount: 1, oldestWaitSecs: 30 }),
        spots({ total: 0 }),
        cfg({ queueWaitThresholdSecs: 30 }),
      ),
      0,
    );
  });
});

// ---------------------------------------------------------------------------
// decideScaleDown
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/decideScaleDown", () => {
  it("returns empty when there are pending jobs", () => {
    const runners = [runner("e2e-spot-1"), runner("e2e-spot-2")];
    assert.deepStrictEqual(
      decideScaleDown(queue({ pendingCount: 1 }), runners, cfg()),
      [],
    );
  });

  it("returns empty when no idle spot runners exist", () => {
    const runners = [runner("e2e-spot-1", "online", true)];
    assert.deepStrictEqual(
      decideScaleDown(queue(), runners, cfg()),
      [],
    );
  });

  it("deletes all idle runners when minSpots is 0", () => {
    const runners = [
      runner("e2e-spot-1"),
      runner("e2e-spot-2"),
      runner("e2e-spot-3"),
    ];
    const result = decideScaleDown(queue(), runners, cfg({ minSpots: 0 }));
    assert.strictEqual(result.length, 3);
  });

  it("keeps 1 runner when minSpots is 1", () => {
    const runners = [
      runner("e2e-spot-1"),
      runner("e2e-spot-2"),
      runner("e2e-spot-3"),
    ];
    const result = decideScaleDown(queue(), runners, cfg({ minSpots: 1 }));
    assert.strictEqual(result.length, 2);
    assert.ok(!result.includes("e2e-spot-1"), "First runner should be kept");
  });

  it("ignores busy runners", () => {
    const runners = [
      runner("e2e-spot-1", "online", true), // busy
      runner("e2e-spot-2", "online", false), // idle
    ];
    const result = decideScaleDown(queue(), runners, cfg({ minSpots: 0 }));
    assert.deepStrictEqual(result, ["e2e-spot-2"]);
  });

  it("ignores runners not matching prefix", () => {
    const runners = [
      runner("other-runner-1"),
      runner("e2e-spot-1"),
    ];
    const result = decideScaleDown(queue(), runners, cfg({ minSpots: 0 }));
    assert.deepStrictEqual(result, ["e2e-spot-1"]);
  });

  it("ignores offline runners (not candidates for scale-down)", () => {
    const runners = [
      runner("e2e-spot-1", "offline", false),
      runner("e2e-spot-2", "online", false),
    ];
    const result = decideScaleDown(queue(), runners, cfg({ minSpots: 0 }));
    assert.deepStrictEqual(result, ["e2e-spot-2"]);
  });

  it("returns empty when idle count equals minSpots", () => {
    const runners = [runner("e2e-spot-1")];
    assert.deepStrictEqual(
      decideScaleDown(queue(), runners, cfg({ minSpots: 1 })),
      [],
    );
  });
});

// ---------------------------------------------------------------------------
// decideCleanup
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/decideCleanup", () => {
  it("returns only offline runners matching prefix", () => {
    const runners = [
      runner("e2e-spot-1", "offline"),
      runner("e2e-spot-2", "online"),
      runner("other-runner", "offline"),
    ];
    const result = decideCleanup(runners, cfg());
    assert.strictEqual(result.length, 1);
    assert.strictEqual(result[0].name, "e2e-spot-1");
  });

  it("returns empty when no offline runners", () => {
    const runners = [
      runner("e2e-spot-1", "online"),
      runner("e2e-spot-2", "online"),
    ];
    assert.deepStrictEqual(decideCleanup(runners, cfg()), []);
  });

  it("returns empty when all offline runners are non-matching prefix", () => {
    const runners = [runner("my-runner", "offline")];
    assert.deepStrictEqual(decideCleanup(runners, cfg()), []);
  });
});

// ---------------------------------------------------------------------------
// decideEnsureMinimum
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/decideEnsureMinimum", () => {
  it("returns 0 when minSpots is 0", () => {
    assert.strictEqual(decideEnsureMinimum(spots(), 0, cfg({ minSpots: 0 })), 0);
  });

  it("returns 0 when total already meets minimum", () => {
    assert.strictEqual(
      decideEnsureMinimum(spots({ total: 2 }), 0, cfg({ minSpots: 1 })),
      0,
    );
  });

  it("accounts for pending deletions", () => {
    assert.strictEqual(
      decideEnsureMinimum(spots({ total: 2 }), 2, cfg({ minSpots: 1 })),
      1,
    );
  });

  it("creates enough to reach minimum from zero", () => {
    assert.strictEqual(
      decideEnsureMinimum(spots({ total: 0 }), 0, cfg({ minSpots: 3 })),
      3,
    );
  });
});

// ---------------------------------------------------------------------------
// buildDecision (integration of all decisions)
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/buildDecision", () => {
  it("scales up when queue is backed up", () => {
    const decision = buildDecision(
      queue({ pendingCount: 3, oldestWaitSecs: 60 }),
      spots({ total: 1 }),
      [],
      cfg({ maxSpots: 5, minSpots: 0 }),
    );
    assert.strictEqual(decision.instancesToCreate, 3);
    assert.strictEqual(decision.instancesToDelete.length, 0);
    assert.strictEqual(decision.runnersToCleanup.length, 0);
  });

  it("scales down and cleans up when queue is empty", () => {
    const runners = [
      runner("e2e-spot-1", "online", false),
      runner("e2e-spot-2", "online", false),
      runner("e2e-spot-3", "offline", false),
    ];
    const decision = buildDecision(
      queue(),
      spots({ running: 2, total: 2 }),
      runners,
      cfg({ minSpots: 0 }),
    );
    assert.strictEqual(decision.instancesToCreate, 0);
    assert.strictEqual(decision.instancesToDelete.length, 2);
    assert.strictEqual(decision.runnersToCleanup.length, 1);
    assert.strictEqual(decision.runnersToCleanup[0].name, "e2e-spot-3");
  });

  it("ensures minimum when scaling down would go below", () => {
    const runners = [
      runner("e2e-spot-1", "online", false),
      runner("e2e-spot-2", "online", false),
    ];
    const decision = buildDecision(
      queue(),
      spots({ running: 2, total: 2 }),
      runners,
      cfg({ minSpots: 1 }),
    );
    // Deletes 1 (keeps 1), ensure-minimum contributes 0 since 2-1 = 1 >= minSpots
    assert.strictEqual(decision.instancesToDelete.length, 1);
    assert.strictEqual(decision.instancesToCreate, 0);
  });

  it("does nothing when queue is empty and no runners exist", () => {
    const decision = buildDecision(queue(), spots(), [], cfg({ minSpots: 0 }));
    assert.strictEqual(decision.instancesToCreate, 0);
    assert.strictEqual(decision.instancesToDelete.length, 0);
    assert.strictEqual(decision.runnersToCleanup.length, 0);
  });

  it("combines scale-up and ensure-minimum without double-counting", () => {
    // No existing spots, 1 pending job, minSpots=2
    // scaleUp = 1 (1 pending), ensureMin = max(0, 2 - (0 + 1)) = 1
    // Total = 1 + 1 = 2 (NOT 3: the scale-up instance counts toward minimum)
    const decision = buildDecision(
      queue({ pendingCount: 1, oldestWaitSecs: 60 }),
      spots({ total: 0 }),
      [],
      cfg({ maxSpots: 5, minSpots: 2 }),
    );
    assert.strictEqual(decision.instancesToCreate, 2);
  });
});

// ---------------------------------------------------------------------------
// generateSpotName
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/generateSpotName", () => {
  it("starts with the given prefix", () => {
    const name = generateSpotName("e2e-spot");
    assert.ok(name.startsWith("e2e-spot-"), `Expected '${name}' to start with 'e2e-spot-'`);
  });

  it("generates unique names on successive calls", () => {
    const names = new Set(Array.from({ length: 20 }, () => generateSpotName("e2e-spot")));
    assert.ok(names.size > 1, "Expected multiple unique names");
  });
});

// ---------------------------------------------------------------------------
// executeDecision
// ---------------------------------------------------------------------------

describe("e2e-runner-scaler/executeDecision", () => {
  it("calls gcp.createSpotInstance for each instance to create", async () => {
    const created = /** @type {string[]} */ ([]);
    const tokenCalls = /** @type {number[]} */ ([]);

    /** @type {import('../types').GcpAdapter} */
    const gcp = {
      createSpotInstance: async (name) => { created.push(name); },
      deleteInstance: async () => {},
    };

    /** @type {import('../types').GitHubAdapter} */
    const gh = {
      getRegistrationToken: async () => { tokenCalls.push(1); return "fake-token"; },
      removeRunner: async () => {},
    };

    await executeDecision(
      { instancesToCreate: 3, instancesToDelete: [], runnersToCleanup: [] },
      gcp,
      gh,
      cfg(),
      () => {},
    );

    assert.strictEqual(created.length, 3);
    assert.strictEqual(tokenCalls.length, 1, "Should request token exactly once");
  });

  it("calls gcp.deleteInstance for each instance to delete", async () => {
    const deleted = /** @type {string[]} */ ([]);

    /** @type {import('../types').GcpAdapter} */
    const gcp = {
      createSpotInstance: async () => {},
      deleteInstance: async (name) => { deleted.push(name); },
    };

    /** @type {import('../types').GitHubAdapter} */
    const gh = {
      getRegistrationToken: async () => "fake",
      removeRunner: async () => {},
    };

    await executeDecision(
      { instancesToCreate: 0, instancesToDelete: ["vm-1", "vm-2"], runnersToCleanup: [] },
      gcp,
      gh,
      cfg(),
      () => {},
    );

    assert.deepStrictEqual(deleted, ["vm-1", "vm-2"]);
  });

  it("calls gh.removeRunner for each offline runner", async () => {
    const removed = /** @type {number[]} */ ([]);

    /** @type {import('../types').GcpAdapter} */
    const gcp = {
      createSpotInstance: async () => {},
      deleteInstance: async () => {},
    };

    /** @type {import('../types').GitHubAdapter} */
    const gh = {
      getRegistrationToken: async () => "fake",
      removeRunner: async (id) => { removed.push(id); },
    };

    await executeDecision(
      {
        instancesToCreate: 0,
        instancesToDelete: [],
        runnersToCleanup: [
          { id: 10, name: "e2e-spot-old", status: "offline", busy: false },
          { id: 20, name: "e2e-spot-older", status: "offline", busy: false },
        ],
      },
      gcp,
      gh,
      cfg(),
      () => {},
    );

    assert.deepStrictEqual(removed, [10, 20]);
  });

  it("does nothing when decision is empty", async () => {
    let anyCalled = false;

    /** @type {import('../types').GcpAdapter} */
    const gcp = {
      createSpotInstance: async () => { anyCalled = true; },
      deleteInstance: async () => { anyCalled = true; },
    };

    /** @type {import('../types').GitHubAdapter} */
    const gh = {
      getRegistrationToken: async () => { anyCalled = true; return "x"; },
      removeRunner: async () => { anyCalled = true; },
    };

    await executeDecision(
      { instancesToCreate: 0, instancesToDelete: [], runnersToCleanup: [] },
      gcp,
      gh,
      cfg(),
      () => {},
    );

    assert.strictEqual(anyCalled, false);
  });
});
