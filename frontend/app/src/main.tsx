import React from "react";
import ReactDOM from "react-dom/client";
import { Link, Navigate, RouterProvider, createBrowserRouter, useLocation, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

import { Alert, Button, Card, CardContent, Input, Toaster, toast } from "@lingshu/ui";
import { AuthProvider, useAuth } from "./providers/auth";
import { ThemeProvider } from "./providers/theme";
import { SiteInfoProvider, displaySiteName, useSiteInfo } from "./providers/site-info";
import { GuardedRoute } from "./router/guards";
import { ensureNamespaces, i18n, setDocumentLanguage, setDocumentTitle } from "./i18n";
import { ErrorBoundary } from "@/components/error-boundary";
import { LegalConsent } from "@/components/legal-consent";
import { PublicFooter } from "@/components/public-footer";

import { PricingPage } from "@/routes/pricing";
import { LandingPage } from "@/routes/landing";
import { RegisterPage } from "@/routes/register";
import { ForgotPage } from "@/routes/forgot";
import { LegalPage } from "@/routes/legal";
import { DashboardPage } from "@/routes/dashboard";
import { ApiKeysPage as UserApiKeysPage } from "@/routes/api-keys";
import { UsagePage } from "@/routes/usage";
import { ModelsPage as UserModelsPage } from "@/routes/models";
import { RedeemPage as UserRedeemPage } from "@/routes/redeem";
import { AnnouncementsPage as UserAnnouncementsPage } from "@/routes/announcements";
import { SettingsPage as UserSettingsPage } from "@/routes/settings";

import { AdminDashboardPage } from "@/pages/admin/admin-dashboard";
import { UsersPage, UserDetailPage } from "@/pages/admin/users";
import { ApiKeysPage as AdminApiKeysPage } from "@/pages/admin/api-keys";
import { ModelsPage as AdminModelsPage, ModelDetailPage } from "@/pages/admin/models";
import { ChannelsPage, ChannelDetailPage } from "@/pages/admin/channels";
import { AnnouncementsPage as AdminAnnouncementsPage } from "@/pages/admin/announcements";
import { RedeemPage as AdminRedeemPage } from "@/pages/admin/redeem";
import { ReportsPage } from "@/pages/admin/reports";
import { OpsPage } from "@/pages/admin/ops";
import { SettingsPage as AdminSettingsPage } from "@/pages/admin/settings";
import { AuditPage } from "@/pages/admin/audit";

import "@fontsource/inter/400.css";
import "@fontsource/inter/500.css";
import "@fontsource/inter/600.css";
import "@fontsource/jetbrains-mono/400.css";
import "./styles.css";

void ensureNamespaces(["common", "navigation", "auth", "dashboard", "keys", "usage", "pricing", "models", "redeem", "announcements", "settings", "admin"]);

function LoginPage() {
  const { login, token, user, authStatus } = useAuth();
  const { siteInfo } = useSiteInfo();
  const { t } = useTranslation(["auth"]);
  const location = useLocation();
  const navigate = useNavigate();
  const [loginName, setLoginName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");
  const [agreed, setAgreed] = React.useState(false);
  const [captchaToken, setCaptchaToken] = React.useState("");

  if (token && authStatus === "checking") {
    return <main className="min-h-screen bg-background" />;
  }

  if (token && user) {
    return <Navigate to={user.role === "admin" ? "/admin/dashboard" : "/dashboard"} replace />;
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (!agreed) {
      setError(t("auth:mustAgree"));
      return;
    }
    setLoading(true);
    setError("");
    try {
      const loggedIn = await login(loginName, password, captchaToken.trim() || undefined);
      const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname;
      const fallback = loggedIn.role === "admin" ? "/admin/dashboard" : "/dashboard";
      navigate(from && (loggedIn.role === "admin" || !from.startsWith("/admin")) ? from : fallback, { replace: true });
      toast.success(t("auth:success"));
    } catch (err) {
      const text = err instanceof Error ? err.message : t("auth:failed");
      setError(text);
      toast.error(text);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-background">
      <div className="mx-auto grid min-h-[calc(100vh-5rem)] w-full max-w-5xl gap-8 px-4 py-10 lg:grid-cols-[1fr_420px] lg:items-center">
        <section className="space-y-6">
          <div className="inline-flex rounded-md border border-border bg-card px-3 py-1 text-xs text-muted-foreground">{t("auth:badge")}</div>
          <div className="space-y-4">
            <h1 className="max-w-2xl font-serif text-5xl font-semibold leading-tight text-foreground">{t("auth:title")}</h1>
            <p className="max-w-xl text-base leading-7 text-muted-foreground">{t("auth:description", { site: displaySiteName(siteInfo) })}</p>
          </div>
          <div className="grid max-w-xl gap-3 text-sm text-muted-foreground sm:grid-cols-3">
            <div className="rounded-lg border border-border bg-card p-4">{t("auth:featureUnified")}</div>
            <div className="rounded-lg border border-border bg-card p-4">{t("auth:featureBilling")}</div>
            <div className="rounded-lg border border-border bg-card p-4">{t("auth:featureRoles")}</div>
          </div>
        </section>
        <Card>
          <CardContent className="p-6">
            <form className="grid gap-4" onSubmit={submit}>
              <div>
                <h2 className="font-serif text-2xl font-semibold">{t("auth:loginTitle")}</h2>
                <p className="mt-2 text-sm text-muted-foreground">{t("auth:loginDescription")}</p>
              </div>
              {error ? <Alert variant="danger">{error}</Alert> : null}
              <label className="grid gap-2 text-sm">
                {t("auth:username")}
                <Input value={loginName} onChange={(event) => setLoginName(event.target.value)} autoComplete="username" required />
              </label>
              <label className="grid gap-2 text-sm">
                {t("auth:password")}
                <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" required />
              </label>
              {siteInfo?.captcha_enabled ? (
                <label className="grid gap-2 text-sm">
                  {t("auth:captchaToken")}
                  <Input value={captchaToken} onChange={(event) => setCaptchaToken(event.target.value)} autoComplete="off" required />
                </label>
              ) : null}
              <div className="flex items-center justify-between text-sm">
                <Button asChild variant="link">
                  <Link to="/forgot">{t("auth:forgotLink")}</Link>
                </Button>
                {siteInfo?.registration_enabled ? (
                  <Button asChild variant="link">
                    <Link to="/register">{t("auth:registerLink")}</Link>
                  </Button>
                ) : null}
              </div>
              <LegalConsent checked={agreed} onCheckedChange={setAgreed} />
              <Button type="submit" disabled={loading || !agreed}>{loading ? t("auth:submitting") : t("auth:submit")}</Button>
            </form>
          </CardContent>
        </Card>
      </div>
      <PublicFooter compact />
    </main>
  );
}

function AdminDashboardRoute() {
  const { api, user } = useAuth();
  return <AdminDashboardPage api={api} me={user!} />;
}

function AdminRoute({ page }: { page: (api: ReturnType<typeof useAuth>["api"]) => React.ReactNode }) {
  const { api } = useAuth();
  return <>{page(api)}</>;
}

function I18nEffects() {
  const { t } = useTranslation(["navigation"]);
  const { siteInfo } = useSiteInfo();

  React.useEffect(() => {
    const sync = () => {
      setDocumentLanguage();
      setDocumentTitle(displaySiteName(siteInfo) || t("navigation:appTitle"));
    };
    sync();
    i18n.on("languageChanged", sync);
    return () => {
      i18n.off("languageChanged", sync);
    };
  }, [t, siteInfo]);

  return null;
}

const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  { path: "/forgot", element: <ForgotPage /> },
  { path: "/legal/:slug", element: <LegalPage /> },
  { path: "/pricing", element: <PricingPage /> },
  { path: "/", element: <LandingPage /> },
  {
    path: "/",
    element: <GuardedRoute />,
    children: [
      { path: "dashboard", element: <DashboardPage /> },
      { path: "api-keys", element: <UserApiKeysPage /> },
      { path: "usage", element: <UsagePage /> },
      { path: "models", element: <UserModelsPage /> },
      { path: "redeem", element: <UserRedeemPage /> },
      { path: "announcements", element: <UserAnnouncementsPage /> },
      { path: "settings", element: <UserSettingsPage /> }
    ]
  },
  {
    path: "/admin",
    element: <GuardedRoute meta={{ requiresAdmin: true }} />,
    children: [
      { index: true, element: <Navigate to="/admin/dashboard" replace /> },
      { path: "dashboard", element: <AdminDashboardRoute /> },
      { path: "users", element: <AdminRoute page={(api) => <UsersPage api={api} />} /> },
      { path: "users/:id", element: <AdminRoute page={(api) => <UserDetailPage api={api} />} /> },
      { path: "api-keys", element: <AdminRoute page={(api) => <AdminApiKeysPage api={api} />} /> },
      { path: "models", element: <AdminRoute page={(api) => <AdminModelsPage api={api} />} /> },
      { path: "models/:id", element: <AdminRoute page={(api) => <ModelDetailPage api={api} />} /> },
      { path: "channels", element: <AdminRoute page={(api) => <ChannelsPage api={api} />} /> },
      { path: "channels/:id", element: <AdminRoute page={(api) => <ChannelDetailPage api={api} />} /> },
      { path: "announcements", element: <AdminRoute page={(api) => <AdminAnnouncementsPage api={api} />} /> },
      { path: "redeem", element: <AdminRoute page={(api) => <AdminRedeemPage api={api} />} /> },
      { path: "reports", element: <AdminRoute page={(api) => <ReportsPage api={api} />} /> },
      { path: "ops", element: <AdminRoute page={(api) => <OpsPage api={api} />} /> },
      { path: "settings", element: <AdminRoute page={(api) => <AdminSettingsPage api={api} />} /> },
      { path: "audit", element: <AdminRoute page={(api) => <AuditPage api={api} />} /> }
    ]
  },
  { path: "*", element: <Navigate to="/" replace /> }
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <AuthProvider>
        <ThemeProvider>
          <SiteInfoProvider>
            <I18nEffects />
            <RouterProvider router={router} />
          </SiteInfoProvider>
          <Toaster />
        </ThemeProvider>
      </AuthProvider>
    </ErrorBoundary>
  </React.StrictMode>
);
