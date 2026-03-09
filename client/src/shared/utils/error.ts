export const getErrorMessage = (error: unknown, fallback = "Request failed") => {
  if (typeof error === "string") {
    return error;
  }

  if (error && typeof error === "object") {
    if ("message" in error && typeof (error as { message?: unknown }).message === "string") {
      return (error as { message: string }).message;
    }

    if ("response" in error) {
      const response = (error as { response?: { data?: unknown; status?: number } }).response;
      if (response?.data) {
        if (typeof response.data === "string") {
          return response.data;
        }
        if (typeof response.data === "object" && response.data !== null && "message" in response.data) {
          const message = (response.data as { message?: unknown }).message;
          if (typeof message === "string") {
            return message;
          }
        }
      }
      if (response?.status) {
        return `Request failed (status ${response.status})`;
      }
    }
  }

  return fallback;
};
