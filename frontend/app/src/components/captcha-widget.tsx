import React from "react";
import { Alert } from "@lingshu/ui";
import { useTranslation } from "react-i18next";

import type { PublicSiteInfo } from "@lingshu/shared";

type CaptchaProvider = "turnstile" | "cloudflare_turnstile" | "hcaptcha";

declare global {
  interface Window {
    turnstile?: {
      render: (container: HTMLElement, options: Record<string, unknown>) => string;
      remove?: (widgetID: string) => void;
      reset?: (widgetID: string) => void;
    };
    hcaptcha?: {
      render: (container: HTMLElement, options: Record<string, unknown>) => string;
      remove?: (widgetID: string) => void;
      reset?: (widgetID: string) => void;
    };
  }
}

export function CaptchaWidget({ siteInfo, onToken }: { siteInfo: PublicSiteInfo | null; onToken: (token: string) => void }) {
  const { t } = useTranslation("auth");
  const containerRef = React.useRef<HTMLDivElement | null>(null);
  const widgetRef = React.useRef<string>("");
  const provider = normalizeProvider(siteInfo?.captcha_provider);
  const siteKey = siteInfo?.captcha_site_key?.trim() ?? "";
  const enabled = Boolean(siteInfo?.captcha_enabled);
  const [status, setStatus] = React.useState<"idle" | "loading" | "ready" | "error">("idle");

  React.useEffect(() => {
    onToken("");
    if (!enabled) {
      setStatus("idle");
      return;
    }
    if (!provider || !siteKey) {
      setStatus("error");
      return;
    }
    let cancelled = false;
    setStatus("loading");
    loadCaptchaScript(provider)
      .then(() => {
        if (cancelled || !containerRef.current) return;
        containerRef.current.innerHTML = "";
        const api = provider === "hcaptcha" ? window.hcaptcha : window.turnstile;
        if (!api) {
          setStatus("error");
          return;
        }
        widgetRef.current = api.render(containerRef.current, {
          sitekey: siteKey,
          callback: (token: string) => onToken(token),
          "expired-callback": () => onToken(""),
          "error-callback": () => {
            onToken("");
            setStatus("error");
          }
        });
        setStatus("ready");
      })
      .catch(() => setStatus("error"));
    return () => {
      cancelled = true;
      const api = provider === "hcaptcha" ? window.hcaptcha : window.turnstile;
      if (widgetRef.current && api?.remove) {
        api.remove(widgetRef.current);
      }
      widgetRef.current = "";
    };
  }, [enabled, onToken, provider, siteKey]);

  if (!enabled) return null;
  return (
    <div className="grid gap-2">
      <div ref={containerRef} className="min-h-[65px] rounded-md border border-border bg-[var(--bg-subtle)] p-2" />
      {status === "loading" ? <p className="text-xs text-muted-foreground">{t("captchaLoading")}</p> : null}
      {status === "error" ? <Alert variant="danger">{t("captchaUnavailable")}</Alert> : null}
    </div>
  );
}

function normalizeProvider(value?: string): CaptchaProvider | "" {
  const normalized = String(value ?? "").trim().toLowerCase();
  if (normalized === "turnstile" || normalized === "cloudflare_turnstile" || normalized === "hcaptcha") {
    return normalized;
  }
  return "";
}

function loadCaptchaScript(provider: CaptchaProvider) {
  const src = provider === "hcaptcha" ? "https://js.hcaptcha.com/1/api.js?render=explicit" : "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
  const id = provider === "hcaptcha" ? "lingshu-hcaptcha" : "lingshu-turnstile";
  if (document.getElementById(id)) return Promise.resolve();
  return new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.id = id;
    script.src = src;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("captcha script failed"));
    document.head.appendChild(script);
  });
}
