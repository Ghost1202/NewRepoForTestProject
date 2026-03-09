import { z } from "zod";

export const bookingSchema = z
  .object({
    ticket_ids: z.array(z.number().int().positive()).min(1, "Select at least one ticket"),
    user_email: z.string().email(),
    promo_code: z.string().optional(),
    early_id: z.number().int().optional(),
    bundle_id: z.number().int().optional(),
  })
  .superRefine((value, ctx) => {
    const hasPromo = Boolean(value.promo_code && value.promo_code.trim().length > 0);
    const hasEarly = Boolean(value.early_id && value.early_id > 0);
    const hasBundle = Boolean(value.bundle_id && value.bundle_id > 0);
    const selected = [hasPromo, hasEarly, hasBundle].filter(Boolean).length;
    if (selected > 1) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["promo_code"],
        message: "Select only one discount type (promo, early, or bundle).",
      });
    }

    const uniqueIds = new Set(value.ticket_ids);
    if (uniqueIds.size !== value.ticket_ids.length) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["ticket_ids"],
        message: "Ticket IDs must be unique.",
      });
    }
  });

export type BookingFormValues = z.infer<typeof bookingSchema>;

export const transferSchema = z.object({
  ticket_id: z.number().int().positive(),
  to_user_id: z.number().int().positive(),
});

export type TransferFormValues = z.infer<typeof transferSchema>;

export const refundSchema = z.object({
  event_id: z.number().int().positive(),
  ticket_id: z.number().int().positive(),
});

export type RefundFormValues = z.infer<typeof refundSchema>;

export const waitlistSchema = z.object({
  event_id: z.number().int().positive(),
  email: z.string().email(),
});

export type WaitlistFormValues = z.infer<typeof waitlistSchema>;
