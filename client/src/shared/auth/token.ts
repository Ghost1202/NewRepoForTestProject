const TOKEN_KEY = "auth_token";

export const authStore = {
  getToken: () => localStorage.getItem(TOKEN_KEY),
  setToken: (token: string) => localStorage.setItem(TOKEN_KEY, token),
  clearToken: () => localStorage.removeItem(TOKEN_KEY),
};

type LogoutHandler = () => void;

export const authEvents = {
  emitLogout: () => window.dispatchEvent(new CustomEvent("auth:logout")),
  onLogout: (handler: LogoutHandler) => {
    window.addEventListener("auth:logout", handler);
    return () => window.removeEventListener("auth:logout", handler);
  },
};
