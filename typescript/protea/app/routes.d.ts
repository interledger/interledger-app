declare module "routes-gen" {
  export type RouteParams = {
    "/settings/profile-public/name": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/legal/:jurisdiction?/:slug": { "jurisdiction"?: string, "slug": string };
    "/me/identities/:identityId": { "identityId": string };
    "/settings/keys/add-public": Record<string, never>;
    "/accounts/:accountId/name": { "accountId": string };
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/keys/:keyId": { "keyId": string };
    "/temp-cloudflare-error": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/connect/twitter": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/wallet-address": Record<string, never>;
    "/connect/bank": Record<string, never>;
    "/connect/card": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/transactions": Record<string, never>;
    "/transactions/:transactionId": { "transactionId": string };
    "/blog/:slug": { "slug": string };
    "/identities": Record<string, never>;
    "/identities/:identityId": { "identityId": string };
    "/api/fern": Record<string, never>;
    "/accounts": Record<string, never>;
    "/accounts/:accountId": { "accountId": string };
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
    "/discord": Record<string, never>;
    "/support": Record<string, never>;
    "/": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/wallet": Record<string, never>;
    "/about": Record<string, never>;
    "/legal": Record<string, never>;
    "/login": Record<string, never>;
    "/slack": Record<string, never>;
    "/blog": Record<string, never>;
    "/docs": Record<string, never>;
    "/docs/:slug": { "slug": string };
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/settings/profile-public/name"]
      | ["/api/maps/placesAutocomplete"]
      | ["/legal/:jurisdiction?/:slug", RouteParams["/legal/:jurisdiction?/:slug"]]
      | ["/me/identities/:identityId", RouteParams["/me/identities/:identityId"]]
      | ["/settings/keys/add-public"]
      | ["/accounts/:accountId/name", RouteParams["/accounts/:accountId/name"]]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/keys/:keyId", RouteParams["/settings/keys/:keyId"]]
      | ["/temp-cloudflare-error"]
      | ["/api/maps/geocode"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/waitlist/success"]
      | ["/connect/twitter"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/personal-details"]
      | ["/wallet-address"]
      | ["/connect/bank"]
      | ["/connect/card"]
      | ["/api/sendOtp"]
      | ["/transactions"]
      | ["/transactions/:transactionId", RouteParams["/transactions/:transactionId"]]
      | ["/blog/:slug", RouteParams["/blog/:slug"]]
      | ["/identities"]
      | ["/identities/:identityId", RouteParams["/identities/:identityId"]]
      | ["/api/fern"]
      | ["/accounts"]
      | ["/accounts/:accountId", RouteParams["/accounts/:accountId"]]
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
      | ["/discord"]
      | ["/support"]
      | ["/"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/wallet"]
      | ["/about"]
      | ["/legal"]
      | ["/login"]
      | ["/slack"]
      | ["/blog"]
      | ["/docs"]
      | ["/docs/:slug", RouteParams["/docs/:slug"]]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
