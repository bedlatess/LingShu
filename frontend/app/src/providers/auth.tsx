import React from "react";
import { createAPI } from "@lingshu/shared";
import type { User } from "@lingshu/shared";
import { getDeviceID } from "@/lib/fingerprint";

const TOKEN_KEY = "lingshu_token";
const LEGACY_USER_TOKEN_KEY = "lingshu_user_token";
const LEGACY_ADMIN_TOKEN_KEY = "lingshu_admin_token";

type AuthContextValue = {
  token: string;
  user: User | null;
  authStatus: "checking" | "authenticated" | "anonymous";
  api: ReturnType<typeof createAPI>;
  isAdmin: boolean;
  login: (login: string, password: string, captchaToken?: string) => Promise<User>;
  logout: () => void;
  refreshMe: () => Promise<void>;
};

const AuthContext = React.createContext<AuthContextValue | null>(null);

function readInitialToken() {
  return localStorage.getItem(TOKEN_KEY) ?? localStorage.getItem(LEGACY_ADMIN_TOKEN_KEY) ?? localStorage.getItem(LEGACY_USER_TOKEN_KEY) ?? "";
}

function clearStoredTokens() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(LEGACY_USER_TOKEN_KEY);
  localStorage.removeItem(LEGACY_ADMIN_TOKEN_KEY);
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = React.useState(readInitialToken);
  const [user, setUser] = React.useState<User | null>(null);
  const [authStatus, setAuthStatus] = React.useState<"checking" | "authenticated" | "anonymous">(() => (readInitialToken() ? "checking" : "anonymous"));
  const api = React.useMemo(() => createAPI(token), [token]);

  React.useEffect(() => {
    getDeviceID();
  }, []);

  const logout = React.useCallback(() => {
    clearStoredTokens();
    setToken("");
    setUser(null);
    setAuthStatus("anonymous");
  }, []);

  const refreshMe = React.useCallback(async () => {
    if (!token) {
      setAuthStatus("anonymous");
      return;
    }
    const me = await api.me();
    localStorage.setItem(TOKEN_KEY, token);
    setUser(me);
    setAuthStatus("authenticated");
  }, [api, token]);

  async function login(loginName: string, password: string, captchaToken?: string) {
    const result = await createAPI().login(loginName, password, captchaToken);
    clearStoredTokens();
    localStorage.setItem(TOKEN_KEY, result.token);
    setToken(result.token);
    setUser(result.user);
    setAuthStatus("authenticated");
    return result.user;
  }

  React.useEffect(() => {
    const onUnauthorized = () => {
      logout();
      window.dispatchEvent(new CustomEvent("lingshu:auth-expired"));
    };
    window.addEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
    if (!token) {
      setAuthStatus("anonymous");
      return () => window.removeEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
    }
    setAuthStatus((prev) => (prev === "authenticated" ? prev : "checking"));
    refreshMe().catch(() => logout());
    return () => window.removeEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
  }, [logout, refreshMe, token]);

  return (
    <AuthContext.Provider value={{ token, user, authStatus, api, isAdmin: user?.role === "admin", login, logout, refreshMe }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
