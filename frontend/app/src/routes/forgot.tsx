import React from "react";
import { Link } from "react-router-dom";
import { ArrowLeft, MailCheck } from "lucide-react";
import { useTranslation } from "react-i18next";
import { createAPI } from "@lingshu/shared";
import { Alert, Button, Card, CardContent, Field, Input, toast } from "@lingshu/ui";

import { CaptchaWidget } from "@/components/captcha-widget";
import { PublicFooter } from "@/components/public-footer";
import { SiteNav } from "@/components/site-nav";
import { useSiteInfo } from "@/providers/site-info";

export function ForgotPage() {
  const { t } = useTranslation("auth");
  const [email, setEmail] = React.useState("");
  const [code, setCode] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");
  const [cooldown, setCooldown] = React.useState(0);
  const [sending, setSending] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [error, setError] = React.useState("");
  const [done, setDone] = React.useState(false);
  const { siteInfo } = useSiteInfo();
  const handleCaptchaToken = React.useCallback((token: string) => setCaptchaToken(token), []);

  React.useEffect(() => {
    if (cooldown <= 0) return;
    const timer = window.setTimeout(() => setCooldown((value) => value - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [cooldown]);

  async function sendCode() {
    setError("");
    if (!email.trim()) {
      setError(t("emailRequired"));
      return;
    }
    setSending(true);
    try {
      await createAPI().forgotPassword(email.trim(), captchaToken.trim() || undefined);
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
    setSubmitting(true);
    try {
      await createAPI().resetPassword({ email: email.trim(), code: code.trim(), new_password: password });
      setDone(true);
      toast.success(t("resetSuccess"));
    } catch (err) {
      const text = err instanceof Error ? err.message : t("resetFailed");
      setError(text);
      toast.error(text);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen bg-background">
      <SiteNav />
      <section className="mx-auto grid min-h-[calc(100vh-10rem)] max-w-lg place-items-center px-4 py-10 sm:px-6">
        <Card className="w-full">
          <CardContent className="grid gap-5 p-6">
            <Button asChild variant="link" className="w-fit">
              <Link to="/login"><ArrowLeft className="h-4 w-4" />{t("backToLogin")}</Link>
            </Button>
            <div>
              <h1 className="font-serif text-3xl font-semibold text-foreground">{t("forgotTitle")}</h1>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{t("forgotDescription")}</p>
            </div>
            {done ? (
              <Alert variant="success" title={t("resetDoneTitle")}>
                <Link className="text-[var(--clay)] hover:underline" to="/login">{t("resetDoneDescription")}</Link>
              </Alert>
            ) : (
              <form className="grid gap-4" onSubmit={submit}>
                {error ? <Alert variant="danger">{error}</Alert> : null}
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
                <Field label={t("newPassword")}>
                  <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="new-password" minLength={8} required />
                </Field>
                <CaptchaWidget siteInfo={siteInfo} onToken={handleCaptchaToken} />
                <Button type="submit" disabled={submitting}>{submitting ? t("submitting") : t("resetSubmit")}</Button>
              </form>
            )}
          </CardContent>
        </Card>
      </section>
      <PublicFooter compact />
    </main>
  );
}
