import { ssoApi } from "@/shared/api";
import { AuthResponse, User } from "@/entities/user";
import {
  ConfirmOtpValues,
  LoginFormValues,
  PhoneLoginFormValues,
  RegisterFormValues,
  ResetEmailConfirmValues,
  ResetEmailValues,
  ResetPhoneConfirmValues,
  ResetPhoneValues,
  SendOtpValues,
} from "./schema";

export const registerUser = async (payload: RegisterFormValues) => {
  const response = await ssoApi.post<AuthResponse>("/users/register", payload, { authMode: "none" });
  return response.data;
};

export const loginUser = async (payload: LoginFormValues) => {
  const response = await ssoApi.post<AuthResponse>("/users/auth", payload, { authMode: "none" });
  return response.data;
};

export const loginByPhone = async (payload: PhoneLoginFormValues) => {
  const response = await ssoApi.post<AuthResponse>("/users/auth/phone", payload, { authMode: "none" });
  return response.data;
};

export const sendOtpPhone = async (payload: SendOtpValues) => {
  const response = await ssoApi.post<void>("/users/otp/phone/send", payload, { authMode: "none" });
  return response.data;
};

export const confirmOtpPhone = async (payload: ConfirmOtpValues) => {
  const response = await ssoApi.post<AuthResponse>("/users/otp/phone/confirm", payload, { authMode: "none" });
  return response.data;
};

export const resetPhone = async (payload: ResetPhoneValues) => {
  const response = await ssoApi.post<void>("/users/reset/phone", payload, { authMode: "none" });
  return response.data;
};

export const confirmResetPhone = async (payload: ResetPhoneConfirmValues) => {
  const response = await ssoApi.post<void>("/users/reset/phone/confirm", payload, { authMode: "none" });
  return response.data;
};

export const resetEmail = async (payload: ResetEmailValues) => {
  const response = await ssoApi.post<void>("/users/reset/email", payload, { authMode: "none" });
  return response.data;
};

export const confirmResetEmail = async (payload: ResetEmailConfirmValues) => {
  const response = await ssoApi.post<void>("/users/reset/email/confirm", payload, { authMode: "none" });
  return response.data;
};

export const getProfile = async () => {
  const response = await ssoApi.get<User>("/users/login", { authMode: "required" });
  return response.data;
};
