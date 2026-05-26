/** @type {Array<Platform>} */
export const PLATFORMS = [{ os: "linux", arch: "amd64" }];

/** @type {Array<GoPackage>} */
export const GO_PACKAGES = [
  {
    name: "backend",
    type: "go",
    context: "./go",
    file: "./go/backend/Dockerfile",
  },
  {
    name: "mockgatehub",
    type: "go",
    context: "./go",
    file: "./go/mock/mockgatehub/Dockerfile",
  },
  {
    name: "mockpti",
    type: "go",
    context: "./go",
    file: "./go/mock/mockpti/Dockerfile",
  },
  {
    name: "mockxago",
    type: "go",
    context: "./go",
    file: "./go/mock/mockxago/Dockerfile",
  },
];

/** @type {Array<TsPackage>} */
export const TS_PACKAGES = [
  {
    name: "protea",
    type: "ts",
    context: "./typescript/protea",
    file: "./typescript/protea/Dockerfile",
    sentry_project: "interledger-wallet",
    cdn_bucket: "common-interledger-cdn-assets",
    cdn_public_path: "https://cdn.interledger.app/protea",
  },
  {
    name: "botanist",
    type: "ts",
    context: "./typescript/botanist",
    file: "./typescript/botanist/Dockerfile",
    sentry_project: "",
    cdn_bucket: "common-interledger-cdn-assets",
    cdn_public_path: "https://cdn.interledger.app/botanist",    
  },
];
