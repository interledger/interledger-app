declare module "routes-gen" {
  export type RouteParams = {
    "/blog/connecting-the-internet-economy": {};
    "/confirmation/:flowId": { flowId: string };
    "/recovery/password": {};
    "/login/challenge": {};
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
    "/flows/:flowId": { flowId: string };
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
      | ["/confirmation/:flowId", RouteParams["/confirmation/:flowId"]]
      | ["/recovery/password"]
      | ["/login/challenge"]
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
      | ["/flows/:flowId", RouteParams["/flows/:flowId"]]
      | ["/blog"]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
  >(...args: T): typeof args[0];
}
