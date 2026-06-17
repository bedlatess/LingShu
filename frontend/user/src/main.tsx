import React from "react";
import ReactDOM from "react-dom/client";
import { Navigate, Outlet, RouterProvider, createBrowserRouter, useLocation } from "react-router-dom";

import { AuthProvider, useAuth } from "@/providers/auth";
import { AppLayout } from "@/components/app-layout";
import { LoginPage } from "@/routes/login";
import { DashboardPage } from "@/routes/dashboard";
import { ApiKeysPage } from "@/routes/api-keys";
import { UsagePage } from "@/routes/usage";
import { ModelsPage } from "@/routes/models";
import { RedeemPage } from "@/routes/redeem";
import { AnnouncementsPage } from "@/routes/announcements";
import { SettingsPage } from "@/routes/settings";
import "./styles.css";

function ProtectedRoute() {
  const { token } = useAuth();
  const location = useLocation();
  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
}

const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: <ProtectedRoute />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: "dashboard", element: <DashboardPage /> },
      { path: "api-keys", element: <ApiKeysPage /> },
      { path: "usage", element: <UsagePage /> },
      { path: "models", element: <ModelsPage /> },
      { path: "redeem", element: <RedeemPage /> },
      { path: "announcements", element: <AnnouncementsPage /> },
      { path: "settings", element: <SettingsPage /> }
    ]
  },
  { path: "*", element: <Navigate to="/dashboard" replace /> }
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  </React.StrictMode>
);
