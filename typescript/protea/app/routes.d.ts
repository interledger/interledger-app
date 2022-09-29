declare module "routes-gen" {
  export type RouteParams = {
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/activity/transaction/:id": { "id": string };
    "/settings/linked-accounts": Record<string, never>;
    "/signup/:flowId/password": { "flowId": string };
    "/legal/privacy-policy": Record<string, never>;
    "/signup/:flowId/about": { "flowId": string };
    "/signup/:flowId/phone": { "flowId": string };
    "/legal/terms-of-use": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
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
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/maps/placesAutocomplete"]
      | ["/what-is-a-payment-pointer"]
      | ["/activity/transaction/:id", RouteParams["/activity/transaction/:id"]]
      | ["/settings/linked-accounts"]
      | ["/signup/:flowId/password", RouteParams["/signup/:flowId/password"]]
      | ["/legal/privacy-policy"]
      | ["/signup/:flowId/about", RouteParams["/signup/:flowId/about"]]
      | ["/signup/:flowId/phone", RouteParams["/signup/:flowId/phone"]]
      | ["/legal/terms-of-use"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/api/maps/geocode"]
      | ["/waitlist/success"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/payment-pointer"]
      | ["/api/sendOtp"]
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
      | ["/pay"]
  >(...args: T): typeof args[0];
}
