import { useState } from "react";
import { Link } from "react-router-dom";
import clsx from "clsx";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { useAuth } from "@/app/providers";
import { env } from "@/shared/config";
import {
  confirmOtpSchema,
  ConfirmOtpValues,
  loginSchema,
  LoginFormValues,
  phoneLoginSchema,
  PhoneLoginFormValues,
  sendOtpSchema,
  SendOtpValues,
} from "@/features/auth/schema";
import { useConfirmOtp, useLogin, usePhoneLogin, useSendOtp } from "@/features/auth/hooks";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./AuthPage.module.scss";

export const AuthPage = () => {
  const auth = useAuth();
  const [notice, setNotice] = useState<string>("");
  const [noticeType, setNoticeType] = useState<"success" | "error" | "info">("info");
  const loginMutation = useLogin();
  const sendOtpMutation = useSendOtp();
  const confirmOtpMutation = useConfirmOtp();
  const phoneLoginMutation = usePhoneLogin();

  const loginForm = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { is_email: true },
  });

  const phoneLoginForm = useForm<PhoneLoginFormValues>({
    resolver: zodResolver(phoneLoginSchema),
  });

  const sendOtpForm = useForm<SendOtpValues>({
    resolver: zodResolver(sendOtpSchema),
  });

  const confirmOtpForm = useForm<ConfirmOtpValues>({
    resolver: zodResolver(confirmOtpSchema),
  });

  const handleLogin = async (data: LoginFormValues) => {
    setNoticeType("info");
    setNotice("Signing in...");
    try {
      const response = await loginMutation.mutateAsync(data);
      auth.login(response);
      setNoticeType("success");
      setNotice("Signed in successfully.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Sign in failed."));
    }
  };

  const handleConfirmOtp = async (data: ConfirmOtpValues) => {
    setNoticeType("info");
    setNotice("Confirming OTP...");
    try {
      const response = await confirmOtpMutation.mutateAsync(data);
      auth.login(response);
      setNoticeType("success");
      setNotice("Phone login completed.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "OTP confirmation failed."));
    }
  };

  const handlePhoneLogin = async (data: PhoneLoginFormValues) => {
    setNoticeType("info");
    setNotice("Signing in by phone...");
    try {
      const response = await phoneLoginMutation.mutateAsync(data);
      auth.login(response);
      setNoticeType("success");
      setNotice("Signed in by phone.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Phone sign in failed."));
    }
  };

  const handleGoogleLogin = () => {
    window.location.assign(`${env.ssoPublicUrl}/users/google/login`);
  };

  const handleSendOtp = async (data: SendOtpValues) => {
    setNoticeType("info");
    setNotice("Sending OTP...");
    try {
      await sendOtpMutation.mutateAsync(data);
      setNoticeType("success");
      setNotice("OTP sent.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Failed to send OTP."));
    }
  };

  return (
    <div className="container page">
      <div className={styles.header}>
        <div>
          <h1>Sign in</h1>
          <p className={styles.subtitle}>Access your tickets, bookings, and profile.</p>
        </div>
        {notice && (
          <span
            className={clsx(styles.notice, {
              [styles.noticeSuccess]: noticeType === "success",
              [styles.noticeError]: noticeType === "error",
              [styles.noticeInfo]: noticeType === "info",
            })}
          >
            {notice}
          </span>
        )}
      </div>

      <div className={styles.links}>
        <Link to="/register" className={styles.linkButton}>Create account</Link>
        <Link to="/reset" className={styles.linkButton}>Forgot password?</Link>
      </div>

      <section className={styles.grid}>
        <Card className={styles.card}>
          <h2>Email / login</h2>
          <form className={styles.form} onSubmit={loginForm.handleSubmit(handleLogin)}>
            <Button type="button" variant="outline" onClick={handleGoogleLogin}>
              Sign in with Google
            </Button>
            <label className={styles.checkbox}>
              <input type="checkbox" {...loginForm.register("is_email")} />
              Login with email
            </label>
            <Input
              label="Email or login"
              error={loginForm.formState.errors.identifier?.message}
              {...loginForm.register("identifier")}
            />
            <Input
              label="Password"
              type="password"
              error={loginForm.formState.errors.password?.message}
              {...loginForm.register("password")}
            />
            <Button type="submit" disabled={loginMutation.isPending}>
              {loginMutation.isPending ? "Signing in..." : "Sign in"}
            </Button>
          </form>
        </Card>

        <Card className={styles.card}>
          <h2>Phone + password</h2>
          <form className={styles.form} onSubmit={phoneLoginForm.handleSubmit(handlePhoneLogin)}>
            <Input label="Phone" {...phoneLoginForm.register("phone")} />
            <Input label="Password" type="password" {...phoneLoginForm.register("password")} />
            <Button type="submit" disabled={phoneLoginMutation.isPending}>
              {phoneLoginMutation.isPending ? "Signing in..." : "Sign in"}
            </Button>
          </form>
        </Card>

        <Card className={clsx(styles.card, styles.full)}>
          <h2>Phone OTP login</h2>
          <form className={styles.form} onSubmit={sendOtpForm.handleSubmit(handleSendOtp)}>
            <Input label="Phone" {...sendOtpForm.register("phone")} />
            <Button type="submit" disabled={sendOtpMutation.isPending}>
              {sendOtpMutation.isPending ? "Sending..." : "Send OTP"}
            </Button>
          </form>
          <form className={styles.form} onSubmit={confirmOtpForm.handleSubmit(handleConfirmOtp)}>
            <Input label="Phone" {...confirmOtpForm.register("phone")} />
            <Input label="Code" type="number" {...confirmOtpForm.register("code", { valueAsNumber: true })} />
            <Button type="submit" disabled={confirmOtpMutation.isPending}>
              {confirmOtpMutation.isPending ? "Confirming..." : "Confirm OTP"}
            </Button>
          </form>
        </Card>
      </section>
    </div>
  );
};
