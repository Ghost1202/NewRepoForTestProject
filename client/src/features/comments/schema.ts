import { z } from "zod";

export const addCommentSchema = z.object({
  event_id: z.number().positive(),
  rating: z.number().min(0).max(5),
  text: z.string().min(3),
  nick: z.string().min(1),
});

export type AddCommentFormValues = z.infer<typeof addCommentSchema>;
