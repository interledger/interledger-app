export {};

type BasePackage = {
  name: string;
  context: string;
  file: string;
};

declare global {
  type Platform = {
    os: string;
    arch: string;
  };

  type GoPackage = BasePackage & {
    type: "go";
  };
  type TsPackage = BasePackage & {
    type: "ts";
    /**
     * If a TS package does not have a sentry project, use an empty string.
     */
    sentry_project: string;
    cdn_bucket: string;
    cdn_public_path: string;
  };

  type Builds = {
    package: Array<GoPackage | TsPackage>;
    platform: Array<Platform>;
  };

  // -- E2E Runner Scaler types ------------------------------------------------

  type ScalerConfig = {
    maxSpots: number;
    queueWaitThresholdSecs: number;
    spotPrefix: string;
    runnerLabel: string;
    machineImage: string;
    machineType: string;
    minSpots: number;
  };

  type QueueState = {
    pendingCount: number;
    oldestWaitSecs: number;
  };

  type SpotState = {
    running: number;
    booting: number;
    total: number;
  };

  type RunnerInfo = {
    id: number;
    name: string;
    status: string;
    busy: boolean;
  };

  type ScaleDecision = {
    instancesToCreate: number;
    instancesToDelete: string[];
    runnersToCleanup: RunnerInfo[];
  };

  type SpotInstanceOptions = {
    machineImage: string;
    machineType: string;
    metadata: Record<string, string>;
  };

  type GcpAdapter = {
    createSpotInstance: (name: string, opts: SpotInstanceOptions) => Promise<void>;
    deleteInstance: (name: string) => Promise<void>;
  };

  type GitHubAdapter = {
    getRegistrationToken: () => Promise<string>;
    removeRunner: (id: number) => Promise<void>;
  };
}
