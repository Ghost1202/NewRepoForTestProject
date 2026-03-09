import { z } from "zod";

export const loginSchema = z.object({
  is_email: z.boolean(),
  identifier: z.string().min(3),
  password: z.string().min(6),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

export const registerSchema = z.object({
  name: z.string().min(1),
  last_name: z.string().min(1),
  login: z.string().min(3),
  email: z.string().email(),
  password: z.string().min(8),
});

export type RegisterFormValues = z.infer<typeof registerSchema>;

export const phoneLoginSchema = z.object({
  phone: z.string().min(8),
  password: z.string().min(6),
});

export type PhoneLoginFormValues = z.infer<typeof phoneLoginSchema>;

export const sendOtpSchema = z.object({
  phone: z.string().min(8),
});

export type SendOtpValues = z.infer<typeof sendOtpSchema>;

export const confirmOtpSchema = z.object({
  phone: z.string().min(8),
  code: z.number().int().min(0),
});

export type ConfirmOtpValues = z.infer<typeof confirmOtpSchema>;

export const resetPhoneSchema = z.object({
  phone: z.string().min(8),
});

export type ResetPhoneValues = z.infer<typeof resetPhoneSchema>;

export const resetPhoneConfirmSchema = z.object({
  phone: z.string().min(8),
  code: z.number().int().min(0),
  new_password: z.string().min(8),
});

export type ResetPhoneConfirmValues = z.infer<typeof resetPhoneConfirmSchema>;

export const resetEmailSchema = z.object({
  email: z.string().email(),
});

export type ResetEmailValues = z.infer<typeof resetEmailSchema>;

export const resetEmailConfirmSchema = z.object({
  email: z.string().email(),
  code: z.number().int().min(0),
  new_password: z.string().min(8),
});

export type ResetEmailConfirmValues = z.infer<typeof resetEmailConfirmSchema>;
