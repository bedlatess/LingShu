import React from "react";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { Skeleton } from "@lingshu/ui";

import { useAuth } from "../providers/auth";
import type { RouteMeta } from "./meta";
import { AppShell } from "../layout/AppShell";

function PageFallback() {
  return (
    <div className="min-h-screen bg-background">
      <div className="h-16 border-b border-border bg-card/95" />
      <div className="mx-auto grid max-w-[1600px] gap-5 px-4 py-5 sm:px-6 2xl:px-10 lg:grid-cols-[232px_1fr]">
        <div className="hidden gap-2 lg:grid">
          {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-9 w-full" />)}
        </div>
        <div className="grid gap-4">
          <Skeleton className="h-9 w-1/3" />
          <div className="grid gap-4 md:grid-cols-3">
            {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-24 w-full" />)}
          </div>
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    </div>
  );
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
