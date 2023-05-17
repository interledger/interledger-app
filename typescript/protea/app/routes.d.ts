declare module "routes-gen" {
  export type RouteParams = {
    "/blog/the-future-digital-wallets-and-payment-pointers": Record<string, never>;
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/blog/how-technical-standards-promote-innovation": Record<string, never>;
    "/blog/connecting-the-internet-economy": Record<string, never>;
    "/settings/linked-accounts/:accountId": { "accountId": string };
    "/blog/our-fynbos-family-meet-barnard": Record<string, never>;
    "/blog/our-fynbos-family-meet-adrian": Record<string, never>;
    "/blog/our-fynbos-family-meet-cairin": Record<string, never>;
    "/blog/our-fynbos-family-meet-justin": Record<string, never>;
    "/blog/why-payment-pointers-are-urls": Record<string, never>;
    "/blog/our-fynbos-family-meet-matt": Record<string, never>;
    "/blog/our-fynbos-family-meet-omer": Record<string, never>;
    "/blog/our-fynbos-family-meet-don": Record<string, never>;
    "/blog/card-payments-still-suck": Record<string, never>;
    "/legal/accessibility-statement": Record<string, never>;
    "/settings/profile-public/name": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/connections/add-a-public-key": Record<string, never>;
    "/legal/us/e-sign-agreement": Record<string, never>;
    "/connections/:connectionId": { "connectionId": string };
    "/settings/profile-personal": Record<string, never>;
    "/legal/us/privacy-policy": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/profile-contact": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/settings/profile-public": Record<string, never>;
    "/legal/terms-of-service": Record<string, never>;
    "/legal/us/terms-of-use": Record<string, never>;
    "/blog/joining-the-owf": Record<string, never>;
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
    "/identities-verify": Record<string, never>;
    "/waitlist/success": Record<string, never>;
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
    "/connections": Record<string, never>;
    "/pay/amount": Record<string, never>;
    "/identities": Record<string, never>;
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
      | ["/blog/the-future-digital-wallets-and-payment-pointers"]
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/blog/how-technical-standards-promote-innovation"]
      | ["/blog/connecting-the-internet-economy"]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
      | ["/blog/our-fynbos-family-meet-barnard"]
      | ["/blog/our-fynbos-family-meet-adrian"]
      | ["/blog/our-fynbos-family-meet-cairin"]
      | ["/blog/our-fynbos-family-meet-justin"]
      | ["/blog/why-payment-pointers-are-urls"]
      | ["/blog/our-fynbos-family-meet-matt"]
      | ["/blog/our-fynbos-family-meet-omer"]
      | ["/blog/our-fynbos-family-meet-don"]
      | ["/blog/card-payments-still-suck"]
      | ["/legal/accessibility-statement"]
      | ["/settings/profile-public/name"]
      | ["/api/maps/placesAutocomplete"]
      | ["/connections/add-a-public-key"]
      | ["/legal/us/e-sign-agreement"]
      | ["/connections/:connectionId", RouteParams["/connections/:connectionId"]]
      | ["/settings/profile-personal"]
      | ["/legal/us/privacy-policy"]
      | ["/settings/linked-accounts"]
      | ["/settings/profile-contact"]
      | ["/what-is-a-payment-pointer"]
      | ["/settings/profile-public"]
      | ["/legal/terms-of-service"]
      | ["/legal/us/terms-of-use"]
      | ["/blog/joining-the-owf"]
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
      | ["/identities-verify"]
      | ["/waitlist/success"]
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
      | ["/connections"]
      | ["/pay/amount"]
      | ["/identities"]
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
