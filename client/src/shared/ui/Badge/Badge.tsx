import { HTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Badge.module.scss";

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: "accent" | "muted" | "warning";
};

export const Badge = ({ tone = "accent", className, ...props }: BadgeProps) => {
  return <span className={clsx(styles.badge, styles[tone], className)} {...props} />;
};
