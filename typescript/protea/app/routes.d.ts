declare module "routes-gen" {
  export type RouteParams = {
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/activity/transaction/:id": { "id": string };
    "/settings/linked-accounts": Record<string, never>;
    "/confirmation/:flowId": { "flowId": string };
    "/confirmation/:flowId/linked-account": { "flowId": string };
    "/confirmation/:flowId/withdraw": { "flowId": string };
    "/confirmation/:flowId/deposit": { "flowId": string };
    "/confirmation/:flowId/send": { "flowId": string };
    "/legal/privacy-policy": Record<string, never>;
    "/legal/terms-of-use": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/activity/filter": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/onboarding/unit": Record<string, never>;
    "/flows/:flowId": { "flowId": string };
    "/flows/:flowId/linked-account/details": { "flowId": string };
    "/flows/:flowId/linked-account/review": { "flowId": string };
    "/flows/:flowId/linked-account/type": { "flowId": string };
    "/flows/:flowId/signup/password": { "flowId": string };
    "/flows/:flowId/signup/about": { "flowId": string };
    "/flows/:flowId/signup/phone": { "flowId": string };
    "/flows/:flowId/signup/sms": { "flowId": string };
    "/disclosures": Record<string, never>;
    "/activity": Record<string, never>;
    "/recovery": Record<string, never>;
    "/settings": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/connect": Record<string, never>;
    "/contact": Record<string, never>;
    "/receive": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/about": Record<string, never>;
    "/": Record<string, never>;
    "/legal": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/blog/connecting-the-internet-economy": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/maps/placesAutocomplete"]
      | ["/what-is-a-payment-pointer"]
      | ["/activity/transaction/:id", RouteParams["/activity/transaction/:id"]]
      | ["/settings/linked-accounts"]
      | ["/confirmation/:flowId", RouteParams["/confirmation/:flowId"]]
      | ["/confirmation/:flowId/linked-account", RouteParams["/confirmation/:flowId/linked-account"]]
      | ["/confirmation/:flowId/withdraw", RouteParams["/confirmation/:flowId/withdraw"]]
      | ["/confirmation/:flowId/deposit", RouteParams["/confirmation/:flowId/deposit"]]
      | ["/confirmation/:flowId/send", RouteParams["/confirmation/:flowId/send"]]
      | ["/legal/privacy-policy"]
      | ["/legal/terms-of-use"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/api/maps/geocode"]
      | ["/waitlist/success"]
      | ["/activity/filter"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/onboarding/unit"]
      | ["/flows/:flowId", RouteParams["/flows/:flowId"]]
      | ["/flows/:flowId/linked-account/details", RouteParams["/flows/:flowId/linked-account/details"]]
      | ["/flows/:flowId/linked-account/review", RouteParams["/flows/:flowId/linked-account/review"]]
      | ["/flows/:flowId/linked-account/type", RouteParams["/flows/:flowId/linked-account/type"]]
      | ["/flows/:flowId/signup/password", RouteParams["/flows/:flowId/signup/password"]]
      | ["/flows/:flowId/signup/about", RouteParams["/flows/:flowId/signup/about"]]
      | ["/flows/:flowId/signup/phone", RouteParams["/flows/:flowId/signup/phone"]]
      | ["/flows/:flowId/signup/sms", RouteParams["/flows/:flowId/signup/sms"]]
      | ["/disclosures"]
      | ["/activity"]
      | ["/recovery"]
      | ["/settings"]
      | ["/waitlist"]
      | ["/connect"]
      | ["/contact"]
      | ["/receive"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/about"]
      | ["/"]
      | ["/legal"]
      | ["/login"]
      | ["/blog"]
      | ["/blog/connecting-the-internet-economy"]
  >(...args: T): typeof args[0];
}
