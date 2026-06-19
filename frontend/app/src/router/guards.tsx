import React from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";

import { useAuth } from "../providers/auth";
import type { RouteMeta } from "./meta";
import { AppShell } from "../layout/AppShell";

function PageFallback() {
  return <div className="min-h-screen bg-background" />;
}

export function GuardedRoute({ meta, shell = true }: { meta?: RouteMeta; shell?: boolean }) {
  const { token, user, authStatus } = useAuth();
  const location = useLocation();
  const requiresAuth = meta?.requiresAuth !== false;
  const requiresAdmin = meta?.requiresAdmin === true;

  if (requiresAuth && token && authStatus === "checking") {
    return <PageFallback />;
  }

  if (requiresAuth && !token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (requiresAdmin && user?.role !== "admin") {
    return <Navigate to="/dashboard" replace />;
  }

  if (!shell) return <Outlet />;
  return (
    <AppShell>
      <Outlet />
    </AppShell>
  );
}

export function HomeRoute() {
  const { token, user, authStatus } = useAuth();
  if (token && authStatus === "checking") return <PageFallback />;
  if (!token) return <Navigate to="/pricing" replace />;
  return <Navigate to={user?.role === "admin" ? "/admin/dashboard" : "/dashboard"} replace />;
}
