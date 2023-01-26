declare module "routes-gen" {
  export type RouteParams = {
    "/api/allowSignup": Record<string, never>;
    "/wallets/:id": { "id": string };
    "/waitlist": Record<string, never>;
    "/wallets": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/allowSignup"]
      | ["/wallets/:id", RouteParams["/wallets/:id"]]
      | ["/waitlist"]
      | ["/wallets"]
      | ["/"]
  >(...args: T): typeof args[0];
}
