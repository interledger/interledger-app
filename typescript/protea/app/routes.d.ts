declare module "routes-gen" {
  export type RouteParams = {
    "/blog/connecting-the-internet-economy": {};
    "/": {};
    "/activity/transaction/:id": { id: string };
    "/settings/password": {};
    "/transact/preview": {};
    "/transact/receive": {};
    "/activity/filter": {};
    "/activity": {};
    "/settings": {};
    "/transact": {};
    "/withdraw": {};
    "/connect": {};
    "/deposit": {};
    "/bank": {};
    "/home": {};
    "/blog": {};
    "/recovery": {};
    "/logout": {};
    "/signup": {};
    "/verify": {};
    "/login": {};
  };

  export function route<
    T extends
      | ["/blog/connecting-the-internet-economy"]
      | ["/"]
      | ["/activity/transaction/:id", RouteParams["/activity/transaction/:id"]]
      | ["/settings/password"]
      | ["/transact/preview"]
      | ["/transact/receive"]
      | ["/activity/filter"]
      | ["/activity"]
      | ["/settings"]
      | ["/transact"]
      | ["/withdraw"]
      | ["/connect"]
      | ["/deposit"]
      | ["/bank"]
      | ["/home"]
      | ["/blog"]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
  >(...args: T): typeof args[0];
}
