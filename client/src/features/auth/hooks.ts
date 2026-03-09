import { useMutation } from "@tanstack/react-query";
import { AuthResponse } from "@/entities/user";
import {
  confirmOtpPhone,
  confirmResetEmail,
  confirmResetPhone,
  loginByPhone,
  loginUser,
  registerUser,
  resetEmail,
  resetPhone,
  sendOtpPhone,
} from "./api";
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

export const useLogin = () =>
  useMutation<AuthResponse, Error, LoginFormValues>({
    mutationFn: loginUser,
  });

export const useRegister = () =>
  useMutation<AuthResponse, Error, RegisterFormValues>({
    mutationFn: registerUser,
  });

export const usePhoneLogin = () =>
  useMutation<AuthResponse, Error, PhoneLoginFormValues>({
    mutationFn: loginByPhone,
  });

export const useSendOtp = () =>
  useMutation<void, Error, SendOtpValues>({
    mutationFn: sendOtpPhone,
  });

export const useConfirmOtp = () =>
  useMutation<AuthResponse, Error, ConfirmOtpValues>({
    mutationFn: confirmOtpPhone,
  });

export const useResetPhone = () =>
  useMutation<void, Error, ResetPhoneValues>({
    mutationFn: resetPhone,
  });

export const useConfirmResetPhone = () =>
  useMutation<void, Error, ResetPhoneConfirmValues>({
    mutationFn: confirmResetPhone,
  });

export const useResetEmail = () =>
  useMutation<void, Error, ResetEmailValues>({
    mutationFn: resetEmail,
  });

export const useConfirmResetEmail = () =>
  useMutation<void, Error, ResetEmailConfirmValues>({
    mutationFn: confirmResetEmail,
  });
