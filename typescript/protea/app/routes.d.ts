declare module "routes-gen" {
  export type RouteParams = {
    "/blog/connecting-the-internet-economy": {};
    "/": {};
    "/activity/transaction/:id": { id: string };
    "/transact/preview": {};
    "/transact/receive": {};
    "/activity/filter": {};
    "/settings/password": {};
    "/activity": {};
    "/settings": {};
    "/transact": {};
    "/withdraw": {};
    "/connect": {};
    "/deposit": {};
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
      | ["/transact/preview"]
      | ["/transact/receive"]
      | ["/activity/filter"]
      | ["/settings/password"]
      | ["/activity"]
      | ["/settings"]
      | ["/transact"]
      | ["/withdraw"]
      | ["/connect"]
      | ["/deposit"]
      | ["/home"]
      | ["/blog"]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
  >(...args: T): typeof args[0];
}
