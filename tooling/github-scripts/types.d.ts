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
}
