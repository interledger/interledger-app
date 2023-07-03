declare module "routes-gen" {
  export type RouteParams = {
    "/settings/profile-public/name": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/legal/:jurisdiction?/:slug": { "jurisdiction"?: string, "slug": string };
    "/me/identities/:identityId": { "identityId": string };
    "/settings/keys/add-public": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/keys/:keyId": { "keyId": string };
    "/temp-cloudflare-error": Record<string, never>;
    "/accounts/:accountId": { "accountId": string };
    "/api/maps/geocode": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/connect/twitter": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/signup/password": Record<string, never>;
    "/wallet-address": Record<string, never>;
    "/connect/bank": Record<string, never>;
    "/connect/card": Record<string, never>;
    "/signup/about": Record<string, never>;
    "/signup/phone": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/pay/confirm": Record<string, never>;
    "/transactions": Record<string, never>;
    "/transactions/:transactionId": { "transactionId": string };
    "/blog/:slug": { "slug": string };
    "/pay/amount": Record<string, never>;
    "/identities": Record<string, never>;
    "/identities/:identityId": { "identityId": string };
    "/accounts": Record<string, never>;
    "/contacts": Record<string, never>;
    "/pay/3ds": Record<string, never>;
    "/recovery": Record<string, never>;
    "/settings": Record<string, never>;
    "/settings/profile-personal": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/settings/keys": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/contact": Record<string, never>;
    "/support": Record<string, never>;
    "/": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/wallet": Record<string, never>;
    "/about": Record<string, never>;
    "/legal": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/docs": Record<string, never>;
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/settings/profile-public/name"]
      | ["/api/maps/placesAutocomplete"]
      | ["/legal/:jurisdiction?/:slug", RouteParams["/legal/:jurisdiction?/:slug"]]
      | ["/me/identities/:identityId", RouteParams["/me/identities/:identityId"]]
      | ["/settings/keys/add-public"]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/keys/:keyId", RouteParams["/settings/keys/:keyId"]]
      | ["/temp-cloudflare-error"]
      | ["/accounts/:accountId", RouteParams["/accounts/:accountId"]]
      | ["/api/maps/geocode"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/waitlist/success"]
      | ["/connect/twitter"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/personal-details"]
      | ["/signup/password"]
      | ["/wallet-address"]
      | ["/connect/bank"]
      | ["/connect/card"]
      | ["/signup/about"]
      | ["/signup/phone"]
      | ["/api/sendOtp"]
      | ["/pay/confirm"]
      | ["/transactions"]
      | ["/transactions/:transactionId", RouteParams["/transactions/:transactionId"]]
      | ["/blog/:slug", RouteParams["/blog/:slug"]]
      | ["/pay/amount"]
      | ["/identities"]
      | ["/identities/:identityId", RouteParams["/identities/:identityId"]]
      | ["/accounts"]
      | ["/contacts"]
      | ["/pay/3ds"]
      | ["/recovery"]
      | ["/settings"]
      | ["/settings/profile-personal"]
      | ["/settings/profile-contact"]
      | ["/settings/profile-public"]
      | ["/settings/keys"]
      | ["/waitlist"]
      | ["/contact"]
      | ["/support"]
      | ["/"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/wallet"]
      | ["/about"]
      | ["/legal"]
      | ["/login"]
      | ["/blog"]
      | ["/docs"]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
