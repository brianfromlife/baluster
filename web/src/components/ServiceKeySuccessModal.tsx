import styles from "./ServiceKeySuccessModal.module.css";

interface ServiceKeySuccessModalProps {
  token: string;
  organizationId?: string;
  onClose: () => void;
}

export function ServiceKeySuccessModal({ token, organizationId, onClose }: ServiceKeySuccessModalProps) {
  const handleCopyToken = () => {
    navigator.clipboard.writeText(token);
    alert("Token copied to clipboard!");
  };

  const handleCopyOrgId = () => {
    if (organizationId) {
      navigator.clipboard.writeText(organizationId);
      alert("Organization ID copied to clipboard!");
    }
  };

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h2 className={styles.modalTitle}>Service Key Created Successfully</h2>
          <button className={styles.modalCloseButton} onClick={onClose} aria-label="Close">
            <svg
              width="20"
              height="20"
              viewBox="0 0 20 20"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M15 5L5 15M5 5L15 15"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </button>
        </div>

        <div className={styles.modalBody}>
          <p className={styles.modalDescription}>
            Save this token now - you won't be able to see it again!
          </p>

          <div className={styles.tokenDisplay}>
            <label className={styles.tokenLabel}>Service Key Token</label>
            <div className={styles.tokenInputGroup}>
              <input type="text" readOnly value={token} className={styles.tokenInput} />
              <button type="button" onClick={handleCopyToken} className={styles.copyButton}>
                Copy
              </button>
            </div>
            <p className={styles.tokenWarning}>⚠️ This token will not be shown again. Make sure to save it securely.</p>
          </div>

          {organizationId && (
            <div className={styles.tokenDisplay} style={{ marginTop: "1.5rem" }}>
              <label className={styles.tokenLabel}>Organization ID (Required for API calls)</label>
              <div className={styles.tokenInputGroup}>
                <input type="text" readOnly value={organizationId} className={styles.tokenInput} />
                <button type="button" onClick={handleCopyOrgId} className={styles.copyButton}>
                  Copy
                </button>
              </div>
              <p className={styles.tokenWarning} style={{ fontSize: "0.875rem", marginTop: "0.5rem" }}>
                ⚠️ Include this in the <code>x-org-id</code> header when calling the <code>/api/v1/access</code> endpoint.
              </p>
            </div>
          )}
        </div>

        <div className={styles.modalFooter}>
          <button type="button" onClick={onClose} className={styles.modalButton}>
            I've Saved the Token
          </button>
        </div>
      </div>
    </div>
  );
}

