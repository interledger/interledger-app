/** @type {Array<Platform>} */
export const PLATFORMS = [{ os: "linux", arch: "amd64" }];

/** @type Array<GoPackage> */
export const GO_PACKAGES = [
  {
    name: "backend",
    type: "go",
    context: "./go",
    file: "./go/backend/Dockerfile",
  },
  {
    name: "mockbos",
    type: "go",
    context: "./go",
    file: "./go/mockbos/Dockerfile",
  },
];

/** @type Array<TsPackage> */
export const TS_PACKAGES = [
  {
    name: "protea",
    type: "ts",
    context: "./typescript/protea",
    file: "./typescript/protea/Dockerfile",
    sentry_project: "interledger-wallet",
  },
  {
    name: "botanist",
    type: "ts",
    context: "./typescript/botanist",
    file: "./typescript/botanist/Dockerfile",
    sentry_project: "",
  },
];
