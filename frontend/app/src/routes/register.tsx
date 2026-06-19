import React from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { ArrowLeft, MailCheck } from "lucide-react";
import { useTranslation } from "react-i18next";
import { createAPI } from "@lingshu/shared";
import { Alert, Button, Card, CardContent, Field, Input, toast } from "@lingshu/ui";

import { LegalConsent } from "@/components/legal-consent";
import { PublicFooter } from "@/components/public-footer";
import { SiteNav } from "@/components/site-nav";
import { displaySiteName, useSiteInfo } from "@/providers/site-info";
import { useAuth } from "@/providers/auth";

export function RegisterPage() {
  const { t } = useTranslation("auth");
  const { token, user } = useAuth();
  const { siteInfo, loading } = useSiteInfo();
  const navigate = useNavigate();
  const [username, setUsername] = React.useState("");
  const [email, setEmail] = React.useState("");
  const [code, setCode] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");
  const [agreed, setAgreed] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [sending, setSending] = React.useState(false);
  const [cooldown, setCooldown] = React.useState(0);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    if (cooldown <= 0) return;
    const timer = window.setTimeout(() => setCooldown((value) => value - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [cooldown]);

  if (token && user) {
    return <Navigate to={user.role === "admin" ? "/admin/dashboard" : "/dashboard"} replace />;
  }

  const registrationOpen = siteInfo?.registration_enabled === true;

  async function sendCode() {
    setError("");
    if (!email.trim()) {
      setError(t("emailRequired"));
      return;
    }
    setSending(true);
    try {
      await createAPI().sendEmailCode({ purpose: "register", email: email.trim(), captcha_token: captchaToken.trim() || undefined });
      setCooldown(60);
      toast.success(t("codeSent"));
    } catch (err) {
      const text = err instanceof Error ? err.message : t("codeSendFailed");
      setError(text);
      toast.error(text);
    } finally {
      setSending(false);
    }
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    if (!agreed) {
      setError(t("mustAgree"));
      return;
    }
    setSubmitting(true);
    try {
      await createAPI().register({ username: username.trim(), email: email.trim(), password, code: code.trim(), captcha_token: captchaToken.trim() || undefined });
      toast.success(t("registerSuccess"));
      navigate("/login", { replace: true });
    } catch (err) {
      const text = err instanceof Error ? err.message : t("registerFailed");
      setError(text);
      toast.error(text);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen bg-background">
      <SiteNav />
      <section className="mx-auto grid min-h-[calc(100vh-10rem)] max-w-5xl gap-8 px-4 py-10 sm:px-6 lg:grid-cols-[1fr_420px] lg:items-center">
        <div className="space-y-6">
          <Button asChild variant="link">
            <Link to="/login"><ArrowLeft className="h-4 w-4" />{t("backToLogin")}</Link>
          </Button>
          <div className="space-y-4">
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--clay)]">{displaySiteName(siteInfo)}</p>
            <h1 className="font-serif text-5xl font-semibold leading-tight text-foreground">{t("registerTitle")}</h1>
            <p className="max-w-xl text-base leading-7 text-muted-foreground">{t("registerDescription")}</p>
          </div>
          <div className="rounded-lg border border-border bg-card p-5 text-sm leading-6 text-muted-foreground">
            {t("registerOpsNote")}
          </div>
        </div>
        <Card>
          <CardContent className="p-6">
            {loading ? (
              <p className="text-sm text-muted-foreground">{t("loadingPolicy")}</p>
            ) : !registrationOpen ? (
              <Alert variant="warning" title={t("registrationClosedTitle")}>{t("registrationClosedDescription")}</Alert>
            ) : (
              <form className="grid gap-4" onSubmit={submit}>
                <div>
                  <h2 className="font-serif text-2xl font-semibold">{t("createAccount")}</h2>
                  <p className="mt-2 text-sm text-muted-foreground">{t("createAccountDescription")}</p>
                </div>
                {error ? <Alert variant="danger">{error}</Alert> : null}
                <Field label={t("usernameOnly")}>
                  <Input value={username} onChange={(event) => setUsername(event.target.value)} autoComplete="username" required />
                </Field>
                <Field label={t("email")}>
                  <div className="flex gap-2">
                    <Input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required />
                    <Button type="button" variant="secondary" disabled={sending || cooldown > 0} onClick={sendCode}>
                      <MailCheck className="h-4 w-4" />
                      {cooldown > 0 ? `${cooldown}s` : t("sendCode")}
                    </Button>
                  </div>
                </Field>
                <Field label={t("emailCode")}>
                  <Input value={code} onChange={(event) => setCode(event.target.value)} inputMode="numeric" required />
                </Field>
                <Field label={t("password")}>
                  <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="new-password" minLength={8} required />
                </Field>
                {siteInfo?.captcha_enabled ? (
                  <Field label={t("captchaToken")}>
                    <Input value={captchaToken} onChange={(event) => setCaptchaToken(event.target.value)} autoComplete="off" required />
                  </Field>
                ) : null}
                <LegalConsent checked={agreed} onCheckedChange={setAgreed} />
                <Button type="submit" disabled={submitting || !agreed}>{submitting ? t("submitting") : t("registerSubmit")}</Button>
              </form>
            )}
          </CardContent>
        </Card>
      </section>
      <PublicFooter compact />
    </main>
  );
}
