import "axios";
import type { AuthMode } from "./http";

declare module "axios" {
  export interface AxiosRequestConfig {
    authMode?: AuthMode;
  }
}
