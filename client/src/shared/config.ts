export const env = {
  appName: import.meta.env.VITE_APP_NAME ?? "Ticket Master",
  ssoBaseUrl: import.meta.env.VITE_SSO_BASE_URL ?? "/api/sso",
  ssoPublicUrl: import.meta.env.VITE_SSO_PUBLIC_URL ?? import.meta.env.VITE_SSO_BASE_URL ?? "/api/sso",
  eventBaseUrl: import.meta.env.VITE_EVENT_BASE_URL ?? "/api/event/api/v1",
  bookingBaseUrl: import.meta.env.VITE_BOOKING_BASE_URL ?? "/api/booking/api/v1",
  searchingBaseUrl: import.meta.env.VITE_SEARCHING_BASE_URL ?? "/api/searching/api/v1",
  walletBaseUrl: import.meta.env.VITE_WALLET_BASE_URL ?? "/api/wallet/api/v1",
};
