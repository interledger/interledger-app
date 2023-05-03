declare module "routes-gen" {
  export type RouteParams = {
    "/api/allowSignup": Record<string, never>;
    "/wallet/:id": { "id": string };
    "/waitlist": Record<string, never>;
    "/wallets": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/allowSignup"]
      | ["/wallet/:id", RouteParams["/wallet/:id"]]
      | ["/waitlist"]
      | ["/wallets"]
      | ["/"]
  >(...args: T): typeof args[0];
}
