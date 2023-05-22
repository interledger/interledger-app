declare module "routes-gen" {
  export type RouteParams = {
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/settings/linked-identities/:identityId": { "identityId": string };
    "/settings/linked-accounts/:accountId": { "accountId": string };
    "/legal/accessibility-statement": Record<string, never>;
    "/settings/profile-public/name": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/connections/add-a-public-key": Record<string, never>;
    "/legal/us/e-sign-agreement": Record<string, never>;
    "/settings/linked-identities": Record<string, never>;
    "/connections/:connectionId": { "connectionId": string };
    "/settings/profile-personal": Record<string, never>;
    "/legal/us/privacy-policy": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/legal/terms-of-service": Record<string, never>;
    "/legal/us/terms-of-use": Record<string, never>;
    "/legal/privacy-policy": Record<string, never>;
    "/legal/us/compliance": Record<string, never>;
    "/legal/wallet-license": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/legal/us/licences": Record<string, never>;
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
    "/twitter": Record<string, never>;
    "/": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/legal": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/settings/linked-identities/:identityId", RouteParams["/settings/linked-identities/:identityId"]]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
      | ["/legal/accessibility-statement"]
      | ["/settings/profile-public/name"]
      | ["/api/maps/placesAutocomplete"]
      | ["/connections/add-a-public-key"]
      | ["/legal/us/e-sign-agreement"]
      | ["/settings/linked-identities"]
      | ["/connections/:connectionId", RouteParams["/connections/:connectionId"]]
      | ["/settings/profile-personal"]
      | ["/legal/us/privacy-policy"]
      | ["/settings/linked-accounts"]
      | ["/settings/profile-contact"]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/profile-public"]
      | ["/legal/terms-of-service"]
      | ["/legal/us/terms-of-use"]
      | ["/legal/privacy-policy"]
      | ["/legal/us/compliance"]
      | ["/legal/wallet-license"]
      | ["/temp-cloudflare-error"]
      | ["/legal/us/licences"]
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
      | ["/twitter"]
      | ["/"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/legal"]
      | ["/login"]
      | ["/blog"]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
