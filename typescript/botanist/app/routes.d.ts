declare module "routes-gen" {
  export type RouteParams = {
    "/api/allowSignup": Record<string, never>;
    "/wallet/:id": { "id": string };
    "/wallet/:id/transactions": { "id": string };
    "/wallet/:id/transactions/:transactionId": { "id": string, "transactionId": string };
    "/wallet/:id/profile": { "id": string };
    "/waitlist": Record<string, never>;
    "/wallets": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/allowSignup"]
      | ["/wallet/:id", RouteParams["/wallet/:id"]]
      | ["/wallet/:id/transactions", RouteParams["/wallet/:id/transactions"]]
      | ["/wallet/:id/transactions/:transactionId", RouteParams["/wallet/:id/transactions/:transactionId"]]
      | ["/wallet/:id/profile", RouteParams["/wallet/:id/profile"]]
      | ["/waitlist"]
      | ["/wallets"]
      | ["/"]
  >(...args: T): typeof args[0];
}
