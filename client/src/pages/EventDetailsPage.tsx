import { useLocation, useNavigate, useParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { EventResponse } from "@/entities/event";
import { useEventDiscounts, useEventTickets, useWaitlist } from "@/features/booking/hooks";
import { waitlistSchema, WaitlistFormValues } from "@/features/booking/schema";
import { useComments, useRating } from "@/features/comments/hooks";
import { CommentForm } from "@/features/comments/CommentForm";
import { useAuth } from "@/app/providers";
import { Badge } from "@/shared/ui/Badge/Badge";
import { Card } from "@/shared/ui/Card/Card";
import { Button } from "@/shared/ui/Button/Button";
import { Input } from "@/shared/ui/Input/Input";
import { formatDateTime, formatMoney, toRating } from "@/shared/utils/format";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./EventDetailsPage.module.scss";

type LocationState = {
  event?: EventResponse;
};

type DiscountType = "promo" | "early" | "bundle";

export const EventDetailsPage = () => {
  const { eventId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, user } = useAuth();
  const event = (location.state as LocationState | null)?.event ?? null;
  const numericEventId = eventId ? Number(eventId) : null;

  const ticketsQuery = useEventTickets(numericEventId);
  const discountsQuery = useEventDiscounts(numericEventId);
  const commentsQuery = useComments(numericEventId, 10, 0);
  const ratingQuery = useRating(numericEventId);
  const waitlistMutation = useWaitlist();

  const [selectedTickets, setSelectedTickets] = useState<number[]>([]);
  const [discountType, setDiscountType] = useState<DiscountType>("promo");
  const [promoCode, setPromoCode] = useState("");
  const [selectedEarly, setSelectedEarly] = useState<number | null>(null);
  const [selectedBundle, setSelectedBundle] = useState<number | null>(null);
  const [discountError, setDiscountError] = useState("");
  const [waitlistStatus, setWaitlistStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);

  const waitlistForm = useForm<WaitlistFormValues>({
    resolver: zodResolver(waitlistSchema),
    defaultValues: {
      event_id: numericEventId ?? 0,
      email: user?.email ?? "",
    },
  });

  useEffect(() => {
    if (numericEventId) {
      waitlistForm.setValue("event_id", numericEventId, { shouldValidate: false });
    }
  }, [numericEventId, waitlistForm]);

  useEffect(() => {
    if (user?.email) {
      waitlistForm.setValue("email", user.email, { shouldValidate: false });
    }
  }, [user?.email, waitlistForm]);

  const toggleTicket = (ticketId: number) => {
    setSelectedTickets((prev) =>
      prev.includes(ticketId) ? prev.filter((id) => id !== ticketId) : [...prev, ticketId],
    );
  };

  const handleCheckout = () => {
    if (!isAuthenticated) {
      navigate("/auth");
      return;
    }

    const hasPromo = discountType === "promo" && promoCode.trim().length > 0;
    const hasEarly = discountType === "early" && Boolean(selectedEarly);
    const hasBundle = discountType === "bundle" && Boolean(selectedBundle);
    setDiscountError("");
    navigate("/checkout", {
      state: {
        eventId: numericEventId,
        ticketIds: selectedTickets,
        promoCode: hasPromo ? promoCode.trim() : "",
        earlyId: hasEarly ? selectedEarly : 0,
        bundleId: hasBundle ? selectedBundle : 0,
      },
    });
  };

  return (
    <div className="container page">
      <section className={styles.hero}>
        <div>
          <Badge tone="muted">Event details</Badge>
          <h1>{event?.name ?? `Event #${eventId}`}</h1>
          <p className={styles.meta}>Start: {formatDateTime(event?.date_start)}</p>
          <p className={styles.meta}>Venue: {event?.venue_id ?? "Unknown"}</p>
          {event?.is_sold_out && <Badge tone="warning">Sold out</Badge>}
        </div>
        <Card className={styles.summaryCard}>
          <h3>Rating</h3>
          <p className={styles.ratingValue}>
            {ratingQuery.data ? `${toRating(ratingQuery.data.avg)} / 5` : "No ratings yet"}
          </p>
          <p className={styles.meta}>Total reviews: {ratingQuery.data?.count ?? 0}</p>
        </Card>
      </section>

      <section className="section">
        <div className={styles.sectionHeader}>
          <h2>Tickets</h2>
          <span className="pill">Select seats</span>
        </div>
        {ticketsQuery.isLoading && <p>Loading tickets...</p>}
        {!ticketsQuery.isLoading && ticketsQuery.data?.tickets.length === 0 && <p>No tickets found.</p>}
        <div className={`${styles.ticketGrid} grid grid-3`}>
          {ticketsQuery.data?.tickets.map((ticket) => (
            <Card key={ticket.id} className={styles.ticketCard}>
              <div className={styles.ticketHeader}>
                <h3>{ticket.sector_name}</h3>
                <Badge tone="muted">Row {ticket.row_no}</Badge>
              </div>
              <p className={styles.meta}>Seat {ticket.seat_no}</p>
              <p className={styles.price}>{formatMoney(ticket.price)}</p>
              <label className={styles.ticketAction}>
                <input
                  type="checkbox"
                  checked={selectedTickets.includes(ticket.id)}
                  onChange={() => toggleTicket(ticket.id)}
                />
                Select ticket
              </label>
            </Card>
          ))}
        </div>
      </section>

      <section className="section">
        <div className={styles.sectionHeader}>
          <h2>Waitlist</h2>
          <span className="pill">Get notified</span>
        </div>
        <Card className={styles.waitlistCard}>
          {isAuthenticated ? (
            <form
              className={styles.waitlistForm}
              onSubmit={waitlistForm.handleSubmit(async (data) => {
                if (!numericEventId) {
                  setWaitlistStatus({ type: "error", message: "Event ID is missing." });
                  return;
                }
                setWaitlistStatus({ type: "info", message: "Joining waitlist..." });
                try {
                  await waitlistMutation.mutateAsync({ event_id: numericEventId, email: data.email });
                  waitlistForm.reset({ event_id: numericEventId, email: "" });
                  setWaitlistStatus({ type: "success", message: "Added to waitlist." });
                } catch (error) {
                  setWaitlistStatus({ type: "error", message: getErrorMessage(error, "Waitlist failed.") });
                }
              })}
            >
              <input type="hidden" {...waitlistForm.register("event_id", { valueAsNumber: true })} />
              <Input
                label="Email"
                type="email"
                error={waitlistForm.formState.errors.email?.message}
                {...waitlistForm.register("email")}
              />
              <Button type="submit" disabled={waitlistMutation.isPending || !numericEventId}>
                {waitlistMutation.isPending ? "Submitting..." : "Join waitlist"}
              </Button>
              {waitlistStatus && (
                <p
                  className={`status ${
                    waitlistStatus.type === "error"
                      ? "statusError"
                      : waitlistStatus.type === "success"
                        ? "statusSuccess"
                        : "statusInfo"
                  }`}
                >
                  {waitlistStatus.message}
                </p>
              )}
            </form>
          ) : (
            <div className={styles.waitlistGuest}>
              <p>Sign in to join the waitlist for this event.</p>
              <Button onClick={() => navigate("/auth")}>Sign in</Button>
            </div>
          )}
        </Card>
      </section>

      <section className="section">
        <div className={styles.sectionHeader}>
          <h2>Discounts</h2>
          <span className="pill">Optional</span>
        </div>
        <Card className={styles.discountCard}>
          <div className={styles.discountTabs}>
            {(["promo", "early", "bundle"] as DiscountType[]).map((type) => (
              <button
                key={type}
                type="button"
                className={type === discountType ? styles.activeTab : styles.tab}
                onClick={() => {
                  setDiscountType(type);
                  setDiscountError("");
                }}
              >
                {type.toUpperCase()}
              </button>
            ))}
          </div>

          {discountType === "promo" && (
            <Input
              label="Promo code"
              placeholder="SUMMER2026"
              value={promoCode}
              onChange={(event) => setPromoCode(event.target.value)}
            />
          )}

          {discountType === "early" && (
            <div className={styles.discountList}>
              {discountsQuery.data?.early?.map((early) => (
                <button
                  key={early.id}
                  type="button"
                  className={selectedEarly === early.id ? styles.discountItemActive : styles.discountItem}
                  onClick={() => setSelectedEarly(early.id)}
                >
                  <span>{early.code}</span>
                  <span>{early.value}%</span>
                  <span>Valid until {formatDateTime(early.valid_until)}</span>
                </button>
              ))}
              {!discountsQuery.data?.early?.length && <p className={styles.meta}>No early promos.</p>}
            </div>
          )}

          {discountType === "bundle" && (
            <div className={styles.discountList}>
              {discountsQuery.data?.bundles?.map((bundle) => (
                <button
                  key={bundle.id}
                  type="button"
                  className={selectedBundle === bundle.id ? styles.discountItemActive : styles.discountItem}
                  onClick={() => setSelectedBundle(bundle.id)}
                >
                  <span>{bundle.code}</span>
                  <span>
                    Buy {bundle.bundle_buy_count} get {bundle.bundle_get_count}
                  </span>
                </button>
              ))}
              {!discountsQuery.data?.bundles?.length && <p className={styles.meta}>No bundle promos.</p>}
            </div>
          )}

          {discountError && <p className={styles.error}>{discountError}</p>}
          <div className={styles.checkoutRow}>
            <Button disabled={selectedTickets.length === 0} onClick={handleCheckout}>
              Continue to checkout ({selectedTickets.length})
            </Button>
          </div>
        </Card>
      </section>

      <section className="section">
        <div className={styles.sectionHeader}>
          <h2>Comments</h2>
          <span className="pill">{commentsQuery.data?.items.length ?? 0} reviews</span>
        </div>
        <div className={styles.commentsGrid}>
          <div className={styles.commentList}>
            {commentsQuery.isLoading && <p>Loading comments...</p>}
            {commentsQuery.data?.items.map((comment, index) => (
              <Card key={`${comment.user_id}-${index}`} className={styles.commentCard}>
                <div className={styles.commentHeader}>
                  <strong>{comment.nick}</strong>
                  <span>{toRating(comment.rating)} / 5</span>
                </div>
                <p>{comment.text}</p>
              </Card>
            ))}
            {!commentsQuery.isLoading && commentsQuery.data?.items.length === 0 && (
              <p>No comments yet.</p>
            )}
          </div>
          <div>
            {isAuthenticated ? (
              <Card>
                <h3>Add a comment</h3>
                {numericEventId ? (
                  <CommentForm eventId={numericEventId} />
                ) : (
                  <p>Event ID is missing.</p>
                )}
              </Card>
            ) : (
              <Card>
                <h3>Sign in to comment</h3>
                <p>You need to be authenticated to leave a review.</p>
              </Card>
            )}
          </div>
        </div>
      </section>
    </div>
  );
};
