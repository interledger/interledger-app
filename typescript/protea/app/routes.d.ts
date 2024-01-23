declare module "routes-gen" {
  export type RouteParams = {
    "/": Record<string, never>;
    "/:slug": { "slug": string };
    "/about": Record<string, never>;
    "/accounts": Record<string, never>;
    "/accounts/:accountId": { "accountId": string };
    "/accounts/:accountId/name": { "accountId": string };
    "/api/fern": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/blog": Record<string, never>;
    "/blog/:slug": { "slug": string };
    "/collectables": Record<string, never>;
    "/connect/bank": Record<string, never>;
    "/connect/bank/za": Record<string, never>;
    "/connect/card": Record<string, never>;
    "/connect/discord": Record<string, never>;
    "/connect/domain": Record<string, never>;
    "/connect/slack": Record<string, never>;
    "/connect/twitter": Record<string, never>;
    "/consent": Record<string, never>;
    "/contact": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/deposit": Record<string, never>;
    "/deposit/:paymentId": { "paymentId": string };
    "/discord": Record<string, never>;
    "/docs": Record<string, never>;
    "/docs/:slug": { "slug": string };
    "/form/:slug": { "slug": string };
    "/identities": Record<string, never>;
    "/identities/:identityId": { "identityId": string };
    "/legal": Record<string, never>;
    "/legal/:jurisdiction?/:slug": { "jurisdiction"?: string, "slug": string };
    "/login": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/logout": Record<string, never>;
    "/me/identities/:identityId": { "identityId": string };
    "/otp/challenge": Record<string, never>;
    "/pay": Record<string, never>;
    "/pay/:paymentId": { "paymentId": string };
    "/pay/3ds": Record<string, never>;
    "/payments": Record<string, never>;
    "/payments/:paymentId": { "paymentId": string };
    "/personal-details": Record<string, never>;
    "/recovery": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/referral": Record<string, never>;
    "/settings": Record<string, never>;
    "/settings/keys": Record<string, never>;
    "/settings/keys/:keyId": { "keyId": string };
    "/settings/keys/add-public": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/settings/phone": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/settings/profile-personal": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/settings/profile-public/name": Record<string, never>;
    "/signup": Record<string, never>;
    "/slack": Record<string, never>;
    "/support": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/thank-you/:slug": { "slug": string };
    "/transactions": Record<string, never>;
    "/transactions/:transactionId": { "transactionId": string };
    "/verify": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/wallet": Record<string, never>;
    "/wallet-address": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/withdraw": Record<string, never>;
    "/withdraw/:paymentId": { "paymentId": string };
  };

  export function route<
    T extends
      | ["/"]
      | ["/:slug", RouteParams["/:slug"]]
      | ["/about"]
      | ["/accounts"]
      | ["/accounts/:accountId", RouteParams["/accounts/:accountId"]]
      | ["/accounts/:accountId/name", RouteParams["/accounts/:accountId/name"]]
      | ["/api/fern"]
      | ["/api/maps/geocode"]
      | ["/api/maps/placesAutocomplete"]
      | ["/api/sendOtp"]
      | ["/blog"]
      | ["/blog/:slug", RouteParams["/blog/:slug"]]
      | ["/collectables"]
      | ["/connect/bank"]
      | ["/connect/bank/za"]
      | ["/connect/card"]
      | ["/connect/discord"]
      | ["/connect/domain"]
      | ["/connect/slack"]
      | ["/connect/twitter"]
      | ["/consent"]
      | ["/contact"]
      | ["/contact/success"]
      | ["/deposit"]
      | ["/deposit/:paymentId", RouteParams["/deposit/:paymentId"]]
      | ["/discord"]
      | ["/docs"]
      | ["/docs/:slug", RouteParams["/docs/:slug"]]
      | ["/form/:slug", RouteParams["/form/:slug"]]
      | ["/identities"]
      | ["/identities/:identityId", RouteParams["/identities/:identityId"]]
      | ["/legal"]
      | ["/legal/:jurisdiction?/:slug", RouteParams["/legal/:jurisdiction?/:slug"]]
      | ["/login"]
      | ["/login/challenge"]
      | ["/logout"]
      | ["/me/identities/:identityId", RouteParams["/me/identities/:identityId"]]
      | ["/otp/challenge"]
      | ["/pay"]
      | ["/pay/:paymentId", RouteParams["/pay/:paymentId"]]
      | ["/pay/3ds"]
      | ["/payments"]
      | ["/payments/:paymentId", RouteParams["/payments/:paymentId"]]
      | ["/personal-details"]
      | ["/recovery"]
      | ["/recovery/password"]
      | ["/referral"]
      | ["/settings"]
      | ["/settings/keys"]
      | ["/settings/keys/:keyId", RouteParams["/settings/keys/:keyId"]]
      | ["/settings/keys/add-public"]
      | ["/settings/password"]
      | ["/settings/phone"]
      | ["/settings/profile-contact"]
      | ["/settings/profile-personal"]
      | ["/settings/profile-public"]
      | ["/settings/profile-public/name"]
      | ["/signup"]
      | ["/slack"]
      | ["/support"]
      | ["/temp-cloudflare-error"]
      | ["/thank-you/:slug", RouteParams["/thank-you/:slug"]]
      | ["/transactions"]
      | ["/transactions/:transactionId", RouteParams["/transactions/:transactionId"]]
      | ["/verify"]
      | ["/waitlist"]
      | ["/waitlist/success"]
      | ["/wallet"]
      | ["/wallet-address"]
      | ["/what-is-a-payment-pointer"]
      | ["/withdraw"]
      | ["/withdraw/:paymentId", RouteParams["/withdraw/:paymentId"]]
  >(...args: T): typeof args[0];
}
