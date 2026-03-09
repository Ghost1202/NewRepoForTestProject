import axios, { AxiosHeaders, AxiosInstance, AxiosRequestConfig, AxiosRequestHeaders } from "axios";
import { authEvents, authStore } from "../auth/token";

export type AuthMode = "required" | "optional" | "none";

export type RequestConfig<D = unknown> = AxiosRequestConfig<D> & {
  authMode?: AuthMode;
};

export const createApiClient = (baseURL: string): AxiosInstance => {
  const client = axios.create({
    baseURL,
    headers: {
      "Content-Type": "application/json",
    },
    withCredentials: true,
  });

  client.interceptors.request.use((config) => {
    const configWithAuth = config as RequestConfig;
    const mode = configWithAuth.authMode ?? "optional";
    const token = authStore.getToken();

    if (mode !== "none" && token) {
      const headers = AxiosHeaders.from((config.headers ?? {}) as AxiosRequestHeaders);
      headers.set("Authorization", `Bearer ${token}`);
      config.headers = headers;
    }

    return config;
  });

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error?.response?.status === 401) {
        authStore.clearToken();
        authEvents.emitLogout();
      }
      return Promise.reject(error);
    },
  );

  return client;
};
