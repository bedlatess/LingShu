import React from "react";
import { Activity, Bell, BookOpen, Command as CommandIcon, FileText, Gauge, KeyRound, LogOut, Moon, PanelTop, RadioTower, ScrollText, Settings, ShieldAlert, Sun, Ticket, Users, WalletCards, Waypoints } from "lucide-react";
import { NavLink, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

import { Button, Command, cn, useHotkeys } from "@lingshu/ui";
import { PublicFooter } from "@/components/public-footer";
import { availableLocales } from "../i18n";
import { useAuth } from "../providers/auth";
import { displaySiteName, useSiteInfo } from "../providers/site-info";
import { useTheme } from "../providers/theme";

type NavItemConfig = {
  to: string;
  labelKey: string;
  fallbackLabel: string;
  icon: React.ComponentType<{ className?: string }>;
  hint?: string;
};

const userNavItems: NavItemConfig[] = [
  { to: "/dashboard", labelKey: "dashboard", fallbackLabel: "Dashboard", icon: Gauge, hint: "g+d" },
  { to: "/docs", labelKey: "docs", fallbackLabel: "Docs", icon: BookOpen },
  { to: "/api-keys", labelKey: "apiKeys", fallbackLabel: "API Keys", icon: KeyRound },
  { to: "/usage", labelKey: "usage", fallbackLabel: "Usage", icon: Activity, hint: "g+u" },
  { to: "/models", labelKey: "models", fallbackLabel: "Models", icon: PanelTop },
  { to: "/redeem", labelKey: "redeem", fallbackLabel: "Redeem", icon: Ticket },
  { to: "/announcements", labelKey: "announcements", fallbackLabel: "Announcements", icon: Bell },
  { to: "/settings", labelKey: "settings", fallbackLabel: "Settings", icon: Settings }
];

const adminNavItems: NavItemConfig[] = [
  { to: "/admin/dashboard", labelKey: "dashboard", fallbackLabel: "Admin Dashboard", icon: Gauge },
  { to: "/admin/ops", labelKey: "ops", fallbackLabel: "Ops", icon: Activity },
  { to: "/admin/users", labelKey: "users", fallbackLabel: "Users", icon: Users },
  { to: "/admin/api-keys", labelKey: "apiKeys", fallbackLabel: "API Keys", icon: KeyRound },
  { to: "/admin/models", labelKey: "models", fallbackLabel: "Models", icon: Waypoints },
  { to: "/admin/channels", labelKey: "channels", fallbackLabel: "Channels", icon: RadioTower },
  { to: "/admin/announcements", labelKey: "announcements", fallbackLabel: "Announcements", icon: Bell },
  { to: "/admin/redeem", labelKey: "redeem", fallbackLabel: "Redeem Codes", icon: Ticket },
  { to: "/admin/reports", labelKey: "reports", fallbackLabel: "Reports", icon: FileText },
  { to: "/admin/audit", labelKey: "audit", fallbackLabel: "Audit Logs", icon: ScrollText },
  { to: "/admin/blacklist", labelKey: "blacklist", fallbackLabel: "Access Blacklist", icon: ShieldAlert },
  { to: "/admin/settings", labelKey: "settings", fallbackLabel: "Settings", icon: Settings }
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const { user, isAdmin, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const { t, i18n } = useTranslation(["navigation", "common"]);
  const { siteInfo } = useSiteInfo();
  const navigate = useNavigate();
  const [commandOpen, setCommandOpen] = React.useState(false);

  const navItems = React.useMemo(() => {
    const source = isAdmin ? [...userNavItems, ...adminNavItems] : userNavItems;
    return source.map((item) => ({
      ...item,
      label: t(item.to.startsWith("/admin") ? `navigation:adminSection.${item.labelKey}` : `navigation:userSection.${item.labelKey}`, item.fallbackLabel)
    }));
  }, [isAdmin, t]);

  useHotkeys({
    "mod+k": () => setCommandOpen(true),
    "g+d": () => navigate(isAdmin ? "/admin/dashboard" : "/dashboard"),
    "g+u": () => navigate("/usage"),
    "/": () => setCommandOpen(true),
    escape: () => setCommandOpen(false)
  });

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-30 border-b border-border bg-card/95 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="grid h-9 w-9 place-items-center overflow-hidden rounded-md border border-border bg-foreground text-sm font-black text-background">
              {siteInfo?.site_logo_url ? <img src={siteInfo.site_logo_url} alt={displaySiteName(siteInfo)} className="h-full w-full object-contain" /> : "LS"}
            </div>
            <div>
              <div className="flex items-center gap-2 font-serif text-sm font-semibold text-foreground">{displaySiteName(siteInfo)}</div>
              <p className="text-xs text-muted-foreground">{isAdmin ? t("navigation:adminTagline") : t("navigation:userTagline")}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 sm:gap-3">
            <Button variant="secondary" size="sm" onClick={() => setCommandOpen(true)} className="hidden sm:inline-flex">
              <CommandIcon className="h-4 w-4" />
              {t("navigation:quickCommand")}
            </Button>
            <div className="hidden items-center gap-1 rounded-md border border-border bg-card p-1 sm:flex" aria-label={t("common:language")}>
              {availableLocales.map((locale) => (
                <Button
                  key={locale.code}
                  type="button"
                  variant={i18n.resolvedLanguage === locale.code ? "secondary" : "ghost"}
                  size="sm"
                  className="h-8 px-2.5"
                  onClick={() => void i18n.changeLanguage(locale.code)}
                >
                  <span aria-hidden>{locale.flag}</span>
                  <span className="text-xs">{locale.code.toUpperCase()}</span>
                </Button>
              ))}
            </div>
            <Button variant="ghost" size="icon" onClick={toggleTheme} aria-label={theme === "dark" ? t("common:lightMode") : t("common:darkMode")}>
              {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            </Button>
            <div className="hidden items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm text-muted-foreground sm:flex">
              <WalletCards className="h-4 w-4 text-primary" />
              {user?.username ?? t("common:loading")}
            </div>
            <Button variant="ghost" size="icon" onClick={logout} aria-label={t("navigation:logout")}>
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[220px_1fr]">
        <aside className="hidden rounded-lg border border-border bg-card p-2 shadow-[var(--shadow-xs)] lg:block">
          <ShellMenu isAdmin={isAdmin} />
        </aside>

        <nav className="flex gap-2 overflow-x-auto lg:hidden" aria-label={t("navigation:appTitle")}>
          {navItems.map((item) => (
            <NavItem key={item.to} item={item} compact />
          ))}
        </nav>

        <main className="min-w-0">{children}</main>
      </div>

      <Command
        open={commandOpen}
        onClose={() => setCommandOpen(false)}
        title={t("navigation:quickCommand")}
        placeholder={t("navigation:commandPlaceholder")}
        emptyText={t("navigation:commandEmpty")}
        items={navItems.map((item) => ({
          label: item.label,
          hint: item.to,
          onSelect: () => navigate(item.to)
        }))}
      />
      <PublicFooter compact />
    </div>
  );
}

function ShellMenu({ isAdmin }: { isAdmin: boolean }) {
  const { t } = useTranslation("navigation");
  return (
    <nav className="grid gap-4" aria-label={t("appTitle")}>
      <div className="grid gap-1">
        {userNavItems.map((item) => (
          <NavItem key={item.to} item={{ ...item, label: t(`userSection.${item.labelKey}`, item.fallbackLabel) }} />
        ))}
      </div>
      {isAdmin ? (
        <div className="grid gap-1 border-t border-border pt-4">
          <p className="px-3 pb-1 text-xs font-medium text-muted-foreground">{t("adminSectionTitle")}</p>
          {adminNavItems.map((item) => (
            <NavItem key={item.to} item={{ ...item, label: t(`adminSection.${item.labelKey}`, item.fallbackLabel) }} />
          ))}
        </div>
      ) : null}
    </nav>
  );
}

function NavItem({ item, compact = false }: { item: NavItemConfig & { label: string }; compact?: boolean }) {
  return (
    <NavLink
      to={item.to}
      className={({ isActive }) =>
        cn(
          compact ? "flex shrink-0 items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm text-muted-foreground" : "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-[var(--bg-subtle)] hover:text-foreground",
          isActive && (compact ? "border-[var(--clay-soft)] bg-[var(--clay-soft)] text-[var(--clay-hover)]" : "bg-[var(--clay-soft)] text-[var(--clay-hover)]")
        )
      }
    >
      <item.icon className="h-4 w-4" />
      <span className={compact ? undefined : "flex-1"}>{item.label}</span>
      {!compact && item.hint ? <span className="font-mono text-[10px] text-muted-foreground">{item.hint}</span> : null}
    </NavLink>
  );
}
