import React from "react";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
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
  const reduceMotion = useReducedMotion();
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

  const content = reduceMotion ? (
    <Outlet />
  ) : (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.2, ease: "easeOut" }}
      >
        <Outlet />
      </motion.div>
    </AnimatePresence>
  );

  if (!shell) return content;
  return (
    <AppShell>
      {content}
    </AppShell>
  );
}

export function HomeRoute() {
  const { token, user, authStatus } = useAuth();
  if (token && authStatus === "checking") return <PageFallback />;
  if (!token) return <Navigate to="/pricing" replace />;
  return <Navigate to={user?.role === "admin" ? "/admin/dashboard" : "/dashboard"} replace />;
}
