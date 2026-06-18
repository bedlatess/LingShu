import React from "react";
import { createAPI } from "@lingshu/shared";
import type { User } from "@lingshu/shared/user-types";

type AuthContextValue = {
  token: string;
  user: User | null;
  authStatus: "checking" | "authenticated" | "anonymous";
  api: ReturnType<typeof createAPI>;
  login: (login: string, password: string) => Promise<void>;
  logout: () => void;
  refreshMe: () => Promise<void>;
};

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = React.useState(() => localStorage.getItem("lingshu_user_token") ?? "");
  const [user, setUser] = React.useState<User | null>(null);
  const [authStatus, setAuthStatus] = React.useState<"checking" | "authenticated" | "anonymous">(() => (localStorage.getItem("lingshu_user_token") ? "checking" : "anonymous"));
  const api = React.useMemo(() => createAPI(token), [token]);

  async function login(loginName: string, password: string) {
    try {
      const result = await createAPI().login(loginName, password);
      if (result.user.role !== "user") {
        throw new Error("请使用普通用户账号登录用户端");
      }
      localStorage.setItem("lingshu_user_token", result.token);
      setToken(result.token);
      setUser(result.user);
      setAuthStatus("authenticated");
    } catch (err) {
      throw err;
    }
  }

  function logout() {
    localStorage.removeItem("lingshu_user_token");
    setToken("");
    setUser(null);
    setAuthStatus("anonymous");
  }

  async function refreshMe() {
    if (!token) {
      setAuthStatus("anonymous");
      return;
    }
    const me = await api.me();
    if (me.role !== "user") {
      logout();
      window.location.replace("/login");
      throw new Error("当前账号不是普通用户");
    }
    setUser(me);
    setAuthStatus("authenticated");
  }

  React.useEffect(() => {
    const onUnauthorized = () => {
      logout();
      window.location.replace("/login");
    };
    window.addEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
    if (!token) {
      setAuthStatus("anonymous");
      return () => window.removeEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
    }
    setAuthStatus((prev) => (prev === "authenticated" ? prev : "checking"));
    refreshMe().catch((err) => {
      const message = err instanceof Error ? err.message : "";
      if (message.includes("登录已过期") || message.includes("没有权限")) {
        logout();
      } else {
        setAuthStatus(user ? "authenticated" : "checking");
      }
    });
    return () => window.removeEventListener("lingshu:unauthorized", onUnauthorized as EventListener);
  }, [token]);

  return <AuthContext.Provider value={{ token, user, authStatus, api, login, logout, refreshMe }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
