import { useEffect, useState } from "react";
import styles from "./Loading.module.css";

const messages = [
  "Loading...",
  "Warming up the servers...",
  "Almost there...",
  "Just a moment...",
  "Preparing everything...",
];

export function Loading() {
  const [messageIndex, setMessageIndex] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setMessageIndex((prevIndex) => (prevIndex + 1) % messages.length);
    }, 2300);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className={styles.loadingContainer}>
      <div className={styles.spinner}>
        <div className={styles.spinnerDot}></div>
        <div className={styles.spinnerDot}></div>
        <div className={styles.spinnerDot}></div>
      </div>
      <div key={messageIndex} className={styles.message}>
        {messages[messageIndex]}
      </div>
    </div>
  );
}
