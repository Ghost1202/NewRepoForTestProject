import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../providers";
import { env } from "@/shared/config";
import styles from "./RootLayout.module.scss";

export const RootLayout = () => {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <div className={styles.shell}>
      <header className={styles.header}>
        <div className={`${styles.headerInner} container`}>
          <div className={styles.brand}>
            <div className={styles.logo}>{env.appName}</div>
            <span className={styles.tag}>Live events and ticketing</span>
          </div>
          <nav className={styles.nav}>
            <NavLink to="/" end className={({ isActive }) => (isActive ? styles.active : undefined)}>
              Events
            </NavLink>
            <NavLink to="/profile" className={({ isActive }) => (isActive ? styles.active : undefined)}>
              Profile
            </NavLink>
            <NavLink to="/admin" className={({ isActive }) => (isActive ? styles.active : undefined)}>
              Admin
            </NavLink>
          </nav>
          <div className={styles.actions}>
            {isAuthenticated ? (
              <>
                <span className={styles.user}>
                  {user?.email || user?.login || "Signed in"}
                </span>
                <button type="button" className={styles.ctaGhost} onClick={logout}>
                  Sign out
                </button>
              </>
            ) : (
              <NavLink to="/auth" className={styles.cta}>
                Sign in
              </NavLink>
            )}
          </div>
        </div>
      </header>
      <main>
        <Outlet />
      </main>
      <footer className={styles.footer}>
        <div className="container">
          <span>Ticket Master UI. Built for your 4 service stack.</span>
        </div>
      </footer>
    </div>
  );
};
