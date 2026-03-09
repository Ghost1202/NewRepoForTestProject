import { z } from "zod";

export const ticketConstructorSchema = z.object({
  name: z.string().min(1),
  type: z.string().min(1),
  price: z.number().int().positive(),
  rows_count: z.number().int().positive(),
  seats_per_row: z.number().int().positive(),
});

export const promoSchema = z.object({
  code: z.string().min(1),
  type: z.string().min(1),
  value: z.number().positive(),
  sector: z.string().optional(),
});

export const earlySchema = promoSchema.extend({
  valid_until: z.string().min(1),
});

export const bundleSchema = z.object({
  code: z.string().min(1),
  sector: z.string().optional(),
  buy_count: z.number().int().positive(),
  get_count: z.number().int().positive(),
});

export const createEventSchema = z.object({
  venue_id: z.number().int().positive(),
  name: z.string().min(2),
  date_start: z.string().min(1),
  post_date: z.string().min(1),
  sale_start_date: z.string().min(1),
  max_price_cof: z.number().positive(),
  min_price_cof: z.number().positive(),
  ticket_constructor: z.array(ticketConstructorSchema).optional(),
  promos: z.array(promoSchema).optional(),
  early: z.array(earlySchema).optional(),
  bundles: z.array(bundleSchema).optional(),
});

export type CreateEventFormValues = z.infer<typeof createEventSchema>;

export const updateEventSchema = z.object({
  event_id: z.number().int().positive(),
  name: z.string().min(2),
  date_start: z.string().min(1),
  max_price_cof: z.number().positive(),
  min_price_cof: z.number().positive(),
});

export type UpdateEventFormValues = z.infer<typeof updateEventSchema>;

export const addPromoSchema = z.object({
  event_id: z.number().int().positive(),
  promos: z.array(promoSchema).min(1, "Add at least one promo"),
});

export type AddPromoFormValues = z.infer<typeof addPromoSchema>;

export const addEarlySchema = z.object({
  event_id: z.number().int().positive(),
  early: z.array(earlySchema).min(1, "Add at least one early promo"),
});

export type AddEarlyFormValues = z.infer<typeof addEarlySchema>;

export const addBundleSchema = z.object({
  event_id: z.number().int().positive(),
  bundles: z.array(bundleSchema).min(1, "Add at least one bundle"),
});

export type AddBundleFormValues = z.infer<typeof addBundleSchema>;

export const searchPromoSchema = z.object({
  event_id: z.number().int().positive().optional(),
  limit: z.number().int().positive(),
  offset: z.number().int().min(0),
});

export type SearchPromoValues = z.infer<typeof searchPromoSchema>;
