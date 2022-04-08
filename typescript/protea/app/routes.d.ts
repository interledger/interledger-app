declare module "routes-gen" {
  export type RouteParams = {
    "/blog/connecting-the-internet-economy": {};
    "/confirmation/:flowId": { flowId: string };
    "/confirmation/:flowId/payment-method": { flowId: string };
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
    "/flows/:flowId/payment-method/details": { flowId: string };
    "/flows/:flowId/payment-method/review": { flowId: string };
    "/flows/:flowId/payment-method/type": { flowId: string };
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
      | ["/confirmation/:flowId/payment-method", RouteParams["/confirmation/:flowId/payment-method"]]
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
      | ["/flows/:flowId/payment-method/details", RouteParams["/flows/:flowId/payment-method/details"]]
      | ["/flows/:flowId/payment-method/review", RouteParams["/flows/:flowId/payment-method/review"]]
      | ["/flows/:flowId/payment-method/type", RouteParams["/flows/:flowId/payment-method/type"]]
      | ["/blog"]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
  >(...args: T): typeof args[0];
}
