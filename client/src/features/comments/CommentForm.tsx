import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { addCommentSchema, AddCommentFormValues } from "./schema";
import { useAddComment } from "./hooks";
import { Input } from "@/shared/ui/Input/Input";
import { Textarea } from "@/shared/ui/Textarea/Textarea";
import { Button } from "@/shared/ui/Button/Button";
import { getErrorMessage } from "@/shared/utils/error";
import styles from "./CommentForm.module.scss";

type Props = {
  eventId: number;
};

export const CommentForm = ({ eventId }: Props) => {
  const { mutateAsync, isPending } = useAddComment(eventId);
  const [status, setStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddCommentFormValues>({
    resolver: zodResolver(addCommentSchema),
    defaultValues: {
      event_id: eventId,
      rating: 5,
      text: "",
      nick: "",
    },
  });

  const onSubmit = async (data: AddCommentFormValues) => {
    setStatus({ type: "info", message: "Sending comment..." });
    try {
      const rating = Math.min(50, Math.max(0, Math.round(data.rating * 10)));
      await mutateAsync({ ...data, rating });
      reset({ event_id: eventId, rating: 5, text: "", nick: "" });
      setStatus({ type: "success", message: "Comment sent successfully." });
    } catch (error) {
      setStatus({ type: "error", message: getErrorMessage(error, "Failed to send comment.") });
    }
  };

  return (
    <form className={styles.form} onSubmit={handleSubmit(onSubmit)}>
      <div className={styles.row}>
        <Input
          label="Nickname"
          placeholder="Your display name"
          error={errors.nick?.message}
          {...register("nick")}
        />
        <Input
          label="Rating (0..5)"
          type="number"
          min={0}
          max={5}
          step={0.1}
          error={errors.rating?.message}
          {...register("rating", { valueAsNumber: true })}
        />
      </div>
      <Textarea
        label="Comment"
        placeholder="Share your experience"
        error={errors.text?.message}
        {...register("text")}
      />
      <Button type="submit" disabled={isPending}>
        {isPending ? "Submitting..." : "Send comment"}
      </Button>
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
  );
};
