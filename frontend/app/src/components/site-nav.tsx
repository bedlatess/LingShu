import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button } from "@lingshu/ui";
import { displaySiteName, useSiteInfo } from "@/providers/site-info";

export function SiteNav() {
  const { t, i18n } = useTranslation(["common", "pricing"]);
  const { siteInfo } = useSiteInfo();
  const siteName = displaySiteName(siteInfo);

  return (
    <header className="border-b border-border bg-card/90">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6">
        <Link to="/" className="inline-flex items-center gap-2 font-serif text-lg font-semibold text-foreground">
          {siteInfo?.site_logo_url ? <img src={siteInfo.site_logo_url} alt={siteName} className="h-7 w-7 rounded-md object-contain" /> : null}
          {siteName}
        </Link>
        <div className="flex items-center gap-2">
          <Button type="button" variant="ghost" size="sm" onClick={() => void i18n.changeLanguage(i18n.resolvedLanguage === "zh" ? "en" : "zh")}>
            {i18n.resolvedLanguage === "zh" ? "EN" : "中文"}
          </Button>
          <Button asChild>
            <Link to="/login">{t("pricing:enterConsole")}</Link>
          </Button>
        </div>
      </div>
    </header>
  );
}
