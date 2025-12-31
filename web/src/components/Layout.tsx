import { type ReactNode } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import styles from "./Layout.module.css";

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth();

  return (
    <div className={styles.appLayout}>
      <header className={styles.appHeader}>
        <div className={styles.headerContent}>
          <div className={styles.headerLeft}>
            <Link to="/dashboard" className={styles.appTitleLink}>
              <h1 className={styles.appTitle}>Baluster</h1>
            </Link>
          </div>
          <div className={styles.headerRight}>
            {user && (
              <>
                <div className={styles.userInfo}>
                  {user?.avatarUrl && (
                    <img src={user.avatarUrl} alt={user.username} className={styles.userAvatar} />
                  )}
                  <span className={styles.userName}>{user?.username}</span>
                </div>
                <button onClick={logout} className={styles.logoutButton}>
                  Logout
                </button>
              </>
            )}
          </div>
        </div>
      </header>

      <main className={styles.appMain}>{children}</main>
    </div>
  );
}
