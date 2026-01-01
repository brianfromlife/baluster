import { useEffect, useState } from "react";
import styles from "./Loading.module.css";

export function Loading() {
  const [showWarmingUp, setShowWarmingUp] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setShowWarmingUp(true);
    }, 750);

    return () => clearTimeout(timer);
  }, []);

  return (
    <div className={styles.loadingContainer}>
      <div className={styles.spinner}>
        <div className={styles.spinnerDot}></div>
        <div className={styles.spinnerDot}></div>
        <div className={styles.spinnerDot}></div>
      </div>
      <div className={styles.message}>
        {showWarmingUp ? "Warming up the servers..." : "Loading..."}
      </div>
    </div>
  );
}
