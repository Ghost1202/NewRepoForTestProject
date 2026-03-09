import { forwardRef, InputHTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Input.module.scss";

export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
  hint?: string;
};

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, className, ...props }, ref) => {
    return (
      <label className={styles.field}>
        {label && <span className={styles.label}>{label}</span>}
        <input
          ref={ref}
          className={clsx(styles.input, error && styles.invalid, className)}
          {...props}
        />
        {hint && !error && <span className={styles.hint}>{hint}</span>}
        {error && <span className={styles.error}>{error}</span>}
      </label>
    );
  },
);

Input.displayName = "Input";
