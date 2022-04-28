declare module "routes-gen" {
  export type RouteParams = {
    "/confirmation/:flowId": { flowId: string };
    "/confirmation/:flowId/payment-method": { flowId: string };
    "/confirmation/:flowId/withdraw": { flowId: string };
    "/confirmation/:flowId/deposit": { flowId: string };
    "/confirmation/:flowId/send": { flowId: string };
    "/recovery/password": {};
    "/login/challenge": {};
    "/": {};
    "/activity/transaction/:id": { id: string };
    "/settings/payment-methods": {};
    "/settings/password": {};
    "/activity/filter": {};
    "/activity": {};
    "/settings": {};
    "/connect": {};
    "/receive": {};
    "/home": {};
    "/flows/:flowId": { flowId: string };
    "/flows/:flowId/withdraw/payment-method": { flowId: string };
    "/flows/:flowId/deposit/payment-method": { flowId: string };
    "/flows/:flowId/payment-method/details": { flowId: string };
    "/flows/:flowId/payment-method/review": { flowId: string };
    "/flows/:flowId/payment-method/type": { flowId: string };
    "/flows/:flowId/withdraw/amount": { flowId: string };
    "/flows/:flowId/withdraw/review": { flowId: string };
    "/flows/:flowId/deposit/amount": { flowId: string };
    "/flows/:flowId/deposit/review": { flowId: string };
    "/flows/:flowId/send/amount": { flowId: string };
    "/flows/:flowId/send/review": { flowId: string };
    "/flows/:flowId/send/to": { flowId: string };
    "/recovery": {};
    "/logout": {};
    "/signup": {};
    "/verify": {};
    "/login": {};
    "/blog": {};
    "/blog/connecting-the-internet-economy": {};
  };

  export function route<
    T extends
      | ["/confirmation/:flowId", RouteParams["/confirmation/:flowId"]]
      | ["/confirmation/:flowId/payment-method", RouteParams["/confirmation/:flowId/payment-method"]]
      | ["/confirmation/:flowId/withdraw", RouteParams["/confirmation/:flowId/withdraw"]]
      | ["/confirmation/:flowId/deposit", RouteParams["/confirmation/:flowId/deposit"]]
      | ["/confirmation/:flowId/send", RouteParams["/confirmation/:flowId/send"]]
      | ["/recovery/password"]
      | ["/login/challenge"]
      | ["/"]
      | ["/activity/transaction/:id", RouteParams["/activity/transaction/:id"]]
      | ["/settings/payment-methods"]
      | ["/settings/password"]
      | ["/activity/filter"]
      | ["/activity"]
      | ["/settings"]
      | ["/connect"]
      | ["/receive"]
      | ["/home"]
      | ["/flows/:flowId", RouteParams["/flows/:flowId"]]
      | ["/flows/:flowId/withdraw/payment-method", RouteParams["/flows/:flowId/withdraw/payment-method"]]
      | ["/flows/:flowId/deposit/payment-method", RouteParams["/flows/:flowId/deposit/payment-method"]]
      | ["/flows/:flowId/payment-method/details", RouteParams["/flows/:flowId/payment-method/details"]]
      | ["/flows/:flowId/payment-method/review", RouteParams["/flows/:flowId/payment-method/review"]]
      | ["/flows/:flowId/payment-method/type", RouteParams["/flows/:flowId/payment-method/type"]]
      | ["/flows/:flowId/withdraw/amount", RouteParams["/flows/:flowId/withdraw/amount"]]
      | ["/flows/:flowId/withdraw/review", RouteParams["/flows/:flowId/withdraw/review"]]
      | ["/flows/:flowId/deposit/amount", RouteParams["/flows/:flowId/deposit/amount"]]
      | ["/flows/:flowId/deposit/review", RouteParams["/flows/:flowId/deposit/review"]]
      | ["/flows/:flowId/send/amount", RouteParams["/flows/:flowId/send/amount"]]
      | ["/flows/:flowId/send/review", RouteParams["/flows/:flowId/send/review"]]
      | ["/flows/:flowId/send/to", RouteParams["/flows/:flowId/send/to"]]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
      | ["/blog"]
      | ["/blog/connecting-the-internet-economy"]
  >(...args: T): typeof args[0];
}
