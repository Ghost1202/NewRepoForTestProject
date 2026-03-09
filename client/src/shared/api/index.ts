import { env } from "../config";
import { createApiClient } from "./http";

export const ssoApi = createApiClient(env.ssoBaseUrl);
export const eventApi = createApiClient(env.eventBaseUrl);
export const bookingApi = createApiClient(env.bookingBaseUrl);
export const searchingApi = createApiClient(env.searchingBaseUrl);
export const walletApi = createApiClient(env.walletBaseUrl);
