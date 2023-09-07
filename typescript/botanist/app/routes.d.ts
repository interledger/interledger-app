declare module "routes-gen" {
  export type RouteParams = {
    "/dynamic-forms": Record<string, never>;
    "/dynamic-forms/:id.csv": { "id": string };
    "/review/:id": { "id": string };
    "/review/:id/details": { "id": string };
    "/wallet/:id": { "id": string };
    "/wallet/:id/linked-accounts": { "id": string };
    "/wallet/:id/transactions": { "id": string };
    "/wallet/:id/transactions/:transactionId": { "id": string, "transactionId": string };
    "/wallet/:id/profile": { "id": string };
    "/wallet/:id/audit": { "id": string };
    "/waitlist": Record<string, never>;
    "/reviews": Record<string, never>;
    "/wallets": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/dynamic-forms"]
      | ["/dynamic-forms/:id.csv", RouteParams["/dynamic-forms/:id.csv"]]
      | ["/review/:id", RouteParams["/review/:id"]]
      | ["/review/:id/details", RouteParams["/review/:id/details"]]
      | ["/wallet/:id", RouteParams["/wallet/:id"]]
      | ["/wallet/:id/linked-accounts", RouteParams["/wallet/:id/linked-accounts"]]
      | ["/wallet/:id/transactions", RouteParams["/wallet/:id/transactions"]]
      | ["/wallet/:id/transactions/:transactionId", RouteParams["/wallet/:id/transactions/:transactionId"]]
      | ["/wallet/:id/profile", RouteParams["/wallet/:id/profile"]]
      | ["/wallet/:id/audit", RouteParams["/wallet/:id/audit"]]
      | ["/waitlist"]
      | ["/reviews"]
      | ["/wallets"]
      | ["/"]
  >(...args: T): typeof args[0];
}
