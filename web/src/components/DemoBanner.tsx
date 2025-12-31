import styles from "./DemoBanner.module.css";

export function DemoBanner() {
  return (
    <div className={styles.demoBanner}>
      <span className={styles.demoBannerText}>
        This is a demo application (not production-ready).
      </span>
    </div>
  );
}

