declare module "routes-gen" {
  export type RouteParams = {
    "/api/maps/geocode": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/blog": Record<string, never>;
    "/blog/:slug": { "slug": string };
    "/connections": Record<string, never>;
    "/connections/:connectionId": { "connectionId": string };
    "/connections/add-a-public-key": Record<string, never>;
    "/contact": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/contacts": Record<string, never>;
    "/legal": Record<string, never>;
    "/legal/accessibility-statement": Record<string, never>;
    "/legal/privacy-policy": Record<string, never>;
    "/legal/terms-of-service": Record<string, never>;
    "/legal/us/compliance": Record<string, never>;
    "/legal/us/e-sign-agreement": Record<string, never>;
    "/legal/us/licences": Record<string, never>;
    "/legal/us/privacy-policy": Record<string, never>;
    "/legal/us/terms-of-use": Record<string, never>;
    "/legal/wallet-license": Record<string, never>;
    "/link-account": Record<string, never>;
    "/link-account/bank": Record<string, never>;
    "/link-account/card": Record<string, never>;
    "/login": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/logout": Record<string, never>;
    "/pay": Record<string, never>;
    "/pay/3ds": Record<string, never>;
    "/pay/amount": Record<string, never>;
    "/pay/confirm": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/recovery": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/linked-accounts/:accountId": { "accountId": string };
    "/settings/linked-identities": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/settings/profile-personal": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/settings/profile-public/name": Record<string, never>;
    "/signup": Record<string, never>;
    "/signup/about": Record<string, never>;
    "/signup/password": Record<string, never>;
    "/signup/phone": Record<string, never>;
    "/support": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/transactions": Record<string, never>;
    "/twitter": Record<string, never>;
    "/verify": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/maps/geocode"]
      | ["/api/maps/placesAutocomplete"]
      | ["/api/sendOtp"]
      | ["/blog"]
      | ["/blog/:slug", RouteParams["/blog/:slug"]]
      | ["/connections"]
      | ["/connections/:connectionId", RouteParams["/connections/:connectionId"]]
      | ["/connections/add-a-public-key"]
      | ["/contact"]
      | ["/contact/success"]
      | ["/contacts"]
      | ["/legal"]
      | ["/legal/accessibility-statement"]
      | ["/legal/privacy-policy"]
      | ["/legal/terms-of-service"]
      | ["/legal/us/compliance"]
      | ["/legal/us/e-sign-agreement"]
      | ["/legal/us/licences"]
      | ["/legal/us/privacy-policy"]
      | ["/legal/us/terms-of-use"]
      | ["/legal/wallet-license"]
      | ["/link-account"]
      | ["/link-account/bank"]
      | ["/link-account/card"]
      | ["/login"]
      | ["/login/challenge"]
      | ["/logout"]
      | ["/pay"]
      | ["/pay/3ds"]
      | ["/pay/amount"]
      | ["/pay/confirm"]
      | ["/payment-pointer"]
      | ["/personal-details"]
      | ["/recovery"]
      | ["/recovery/password"]
      | ["/settings"]
      | ["/settings/linked-accounts"]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
      | ["/settings/linked-identities"]
      | ["/settings/password"]
      | ["/settings/profile-contact"]
      | ["/settings/profile-personal"]
      | ["/settings/profile-public"]
      | ["/settings/profile-public/name"]
      | ["/signup"]
      | ["/signup/about"]
      | ["/signup/password"]
      | ["/signup/phone"]
      | ["/support"]
      | ["/temp-cloudflare-error"]
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/transactions"]
      | ["/twitter"]
      | ["/verify"]
      | ["/waitlist"]
      | ["/waitlist/success"]
      | ["/what-is-a-payment-pointer"]
      | ["/"]
  >(...args: T): typeof args[0];
}
