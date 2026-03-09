import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useCreateBooking, useCreateBookingWithWallet } from "@/features/booking/hooks";
import { bookingSchema, BookingFormValues } from "@/features/booking/schema";
import { useAuth } from "@/app/providers";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./CheckoutPage.module.scss";

type LocationState = {
  eventId: number;
  ticketIds: number[];
  promoCode?: string;
  earlyId?: number;
  bundleId?: number;
};

export const CheckoutPage = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();
  const booking = location.state as LocationState | null;
  const mutation = useCreateBooking();
  const walletMutation = useCreateBookingWithWallet();
  const [status, setStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);

  const {
    register,
    setValue,
    handleSubmit,
    formState: { errors },
  } = useForm<BookingFormValues>({
    resolver: zodResolver(bookingSchema),
    defaultValues: {
      ticket_ids: booking?.ticketIds ?? [],
      user_email: user?.email ?? "",
      promo_code: booking?.promoCode ?? "",
      early_id: booking?.earlyId ?? 0,
      bundle_id: booking?.bundleId ?? 0,
    },
  });

  useEffect(() => {
    register("ticket_ids");
    register("promo_code");
    register("early_id");
    register("bundle_id");
    if (booking) {
      setValue("ticket_ids", booking.ticketIds);
      setValue("promo_code", booking.promoCode ?? "");
      setValue("early_id", booking.earlyId ?? 0);
      setValue("bundle_id", booking.bundleId ?? 0);
    }
  }, [booking, register, setValue]);

  if (!booking) {
    return (
      <div className="container page">
        <h2>Checkout unavailable</h2>
        <p>Select tickets on the event page first.</p>
        <Button onClick={() => navigate("/")}>Back to events</Button>
      </div>
    );
  }

  const onSubmit = async (data: BookingFormValues) => {
    setStatus({ type: "info", message: "Creating booking..." });
    try {
      const response = await mutation.mutateAsync({
        ticket_ids: data.ticket_ids,
        user_email: data.user_email,
        promo_code: data.promo_code || "",
        early_id: data.early_id || 0,
        bundle_id: data.bundle_id || 0,
      });
      setStatus({ type: "success", message: "Booking created. Redirecting to payment..." });
      window.location.assign(response.payment_url);
    } catch (error) {
      setStatus({ type: "error", message: getErrorMessage(error, "Failed to create booking.") });
    }
  };

  const onWalletSubmit = async (data: BookingFormValues) => {
    setStatus({ type: "info", message: "Charging wallet..." });
    try {
      const response = await walletMutation.mutateAsync({
        ticket_ids: data.ticket_ids,
        user_email: data.user_email,
        promo_code: data.promo_code || "",
        early_id: data.early_id || 0,
        bundle_id: data.bundle_id || 0,
      });
      setStatus({
        type: "success",
        message: `Wallet payment ${response.status}. Booking ID: ${response.booking_id}.`,
      });
    } catch (error) {
      setStatus({ type: "error", message: getErrorMessage(error, "Wallet payment failed.") });
    }
  };

  const isSubmitting = mutation.isPending || walletMutation.isPending;

  return (
    <div className="container page">
      <div className={styles.layout}>
        <Card className={styles.summary}>
          <h2>Booking summary</h2>
          <p>Event ID: {booking.eventId}</p>
          <p>Tickets: {booking.ticketIds.join(", ")}</p>
          <p>Promo code: {booking.promoCode || "None"}</p>
          <p>Early ID: {booking.earlyId || "None"}</p>
          <p>Bundle ID: {booking.bundleId || "None"}</p>
        </Card>

        <Card>
          <h2>Confirm booking</h2>
          <form className={styles.form} onSubmit={handleSubmit(onSubmit)}>
            <Input
              label="Email"
              type="email"
              error={errors.user_email?.message}
              {...register("user_email")}
            />
            {errors.promo_code && <p className={styles.error}>{errors.promo_code.message}</p>}
            <div className={styles.actions}>
              <Button type="submit" disabled={isSubmitting}>
                {mutation.isPending ? "Creating booking..." : "Proceed to payment"}
              </Button>
              <Button
                type="button"
                variant="outline"
                disabled={isSubmitting}
                onClick={handleSubmit(onWalletSubmit)}
              >
                {walletMutation.isPending ? "Charging wallet..." : "Pay with wallet"}
              </Button>
            </div>
            {status && (
              <p
                className={`status ${
                  status.type === "error" ? "statusError" : status.type === "success" ? "statusSuccess" : "statusInfo"
                }`}
              >
                {status.message}
              </p>
            )}
          </form>
        </Card>
      </div>
    </div>
  );
};
