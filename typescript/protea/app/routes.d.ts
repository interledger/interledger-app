declare module "routes-gen" {
  export type RouteParams = {
    "/blog/the-future-digital-wallets-and-payment-pointers": Record<string, never>;
    "/transaction/machnet_wallet_withdrawal/:transactionId": { "transactionId": string };
    "/transaction/open_payments_incoming/:transactionId": { "transactionId": string };
    "/transaction/open_payments_outgoing/:transactionId": { "transactionId": string };
    "/blog/how-technical-standards-promote-innovation": Record<string, never>;
    "/transaction/machnet_wallet_topup/:transactionId": { "transactionId": string };
    "/blog/connecting-the-internet-economy": Record<string, never>;
    "/blog/our-fynbos-family-meet-barnard": Record<string, never>;
    "/settings/linked-accounts/:accountId": { "accountId": string };
    "/blog/our-fynbos-family-meet-adrian": Record<string, never>;
    "/blog/our-fynbos-family-meet-cairin": Record<string, never>;
    "/blog/our-fynbos-family-meet-justin": Record<string, never>;
    "/blog/why-payment-pointers-are-urls": Record<string, never>;
    "/linked-account/:type/almost-there": { "type": string };
    "/blog/our-fynbos-family-meet-matt": Record<string, never>;
    "/blog/our-fynbos-family-meet-omer": Record<string, never>;
    "/blog/our-fynbos-family-meet-don": Record<string, never>;
    "/blog/card-payments-still-suck": Record<string, never>;
    "/legal/electronic-disclosures": Record<string, never>;
    "/linked-account/:type/success": { "type": string };
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/linked-account/:type/widget": { "type": string };
    "/settings/personal-details": Record<string, never>;
    "/what-is-a-payment-pointer": Record<string, never>;
    "/personal-details/address": Record<string, never>;
    "/settings/linked-accounts": Record<string, never>;
    "/settings/account-limits": Record<string, never>;
    "/personal-details/about": Record<string, never>;
    "/legal/privacy-policy": Record<string, never>;
    "/linked-account/:type": { "type": string };
    "/legal/terms-of-use": Record<string, never>;
    "/recovery/password": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/personal-details": Record<string, never>;
    "/waitlist/success": Record<string, never>;
    "/withdraw/confirm": Record<string, never>;
    "/contact/success": Record<string, never>;
    "/deposit/confirm": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/payment-pointer": Record<string, never>;
    "/signup/password": Record<string, never>;
    "/support/success": Record<string, never>;
    "/machnet-terms": Record<string, never>;
    "/signup/about": Record<string, never>;
    "/signup/phone": Record<string, never>;
    "/transactions": Record<string, never>;
    "/api/sendOtp": Record<string, never>;
    "/pay/confirm": Record<string, never>;
    "/pay/amount": Record<string, never>;
    "/recovery": Record<string, never>;
    "/settings": Record<string, never>;
    "/waitlist": Record<string, never>;
    "/withdraw": Record<string, never>;
    "/contact": Record<string, never>;
    "/deposit": Record<string, never>;
    "/support": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/": Record<string, never>;
    "/legal": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/pay": Record<string, never>;
  };

  export function route<
    T extends
      | ["/blog/the-future-digital-wallets-and-payment-pointers"]
      | ["/transaction/machnet_wallet_withdrawal/:transactionId", RouteParams["/transaction/machnet_wallet_withdrawal/:transactionId"]]
      | ["/transaction/open_payments_incoming/:transactionId", RouteParams["/transaction/open_payments_incoming/:transactionId"]]
      | ["/transaction/open_payments_outgoing/:transactionId", RouteParams["/transaction/open_payments_outgoing/:transactionId"]]
      | ["/blog/how-technical-standards-promote-innovation"]
      | ["/transaction/machnet_wallet_topup/:transactionId", RouteParams["/transaction/machnet_wallet_topup/:transactionId"]]
      | ["/blog/connecting-the-internet-economy"]
      | ["/blog/our-fynbos-family-meet-barnard"]
      | ["/settings/linked-accounts/:accountId", RouteParams["/settings/linked-accounts/:accountId"]]
      | ["/blog/our-fynbos-family-meet-adrian"]
      | ["/blog/our-fynbos-family-meet-cairin"]
      | ["/blog/our-fynbos-family-meet-justin"]
      | ["/blog/why-payment-pointers-are-urls"]
      | ["/linked-account/:type/almost-there", RouteParams["/linked-account/:type/almost-there"]]
      | ["/blog/our-fynbos-family-meet-matt"]
      | ["/blog/our-fynbos-family-meet-omer"]
      | ["/blog/our-fynbos-family-meet-don"]
      | ["/blog/card-payments-still-suck"]
      | ["/legal/electronic-disclosures"]
      | ["/linked-account/:type/success", RouteParams["/linked-account/:type/success"]]
      | ["/api/maps/placesAutocomplete"]
      | ["/linked-account/:type/widget", RouteParams["/linked-account/:type/widget"]]
      | ["/settings/personal-details"]
      | ["/what-is-a-payment-pointer"]
      | ["/personal-details/address"]
      | ["/settings/linked-accounts"]
      | ["/settings/account-limits"]
      | ["/personal-details/about"]
      | ["/legal/privacy-policy"]
      | ["/linked-account/:type", RouteParams["/linked-account/:type"]]
      | ["/legal/terms-of-use"]
      | ["/recovery/password"]
      | ["/settings/password"]
      | ["/api/maps/geocode"]
      | ["/personal-details"]
      | ["/waitlist/success"]
      | ["/withdraw/confirm"]
      | ["/contact/success"]
      | ["/deposit/confirm"]
      | ["/login/challenge"]
      | ["/payment-pointer"]
      | ["/signup/password"]
      | ["/support/success"]
      | ["/machnet-terms"]
      | ["/signup/about"]
      | ["/signup/phone"]
      | ["/transactions"]
      | ["/api/sendOtp"]
      | ["/pay/confirm"]
      | ["/pay/amount"]
      | ["/recovery"]
      | ["/settings"]
      | ["/waitlist"]
      | ["/withdraw"]
      | ["/contact"]
      | ["/deposit"]
      | ["/support"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/"]
      | ["/legal"]
      | ["/login"]
      | ["/blog"]
      | ["/pay"]
  >(...args: T): typeof args[0];
}
