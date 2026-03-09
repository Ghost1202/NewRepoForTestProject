import { HTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Card.module.scss";

export type CardProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "default" | "muted";
};

export const Card = ({ variant = "default", className, ...props }: CardProps) => {
  return <div className={clsx(styles.card, styles[variant], className)} {...props} />;
};
