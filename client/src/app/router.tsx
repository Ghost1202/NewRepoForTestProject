import { PropsWithChildren } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { RootLayout } from "./layouts/RootLayout";
import {
  AdminDashboard,
  AuthCallbackPage,
  AuthPage,
  CheckoutPage,
  EventDetailsPage,
  HomePage,
  ProfilePage,
  RegisterPage,
  ResetPasswordPage,
} from "@/pages";
import { useAuth } from "./providers";
import { Spinner } from "@/shared/ui/Spinner/Spinner";

const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="container page">
        <Spinner label="Loading profile..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth" replace />;
  }

  return <>{children}</>;
};

const NotFoundPage = () => {
  return (
    <div className="container page">
      <h2>Page not found</h2>
      <p>The page you requested does not exist.</p>
    </div>
  );
};

export const AppRouter = () => {
  return (
    <Routes>
      <Route element={<RootLayout />}>
        <Route index element={<HomePage />} />
        <Route path="/event/:eventId" element={<EventDetailsPage />} />
        <Route path="/checkout" element={<CheckoutPage />} />
        <Route path="/profile" element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
        <Route path="/admin" element={<ProtectedRoute><AdminDashboard /></ProtectedRoute>} />
        <Route path="/auth" element={<AuthPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/reset" element={<ResetPasswordPage />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
};
