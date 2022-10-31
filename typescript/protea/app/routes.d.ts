declare module "routes-gen" {
  export type RouteParams = {
    "/blog/connecting-the-internet-economy": Record<string, never>;
    "/linked-account/:type/:flowId/success": { "type": string, "flowId": string };
    "/personal-details/:flowId/address": { "flowId": string };
    "/blog/our-fynbos-family-meet-don": Record<string, never>;
    "/personal-details/:flowId/about": { "flowId": string };
    "/blog/card-payments-still-suck": Record<string, never>;
    "/linked-account/:type/:flowId": { "type": string, "flowId": string };
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/settings/personal-details": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/signup/:flowId/password": { "flowId": string };
    "/legal/privacy-policy": Record<string, never>;
    "/linked-account/:type": { "type": string };
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
    "/pay/amount": Record<string, never>;
    "/recovery": Record<string, never>;
    "/settings": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/contact": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/blog/connecting-the-internet-economy"]
      | ["/linked-account/:type/:flowId/success", RouteParams["/linked-account/:type/:flowId/success"]]
      | ["/personal-details/:flowId/address", RouteParams["/personal-details/:flowId/address"]]
      | ["/blog/our-fynbos-family-meet-don"]
      | ["/personal-details/:flowId/about", RouteParams["/personal-details/:flowId/about"]]
      | ["/blog/card-payments-still-suck"]
      | ["/linked-account/:type/:flowId", RouteParams["/linked-account/:type/:flowId"]]
      | ["/api/maps/placesAutocomplete"]
      | ["/settings/personal-details"]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/linked-accounts"]
      | ["/signup/:flowId/password", RouteParams["/signup/:flowId/password"]]
      | ["/legal/privacy-policy"]
      | ["/linked-account/:type", RouteParams["/linked-account/:type"]]
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
      | ["/pay/amount"]
      | ["/recovery"]
      | ["/settings"]
      | ["/waitlist"]
      | ["/contact"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/"]
      | ["/login"]
      | ["/blog"]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
