import { forwardRef, TextareaHTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Textarea.module.scss";

export type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label?: string;
  error?: string;
  hint?: string;
};

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, hint, className, ...props }, ref) => {
    return (
      <label className={styles.field}>
        {label && <span className={styles.label}>{label}</span>}
        <textarea
          ref={ref}
          className={clsx(styles.textarea, error && styles.invalid, className)}
          {...props}
        />
        {hint && !error && <span className={styles.hint}>{hint}</span>}
        {error && <span className={styles.error}>{error}</span>}
      </label>
    );
  },
);

Textarea.displayName = "Textarea";
