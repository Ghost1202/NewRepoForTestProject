import { forwardRef, SelectHTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Select.module.scss";

export type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  label?: string;
  error?: string;
  hint?: string;
};

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, error, hint, className, children, ...props }, ref) => {
    return (
      <label className={styles.field}>
        {label && <span className={styles.label}>{label}</span>}
        <select
          ref={ref}
          className={clsx(styles.select, error && styles.invalid, className)}
          {...props}
        >
          {children}
        </select>
        {hint && !error && <span className={styles.hint}>{hint}</span>}
        {error && <span className={styles.error}>{error}</span>}
      </label>
    );
  },
);

Select.displayName = "Select";
