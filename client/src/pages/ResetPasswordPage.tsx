import { useState } from "react";
import { Link } from "react-router-dom";
import clsx from "clsx";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import {
  resetEmailConfirmSchema,
  ResetEmailConfirmValues,
  resetEmailSchema,
  ResetEmailValues,
  resetPhoneConfirmSchema,
  ResetPhoneConfirmValues,
  resetPhoneSchema,
  ResetPhoneValues,
} from "@/features/auth/schema";
import { useConfirmResetEmail, useConfirmResetPhone, useResetEmail, useResetPhone } from "@/features/auth/hooks";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./AuthPage.module.scss";

export const ResetPasswordPage = () => {
  const [notice, setNotice] = useState<string>("");
  const [noticeType, setNoticeType] = useState<"success" | "error" | "info">("info");
  const resetPhoneMutation = useResetPhone();
  const resetPhoneConfirmMutation = useConfirmResetPhone();
  const resetEmailMutation = useResetEmail();
  const resetEmailConfirmMutation = useConfirmResetEmail();

  const resetPhoneForm = useForm<ResetPhoneValues>({
    resolver: zodResolver(resetPhoneSchema),
  });

  const resetPhoneConfirmForm = useForm<ResetPhoneConfirmValues>({
    resolver: zodResolver(resetPhoneConfirmSchema),
  });

  const resetEmailForm = useForm<ResetEmailValues>({
    resolver: zodResolver(resetEmailSchema),
  });

  const resetEmailConfirmForm = useForm<ResetEmailConfirmValues>({
    resolver: zodResolver(resetEmailConfirmSchema),
  });

  const handleResetPhone = async (data: ResetPhoneValues) => {
    setNoticeType("info");
    setNotice("Sending reset OTP...");
    try {
      await resetPhoneMutation.mutateAsync(data);
      setNoticeType("success");
      setNotice("Reset OTP sent.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Failed to send reset OTP."));
    }
  };

  const handleResetPhoneConfirm = async (data: ResetPhoneConfirmValues) => {
    setNoticeType("info");
    setNotice("Confirming reset...");
    try {
      await resetPhoneConfirmMutation.mutateAsync(data);
      setNoticeType("success");
      setNotice("Phone password reset.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Failed to confirm reset."));
    }
  };

  const handleResetEmail = async (data: ResetEmailValues) => {
    setNoticeType("info");
    setNotice("Sending email OTP...");
    try {
      await resetEmailMutation.mutateAsync(data);
      setNoticeType("success");
      setNotice("Email OTP sent.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Failed to send email OTP."));
    }
  };

  const handleResetEmailConfirm = async (data: ResetEmailConfirmValues) => {
    setNoticeType("info");
    setNotice("Confirming reset...");
    try {
      await resetEmailConfirmMutation.mutateAsync(data);
      setNoticeType("success");
      setNotice("Email password reset.");
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Failed to confirm reset."));
    }
  };

  return (
    <div className="container page">
      <div className={styles.header}>
        <div>
          <h1>Reset password</h1>
          <p className={styles.subtitle}>Recover access by phone or email.</p>
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
        <Link to="/auth">Back to sign in</Link>
      </div>

      <Card className={styles.card}>
        <h2>Reset options</h2>
        <div className={styles.resetGrid}>
          <div>
            <h3>By phone</h3>
            <form className={styles.form} onSubmit={resetPhoneForm.handleSubmit(handleResetPhone)}>
              <Input label="Phone" {...resetPhoneForm.register("phone")} />
              <Button type="submit" disabled={resetPhoneMutation.isPending}>
                {resetPhoneMutation.isPending ? "Sending..." : "Send OTP"}
              </Button>
            </form>
            <form className={styles.form} onSubmit={resetPhoneConfirmForm.handleSubmit(handleResetPhoneConfirm)}>
              <Input label="Phone" {...resetPhoneConfirmForm.register("phone")} />
              <Input label="Code" type="number" {...resetPhoneConfirmForm.register("code", { valueAsNumber: true })} />
              <Input label="New password" type="password" {...resetPhoneConfirmForm.register("new_password")} />
              <Button type="submit" disabled={resetPhoneConfirmMutation.isPending}>
                {resetPhoneConfirmMutation.isPending ? "Confirming..." : "Confirm reset"}
              </Button>
            </form>
          </div>
          <div>
            <h3>By email</h3>
            <form className={styles.form} onSubmit={resetEmailForm.handleSubmit(handleResetEmail)}>
              <Input label="Email" type="email" {...resetEmailForm.register("email")} />
              <Button type="submit" disabled={resetEmailMutation.isPending}>
                {resetEmailMutation.isPending ? "Sending..." : "Send OTP"}
              </Button>
            </form>
            <form className={styles.form} onSubmit={resetEmailConfirmForm.handleSubmit(handleResetEmailConfirm)}>
              <Input label="Email" type="email" {...resetEmailConfirmForm.register("email")} />
              <Input label="Code" type="number" {...resetEmailConfirmForm.register("code", { valueAsNumber: true })} />
              <Input label="New password" type="password" {...resetEmailConfirmForm.register("new_password")} />
              <Button type="submit" disabled={resetEmailConfirmMutation.isPending}>
                {resetEmailConfirmMutation.isPending ? "Confirming..." : "Confirm reset"}
              </Button>
            </form>
          </div>
        </div>
      </Card>
    </div>
  );
};
