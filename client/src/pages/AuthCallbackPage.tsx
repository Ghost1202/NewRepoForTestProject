import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import clsx from "clsx";
import { Card } from "@/shared/ui/Card/Card";
import { Spinner } from "@/shared/ui/Spinner/Spinner";
import { Button } from "@/shared/ui/Button/Button";
import { User } from "@/entities/user";
import { useAuth } from "@/app/providers";
import { getErrorMessage } from "@/shared/utils/error";
import { ssoApi } from "@/shared/api";
import styles from "./AuthCallbackPage.module.scss";

type StatusType = "success" | "error" | "info";

type StatusState = {
  type: StatusType;
  message: string;
};

export const AuthCallbackPage = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const auth = useAuth();
  const [status, setStatus] = useState<StatusState>({
    type: "info",
    message: "Completing Google sign-in...",
  });
  const [isWorking, setIsWorking] = useState(true);

  useEffect(() => {
    let isActive = true;
    let redirectTimer: number | undefined;
    const params = new URLSearchParams(location.search);
    const token = params.get("token") ?? params.get("access_token");

    const finishSignIn = async () => {
      setIsWorking(true);
      setStatus({ type: "info", message: "Completing Google sign-in..." });

      if (!token) {
        auth.logout();
        setStatus({ type: "error", message: "Missing token in callback URL." });
        setIsWorking(false);
        return;
      }

      auth.login({ token });

      try {
        const response = await ssoApi.get<User>("/users/login", { authMode: "required" });
        if (!isActive) {
          return;
        }
        auth.login({ token, user: response.data });
        setStatus({ type: "success", message: "Signed in successfully. Redirecting..." });
        redirectTimer = window.setTimeout(() => {
          if (isActive) {
            navigate("/profile", { replace: true });
          }
        }, 600);
      } catch (error) {
        if (!isActive) {
          return;
        }
        setStatus({ type: "error", message: getErrorMessage(error, "Google sign-in failed.") });
      } finally {
        if (isActive) {
          setIsWorking(false);
        }
      }
    };

    void finishSignIn();

    return () => {
      isActive = false;
      if (redirectTimer) {
        window.clearTimeout(redirectTimer);
      }
    };
  }, [location.search, navigate]);

  return (
    <div className="container page">
      <Card className={styles.card}>
        <h1>Google sign-in</h1>
        <p className={styles.subtitle}>
          We are finishing your authentication with the backend.
        </p>
        {isWorking && <Spinner label="Checking session..." />}
        <p
          className={clsx("status", {
            statusSuccess: status.type === "success",
            statusError: status.type === "error",
            statusInfo: status.type === "info",
          })}
        >
          {status.message}
        </p>
        {status.type === "error" && (
          <div className={styles.actions}>
            <Button variant="outline" onClick={() => navigate("/auth", { replace: true })}>
              Back to sign in
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
};
