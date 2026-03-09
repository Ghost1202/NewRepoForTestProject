import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { useRefundBooking, useTransferTicket, useUserTickets } from "@/features/booking/hooks";
import { useWallet } from "@/features/wallet/hooks";
import {
  refundSchema,
  RefundFormValues,
  transferSchema,
  TransferFormValues,
} from "@/features/booking/schema";
import { formatMoney } from "@/shared/utils/format";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./ProfilePage.module.scss";

export const ProfilePage = () => {
  const ticketsQuery = useUserTickets();
  const transferMutation = useTransferTicket();
  const refundMutation = useRefundBooking();
  const walletQuery = useWallet();
  const [transferStatus, setTransferStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [refundStatus, setRefundStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);

  const transferForm = useForm<TransferFormValues>({
    resolver: zodResolver(transferSchema),
  });

  const refundForm = useForm<RefundFormValues>({
    resolver: zodResolver(refundSchema),
  });

  return (
    <div className="container page">
      <section className="section">
        <h1>Wallet</h1>
        {walletQuery.isLoading && <p>Loading wallet...</p>}
        {walletQuery.isError && (
          <p className="status statusError">{getErrorMessage(walletQuery.error, "Failed to load wallet.")}</p>
        )}
        {walletQuery.data && (
          <div className={`${styles.walletGrid} grid grid-3`}>
            <Card className={styles.walletCard}>
              <p className={styles.walletLabel}>Available balance</p>
              <p className={styles.walletValue}>{formatMoney(walletQuery.data.balance, "USD")}</p>
            </Card>
            <Card className={styles.walletCard}>
              <p className={styles.walletLabel}>Pending balance</p>
              <p className={styles.walletValue}>{formatMoney(walletQuery.data.pending_balance, "USD")}</p>
            </Card>
            <Card className={styles.walletCard}>
              <p className={styles.walletLabel}>Bonus points</p>
              <p className={styles.walletValue}>{walletQuery.data.bonus_points}</p>
            </Card>
          </div>
        )}
      </section>

      <section className="section">
        <h1>Your tickets</h1>
        {ticketsQuery.isLoading && <p>Loading tickets...</p>}
        <div className={`${styles.ticketGrid} grid grid-3`}>
          {ticketsQuery.data?.tickets.map((ticket) => (
            <Card key={ticket.id} className={styles.ticketCard}>
              <h3>{ticket.sector_name}</h3>
              <p>Event ID: {ticket.event_id}</p>
              <p>Seat: {ticket.row_no}-{ticket.seat_no}</p>
              <p>{formatMoney(ticket.price)}</p>
              <span className="pill">{ticket.status}</span>
            </Card>
          ))}
        </div>
        {!ticketsQuery.isLoading && ticketsQuery.data?.tickets.length === 0 && <p>No tickets yet.</p>}
      </section>

      <section className="section">
        <div className={styles.forms}>
          <Card>
            <h2>Transfer ticket</h2>
            <form
              className={styles.form}
              onSubmit={transferForm.handleSubmit(async (data) => {
                setTransferStatus({ type: "info", message: "Transferring ticket..." });
                try {
                  await transferMutation.mutateAsync(data);
                  setTransferStatus({ type: "success", message: "Ticket transferred." });
                } catch (error) {
                  setTransferStatus({ type: "error", message: getErrorMessage(error, "Transfer failed.") });
                }
              })}
            >
              <Input
                label="Ticket ID"
                type="number"
                error={transferForm.formState.errors.ticket_id?.message}
                {...transferForm.register("ticket_id", { valueAsNumber: true })}
              />
              <Input
                label="To user ID"
                type="number"
                error={transferForm.formState.errors.to_user_id?.message}
                {...transferForm.register("to_user_id", { valueAsNumber: true })}
              />
              <Button type="submit" disabled={transferMutation.isPending}>
                {transferMutation.isPending ? "Transferring..." : "Transfer"}
              </Button>
              {transferStatus && (
                <p
                  className={`status ${
                    transferStatus.type === "error"
                      ? "statusError"
                      : transferStatus.type === "success"
                        ? "statusSuccess"
                        : "statusInfo"
                  }`}
                >
                  {transferStatus.message}
                </p>
              )}
            </form>
          </Card>

          <Card>
            <h2>Request refund</h2>
            <form
              className={styles.form}
              onSubmit={refundForm.handleSubmit(async (data) => {
                setRefundStatus({ type: "info", message: "Submitting refund..." });
                try {
                  await refundMutation.mutateAsync(data);
                  setRefundStatus({ type: "success", message: "Refund submitted." });
                } catch (error) {
                  setRefundStatus({ type: "error", message: getErrorMessage(error, "Refund failed.") });
                }
              })}
            >
              <Input
                label="Event ID"
                type="number"
                error={refundForm.formState.errors.event_id?.message}
                {...refundForm.register("event_id", { valueAsNumber: true })}
              />
              <Input
                label="Ticket ID"
                type="number"
                error={refundForm.formState.errors.ticket_id?.message}
                {...refundForm.register("ticket_id", { valueAsNumber: true })}
              />
              <Button type="submit" disabled={refundMutation.isPending}>
                {refundMutation.isPending ? "Sending..." : "Submit refund"}
              </Button>
              {refundStatus && (
                <p
                  className={`status ${
                    refundStatus.type === "error"
                      ? "statusError"
                      : refundStatus.type === "success"
                        ? "statusSuccess"
                        : "statusInfo"
                  }`}
                >
                  {refundStatus.message}
                </p>
              )}
            </form>
          </Card>

        </div>
      </section>
    </div>
  );
};
