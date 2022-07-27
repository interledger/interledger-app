declare module "routes-gen" {
  export type RouteParams = {
    "/api/maps/placesAutocomplete": Record<string, never>;
    "/onboarding/country-access": Record<string, never>;
    "/confirmation/:flowId": { "flowId": string };
    "/confirmation/:flowId/payment-method": { "flowId": string };
    "/confirmation/:flowId/withdraw": { "flowId": string };
    "/confirmation/:flowId/deposit": { "flowId": string };
    "/confirmation/:flowId/send": { "flowId": string };
    "/recovery/password": Record<string, never>;
    "/api/maps/geocode": Record<string, never>;
    "/login/challenge": Record<string, never>;
    "/onboarding/unit": Record<string, never>;
    "/": Record<string, never>;
    "/activity/transaction/:id": { "id": string };
    "/settings/payment-methods": Record<string, never>;
    "/settings/password": Record<string, never>;
    "/activity/filter": Record<string, never>;
    "/activity": Record<string, never>;
    "/settings": Record<string, never>;
    "/connect": Record<string, never>;
    "/receive": Record<string, never>;
    "/home": Record<string, never>;
    "/flows/:flowId": { "flowId": string };
    "/flows/:flowId/unit-onboarding/address": { "flowId": string };
    "/flows/:flowId/withdraw/payment-method": { "flowId": string };
    "/flows/:flowId/deposit/payment-method": { "flowId": string };
    "/flows/:flowId/payment-method/details": { "flowId": string };
    "/flows/:flowId/payment-method/review": { "flowId": string };
    "/flows/:flowId/unit-onboarding/about": { "flowId": string };
    "/flows/:flowId/payment-method/type": { "flowId": string };
    "/flows/:flowId/signup/password": { "flowId": string };
    "/flows/:flowId/withdraw/amount": { "flowId": string };
    "/flows/:flowId/withdraw/review": { "flowId": string };
    "/flows/:flowId/deposit/amount": { "flowId": string };
    "/flows/:flowId/deposit/review": { "flowId": string };
    "/flows/:flowId/signup/about": { "flowId": string };
    "/flows/:flowId/signup/phone": { "flowId": string };
    "/flows/:flowId/send/amount": { "flowId": string };
    "/flows/:flowId/send/review": { "flowId": string };
    "/flows/:flowId/signup/sms": { "flowId": string };
    "/flows/:flowId/send/to": { "flowId": string };
    "/disclosures": Record<string, never>;
    "/recovery": Record<string, never>;
    "/logout": Record<string, never>;
    "/signup": Record<string, never>;
    "/verify": Record<string, never>;
    "/login": Record<string, never>;
    "/blog": Record<string, never>;
    "/blog/connecting-the-internet-economy": Record<string, never>;
  };

  export function route<
    T extends
      | ["/api/maps/placesAutocomplete"]
      | ["/onboarding/country-access"]
      | ["/confirmation/:flowId", RouteParams["/confirmation/:flowId"]]
      | ["/confirmation/:flowId/payment-method", RouteParams["/confirmation/:flowId/payment-method"]]
      | ["/confirmation/:flowId/withdraw", RouteParams["/confirmation/:flowId/withdraw"]]
      | ["/confirmation/:flowId/deposit", RouteParams["/confirmation/:flowId/deposit"]]
      | ["/confirmation/:flowId/send", RouteParams["/confirmation/:flowId/send"]]
      | ["/recovery/password"]
      | ["/api/maps/geocode"]
      | ["/login/challenge"]
      | ["/onboarding/unit"]
      | ["/"]
      | ["/activity/transaction/:id", RouteParams["/activity/transaction/:id"]]
      | ["/settings/payment-methods"]
      | ["/settings/password"]
      | ["/activity/filter"]
      | ["/activity"]
      | ["/settings"]
      | ["/connect"]
      | ["/receive"]
      | ["/home"]
      | ["/flows/:flowId", RouteParams["/flows/:flowId"]]
      | ["/flows/:flowId/unit-onboarding/address", RouteParams["/flows/:flowId/unit-onboarding/address"]]
      | ["/flows/:flowId/withdraw/payment-method", RouteParams["/flows/:flowId/withdraw/payment-method"]]
      | ["/flows/:flowId/deposit/payment-method", RouteParams["/flows/:flowId/deposit/payment-method"]]
      | ["/flows/:flowId/payment-method/details", RouteParams["/flows/:flowId/payment-method/details"]]
      | ["/flows/:flowId/payment-method/review", RouteParams["/flows/:flowId/payment-method/review"]]
      | ["/flows/:flowId/unit-onboarding/about", RouteParams["/flows/:flowId/unit-onboarding/about"]]
      | ["/flows/:flowId/payment-method/type", RouteParams["/flows/:flowId/payment-method/type"]]
      | ["/flows/:flowId/signup/password", RouteParams["/flows/:flowId/signup/password"]]
      | ["/flows/:flowId/withdraw/amount", RouteParams["/flows/:flowId/withdraw/amount"]]
      | ["/flows/:flowId/withdraw/review", RouteParams["/flows/:flowId/withdraw/review"]]
      | ["/flows/:flowId/deposit/amount", RouteParams["/flows/:flowId/deposit/amount"]]
      | ["/flows/:flowId/deposit/review", RouteParams["/flows/:flowId/deposit/review"]]
      | ["/flows/:flowId/signup/about", RouteParams["/flows/:flowId/signup/about"]]
      | ["/flows/:flowId/signup/phone", RouteParams["/flows/:flowId/signup/phone"]]
      | ["/flows/:flowId/send/amount", RouteParams["/flows/:flowId/send/amount"]]
      | ["/flows/:flowId/send/review", RouteParams["/flows/:flowId/send/review"]]
      | ["/flows/:flowId/signup/sms", RouteParams["/flows/:flowId/signup/sms"]]
      | ["/flows/:flowId/send/to", RouteParams["/flows/:flowId/send/to"]]
      | ["/disclosures"]
      | ["/recovery"]
      | ["/logout"]
      | ["/signup"]
      | ["/verify"]
      | ["/login"]
      | ["/blog"]
      | ["/blog/connecting-the-internet-economy"]
  >(...args: T): typeof args[0];
}
