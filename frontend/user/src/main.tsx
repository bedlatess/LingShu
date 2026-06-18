import React, { Suspense, lazy } from "react";
import ReactDOM from "react-dom/client";
import { Navigate, Outlet, RouterProvider, createBrowserRouter, useLocation } from "react-router-dom";

import { AuthProvider, useAuth } from "@/providers/auth";
import { AppLayout } from "@/components/app-layout";
import { ErrorBoundary } from "@/components/error-boundary";
import { Toaster } from "@/components/ui/sonner";
import "./styles.css";

const LoginPage = lazy(() => import("@/routes/login").then((module) => ({ default: module.LoginPage })));
const PricingPage = lazy(() => import("@/routes/pricing").then((module) => ({ default: module.PricingPage })));
const DashboardPage = lazy(() => import("@/routes/dashboard").then((module) => ({ default: module.DashboardPage })));
const ApiKeysPage = lazy(() => import("@/routes/api-keys").then((module) => ({ default: module.ApiKeysPage })));
const UsagePage = lazy(() => import("@/routes/usage").then((module) => ({ default: module.UsagePage })));
const ModelsPage = lazy(() => import("@/routes/models").then((module) => ({ default: module.ModelsPage })));
const RedeemPage = lazy(() => import("@/routes/redeem").then((module) => ({ default: module.RedeemPage })));
const AnnouncementsPage = lazy(() => import("@/routes/announcements").then((module) => ({ default: module.AnnouncementsPage })));
const SettingsPage = lazy(() => import("@/routes/settings").then((module) => ({ default: module.SettingsPage })));

function PageFallback() {
  return <div className="min-h-screen bg-background" />;
}

function lazyPage(element: React.ReactNode) {
  return <Suspense fallback={<PageFallback />}>{element}</Suspense>;
}

function ProtectedRoute() {
  const { token, authStatus } = useAuth();
  const location = useLocation();
  if (token && authStatus === "checking") {
    return <PageFallback />;
  }
  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
}

function HomeRoute() {
  const { token } = useAuth();
  return <Navigate to={token ? "/dashboard" : "/pricing"} replace />;
}

const router = createBrowserRouter([
  { path: "/login", element: lazyPage(<LoginPage />) },
  { path: "/pricing", element: lazyPage(<PricingPage />) },
  { path: "/", element: <HomeRoute /> },
  {
    path: "/",
    element: <ProtectedRoute />,
    children: [
      { path: "dashboard", element: lazyPage(<DashboardPage />) },
      { path: "api-keys", element: lazyPage(<ApiKeysPage />) },
      { path: "usage", element: lazyPage(<UsagePage />) },
      { path: "models", element: lazyPage(<ModelsPage />) },
      { path: "redeem", element: lazyPage(<RedeemPage />) },
      { path: "announcements", element: lazyPage(<AnnouncementsPage />) },
      { path: "settings", element: lazyPage(<SettingsPage />) }
    ]
  },
  { path: "*", element: <Navigate to="/" replace /> }
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <AuthProvider>
        <RouterProvider router={router} />
        <Toaster richColors position="top-right" visibleToasts={1} />
      </AuthProvider>
    </ErrorBoundary>
  </React.StrictMode>
);
