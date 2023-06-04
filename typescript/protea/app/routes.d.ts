declare module "routes-gen" {
  export type RouteParams = {
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/settings/linked-identities/:identityId": { "identityId": string };
    "/settings/linked-accounts/:accountId": { "accountId": string };
    "/settings/profile-public/name": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/connections/add-a-public-key": Record<string, never>;
    "/legal/:jurisdiction?/:slug": { "jurisdiction"?: string, "slug": string };
    "/me/identities/:identityId": { "identityId": string };
    "/settings/linked-identities": Record<string, never>;
    "/connections/:connectionId": { "connectionId": string };
    "/settings/profile-personal": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/link-account/bank": Record<string, never>;
    "/link-account/card": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/connect/twitter": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/signup/password": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/signup/about": Record<string, never>;
    "/signup/phone": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/link-account": Record<string, never>;
    "/pay/confirm": Record<string, never>;
    "/transactions": Record<string, never>;
    "/blog/:slug": { "slug": string };
    "/connections": Record<string, never>;
    "/pay/amount": Record<string, never>;
    "/contacts": Record<string, never>;
    "/pay/3ds": Record<string, never>;
    "/recovery": Record<string, never>;
    "/settings": Record<string, never>;
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
    "/test": Record<string, never>;
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/settings/linked-identities/:identityId", RouteParams["/settings/linked-identities/:identityId"]]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
      | ["/settings/profile-public/name"]
      | ["/api/maps/placesAutocomplete"]
      | ["/connections/add-a-public-key"]
      | ["/legal/:jurisdiction?/:slug", RouteParams["/legal/:jurisdiction?/:slug"]]
      | ["/me/identities/:identityId", RouteParams["/me/identities/:identityId"]]
      | ["/settings/linked-identities"]
      | ["/connections/:connectionId", RouteParams["/connections/:connectionId"]]
      | ["/settings/profile-personal"]
      | ["/settings/linked-accounts"]
      | ["/settings/profile-contact"]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/profile-public"]
      | ["/temp-cloudflare-error"]
      | ["/api/maps/geocode"]
      | ["/link-account/bank"]
      | ["/link-account/card"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/waitlist/success"]
      | ["/connect/twitter"]
      | ["/contact/success"]
      | ["/login/challenge"]
      | ["/personal-details"]
      | ["/signup/password"]
      | ["/payment-pointer"]
      | ["/signup/about"]
      | ["/signup/phone"]
      | ["/api/sendOtp"]
      | ["/link-account"]
      | ["/pay/confirm"]
      | ["/transactions"]
      | ["/blog/:slug", RouteParams["/blog/:slug"]]
      | ["/connections"]
      | ["/pay/amount"]
      | ["/contacts"]
      | ["/pay/3ds"]
      | ["/recovery"]
      | ["/settings"]
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
      | ["/test"]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
