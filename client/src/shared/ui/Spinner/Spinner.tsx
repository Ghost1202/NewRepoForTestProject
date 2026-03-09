import styles from "./Spinner.module.scss";

export const Spinner = ({ label = "Loading..." }: { label?: string }) => {
  return (
    <div className={styles.spinner}>
      <div className={styles.dot} />
      <span>{label}</span>
    </div>
  );
};
