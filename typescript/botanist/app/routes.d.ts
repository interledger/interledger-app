declare module "routes-gen" {
  export type RouteParams = {
    "/api/allowSignup": Record<string, never>;
    "/statements": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/allowSignup"]
      | ["/statements"]
      | ["/"]
  >(...args: T): typeof args[0];
}
