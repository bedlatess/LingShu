import { Link } from "react-router-dom";
import { Mail } from "lucide-react";
import { useTranslation } from "react-i18next";

import { displaySiteName, useSiteInfo } from "@/providers/site-info";

export function PublicFooter({ compact = false }: { compact?: boolean }) {
  const { t } = useTranslation("auth");
  const { siteInfo } = useSiteInfo();
  const name = displaySiteName(siteInfo);

  return (
    <footer className="border-t border-border bg-card/70">
      <div className={compact ? "mx-auto grid max-w-5xl gap-3 px-4 py-6 text-xs text-muted-foreground sm:px-6" : "mx-auto grid max-w-7xl gap-4 px-4 py-8 text-sm text-muted-foreground sm:px-6 md:grid-cols-[1fr_auto]"}>
        <div>
          <p className="font-serif text-base font-semibold text-foreground">{name}</p>
          <p className="mt-2 max-w-2xl leading-6">{t("footer.description")}</p>
        </div>
        <div className="flex flex-wrap items-center gap-x-4 gap-y-2 md:justify-end">
          <LegalLink href={siteInfo?.tos_url || "/legal/tos"} label={t("terms")} />
          <LegalLink href={siteInfo?.privacy_url || "/legal/privacy"} label={t("privacy")} />
          {siteInfo?.contact_email ? (
            <a className="inline-flex items-center gap-1 hover:text-foreground" href={`mailto:${siteInfo.contact_email}`}>
              <Mail className="h-3.5 w-3.5" />
              {siteInfo.contact_email}
            </a>
          ) : null}
          {siteInfo?.site_icp ? (
            <a className="hover:text-foreground" href="https://beian.miit.gov.cn/" target="_blank" rel="noreferrer">
              {siteInfo.site_icp}
            </a>
          ) : null}
          {siteInfo?.site_police_beian ? <span>{siteInfo.site_police_beian}</span> : null}
        </div>
      </div>
    </footer>
  );
}

function LegalLink({ href, label }: { href: string; label: string }) {
  if (/^https?:\/\//.test(href)) {
    return <a className="hover:text-foreground" href={href} target="_blank" rel="noreferrer">{label}</a>;
  }
  return <Link className="hover:text-foreground" to={href}>{label}</Link>;
}
