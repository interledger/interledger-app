declare module "routes-gen" {
  export type RouteParams = {
    "/api/maps/geocode": Record<string, never>;
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/blog": Record<string, never>;
    "/blog/card-payments-still-suck": Record<string, never>;
    "/blog/connecting-the-internet-economy": Record<string, never>;
    "/blog/how-technical-standards-promote-innovation": Record<string, never>;
    "/blog/joining-the-owf": Record<string, never>;
    "/blog/our-fynbos-family-meet-adrian": Record<string, never>;
    "/blog/our-fynbos-family-meet-barnard": Record<string, never>;
    "/blog/our-fynbos-family-meet-cairin": Record<string, never>;
    "/blog/our-fynbos-family-meet-don": Record<string, never>;
    "/blog/our-fynbos-family-meet-justin": Record<string, never>;
    "/blog/our-fynbos-family-meet-matt": Record<string, never>;
    "/blog/our-fynbos-family-meet-omer": Record<string, never>;
    "/blog/the-future-digital-wallets-and-payment-pointers": Record<string, never>;
    "/blog/why-payment-pointers-are-urls": Record<string, never>;
    "/connections": Record<string, never>;
    "/connections/:connectionId": { "connectionId": string };
    "/connections/add-a-public-key": Record<string, never>;
    "/contact": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/contacts": Record<string, never>;
    "/": Record<string, never>;
    "/legal": Record<string, never>;
    "/legal/electronic-disclosures": Record<string, never>;
    "/legal/privacy-policy": Record<string, never>;
    "/legal/terms-of-use": Record<string, never>;
    "/linked-account/:type": { "type": string };
    "/linked-account/:type/almost-there": { "type": string };
    "/linked-account/:type/success": { "type": string };
    "/linked-account/bank/widget": Record<string, never>;
    "/linked-account/card/widget": Record<string, never>;
    "/login": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/logout": Record<string, never>;
    "/pay": Record<string, never>;
    "/pay/amount": Record<string, never>;
    "/pay/confirm": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/personal-details/about": Record<string, never>;
    "/personal-details/address": Record<string, never>;
    "/recovery": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/linked-accounts/:accountId": { "accountId": string };
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
    "/support/success": Record<string, never>;
    "/temp-cloudflare-error": Record<string, never>;
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/transactions": Record<string, never>;
    "/verify": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/maps/geocode"]
      | ["/api/maps/placesAutocomplete"]
      | ["/api/sendOtp"]
      | ["/blog"]
      | ["/blog/card-payments-still-suck"]
      | ["/blog/connecting-the-internet-economy"]
      | ["/blog/how-technical-standards-promote-innovation"]
      | ["/blog/joining-the-owf"]
      | ["/blog/our-fynbos-family-meet-adrian"]
      | ["/blog/our-fynbos-family-meet-barnard"]
      | ["/blog/our-fynbos-family-meet-cairin"]
      | ["/blog/our-fynbos-family-meet-don"]
      | ["/blog/our-fynbos-family-meet-justin"]
      | ["/blog/our-fynbos-family-meet-matt"]
      | ["/blog/our-fynbos-family-meet-omer"]
      | ["/blog/the-future-digital-wallets-and-payment-pointers"]
      | ["/blog/why-payment-pointers-are-urls"]
      | ["/connections"]
      | ["/connections/:connectionId", RouteParams["/connections/:connectionId"]]
      | ["/connections/add-a-public-key"]
      | ["/contact"]
      | ["/contact/success"]
      | ["/contacts"]
      | ["/"]
      | ["/legal"]
      | ["/legal/electronic-disclosures"]
      | ["/legal/privacy-policy"]
      | ["/legal/terms-of-use"]
      | ["/linked-account/:type", RouteParams["/linked-account/:type"]]
      | ["/linked-account/:type/almost-there", RouteParams["/linked-account/:type/almost-there"]]
      | ["/linked-account/:type/success", RouteParams["/linked-account/:type/success"]]
      | ["/linked-account/bank/widget"]
      | ["/linked-account/card/widget"]
      | ["/login"]
      | ["/login/challenge"]
      | ["/logout"]
      | ["/pay"]
      | ["/pay/amount"]
      | ["/pay/confirm"]
      | ["/payment-pointer"]
      | ["/personal-details"]
      | ["/personal-details/about"]
      | ["/personal-details/address"]
      | ["/recovery"]
      | ["/recovery/password"]
      | ["/settings"]
      | ["/settings/linked-accounts"]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
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
      | ["/support/success"]
      | ["/temp-cloudflare-error"]
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/transactions"]
      | ["/verify"]
      | ["/waitlist"]
      | ["/waitlist/success"]
      | ["/what-is-a-payment-pointer"]
  >(...args: T): typeof args[0];
}
