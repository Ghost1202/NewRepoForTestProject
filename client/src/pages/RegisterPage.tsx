import { useState } from "react";
import { Link } from "react-router-dom";
import clsx from "clsx";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { useAuth } from "@/app/providers";
import { registerSchema, RegisterFormValues } from "@/features/auth/schema";
import { useRegister } from "@/features/auth/hooks";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./AuthPage.module.scss";

export const RegisterPage = () => {
  const auth = useAuth();
  const [notice, setNotice] = useState<string>("");
  const [noticeType, setNoticeType] = useState<"success" | "error" | "info">("info");
  const registerMutation = useRegister();

  const registerForm = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
  });

  const handleRegister = async (data: RegisterFormValues) => {
    setNoticeType("info");
    setNotice("Creating account...");
    try {
      const response = await registerMutation.mutateAsync(data);
      if (response.token || response.access_token) {
        auth.login(response);
        setNoticeType("success");
        setNotice("Account created and signed in.");
      } else {
        setNoticeType("success");
        setNotice("Account created. Please sign in.");
      }
    } catch (error) {
      setNoticeType("error");
      setNotice(getErrorMessage(error, "Registration failed."));
    }
  };

  return (
    <div className="container page">
      <div className={styles.header}>
        <div>
          <h1>Create account</h1>
          <p className={styles.subtitle}>Join the platform and start booking tickets.</p>
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
        <Link to="/auth">Already have an account? Sign in</Link>
      </div>

      <Card className={clsx(styles.card, styles.single)}>
        <h2>Register</h2>
        <form className={styles.form} onSubmit={registerForm.handleSubmit(handleRegister)}>
          <Input label="First name" {...registerForm.register("name")} />
          <Input label="Last name" {...registerForm.register("last_name")} />
          <Input label="Login" {...registerForm.register("login")} />
          <Input label="Email" type="email" {...registerForm.register("email")} />
          <Input label="Password" type="password" {...registerForm.register("password")} />
          <Button type="submit" disabled={registerMutation.isPending}>
            {registerMutation.isPending ? "Creating..." : "Create account"}
          </Button>
        </form>
      </Card>
    </div>
  );
};
