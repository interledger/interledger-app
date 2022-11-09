declare module "routes-gen" {
  export type RouteParams = {
    "/blog/the-future-digital-wallets-and-payment-pointers": Record<string, never>;
    "/blog/connecting-the-internet-economy": Record<string, never>;
    "/blog/our-fynbos-family-meet-matt": Record<string, never>;
    "/blog/our-fynbos-family-meet-don": Record<string, never>;
    "/blog/card-payments-still-suck": Record<string, never>;
    "/linked-account/:type/success": { "type": string };
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/linked-account/:type/widget": { "type": string };
    "/settings/personal-details": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/personal-details/address": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/personal-details/about": Record<string, never>;
    "/receipt/:transactionId": { "transactionId": string };
    "/legal/privacy-policy": Record<string, never>;
    "/linked-account/:type": { "type": string };
    "/legal/terms-of-use": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/signup/password": Record<string, never>;
    "/signup/about": Record<string, never>;
    "/signup/phone": Record<string, never>;
    "/transactions": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/pay/confirm": Record<string, never>;
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
      | ["/blog/the-future-digital-wallets-and-payment-pointers"]
      | ["/blog/connecting-the-internet-economy"]
      | ["/blog/our-fynbos-family-meet-matt"]
      | ["/blog/our-fynbos-family-meet-don"]
      | ["/blog/card-payments-still-suck"]
      | ["/linked-account/:type/success", RouteParams["/linked-account/:type/success"]]
      | ["/api/maps/placesAutocomplete"]
      | ["/linked-account/:type/widget", RouteParams["/linked-account/:type/widget"]]
      | ["/settings/personal-details"]
      | ["/what-is-a-payment-pointer"]
      | ["/personal-details/address"]
      | ["/settings/linked-accounts"]
      | ["/personal-details/about"]
      | ["/receipt/:transactionId", RouteParams["/receipt/:transactionId"]]
      | ["/legal/privacy-policy"]
      | ["/linked-account/:type", RouteParams["/linked-account/:type"]]
      | ["/legal/terms-of-use"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/api/maps/geocode"]
      | ["/waitlist/success"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/payment-pointer"]
      | ["/signup/password"]
      | ["/signup/about"]
      | ["/signup/phone"]
      | ["/transactions"]
      | ["/api/sendOtp"]
      | ["/pay/confirm"]
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
