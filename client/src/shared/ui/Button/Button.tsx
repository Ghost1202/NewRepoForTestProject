import { ButtonHTMLAttributes } from "react";
import clsx from "clsx";
import styles from "./Button.module.scss";

type Variant = "primary" | "ghost" | "outline" | "danger";
type Size = "sm" | "md" | "lg";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  size?: Size;
};

export const Button = ({ variant = "primary", size = "md", className, ...props }: ButtonProps) => {
  return (
    <button
      type={props.type ?? "button"}
      className={clsx(styles.button, styles[variant], styles[size], className)}
      {...props}
    />
  );
};
