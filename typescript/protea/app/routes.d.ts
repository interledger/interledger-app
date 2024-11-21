declare module "routes-gen" {
  export type RouteParams = {
    "/": Record<string, never>;
    "/:slug": { "slug": string };
    "/accounts": Record<string, never>;
    "/accounts/:accountId": { "accountId": string };
    "/accounts/:accountId/name": { "accountId": string };
    "/api/fern": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/callbacks/chimoney": Record<string, never>;
    "/connect/bank/za": Record<string, never>;
    "/connect/card": Record<string, never>;
    "/connect/interac": Record<string, never>;
    "/consent": Record<string, never>;
    "/contact": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/deposit": Record<string, never>;
    "/deposit/:paymentId": { "paymentId": string };
    "/healthz": Record<string, never>;
    "/legal/:jurisdiction?/:slug": { "jurisdiction"?: string, "slug": string };
    "/login": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/logout": Record<string, never>;
    "/otp/challenge": Record<string, never>;
    "/pay": Record<string, never>;
    "/pay/:paymentId": { "paymentId": string };
    "/payments": Record<string, never>;
    "/payments/:paymentId": { "paymentId": string };
    "/personal-details": Record<string, never>;
    "/recovery": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings": Record<string, never>;
    "/settings/grants": Record<string, never>;
    "/settings/grants/:grantId": { "grantId": string };
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
    "/support": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/transactions": Record<string, never>;
    "/transactions/:transactionId": { "transactionId": string };
    "/verify": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/wallet-address": Record<string, never>;
    "/withdraw": Record<string, never>;
    "/withdraw/:paymentId": { "paymentId": string };
  };

  export function route<
    T extends
      | ["/"]
      | ["/:slug", RouteParams["/:slug"]]
      | ["/accounts"]
      | ["/accounts/:accountId", RouteParams["/accounts/:accountId"]]
      | ["/accounts/:accountId/name", RouteParams["/accounts/:accountId/name"]]
      | ["/api/fern"]
      | ["/api/maps/geocode"]
      | ["/api/maps/placesAutocomplete"]
      | ["/api/sendOtp"]
      | ["/callbacks/chimoney"]
      | ["/connect/bank/za"]
      | ["/connect/card"]
      | ["/connect/interac"]
      | ["/consent"]
      | ["/contact"]
      | ["/contact/success"]
      | ["/deposit"]
      | ["/deposit/:paymentId", RouteParams["/deposit/:paymentId"]]
      | ["/healthz"]
      | ["/legal/:jurisdiction?/:slug", RouteParams["/legal/:jurisdiction?/:slug"]]
      | ["/login"]
      | ["/login/challenge"]
      | ["/logout"]
      | ["/otp/challenge"]
      | ["/pay"]
      | ["/pay/:paymentId", RouteParams["/pay/:paymentId"]]
      | ["/payments"]
      | ["/payments/:paymentId", RouteParams["/payments/:paymentId"]]
      | ["/personal-details"]
      | ["/recovery"]
      | ["/recovery/password"]
      | ["/settings"]
      | ["/settings/grants"]
      | ["/settings/grants/:grantId", RouteParams["/settings/grants/:grantId"]]
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
      | ["/support"]
      | ["/temp-cloudflare-error"]
      | ["/transactions"]
      | ["/transactions/:transactionId", RouteParams["/transactions/:transactionId"]]
      | ["/verify"]
      | ["/waitlist"]
      | ["/waitlist/success"]
      | ["/wallet-address"]
      | ["/withdraw"]
      | ["/withdraw/:paymentId", RouteParams["/withdraw/:paymentId"]]
  >(...args: T): typeof args[0];
}
