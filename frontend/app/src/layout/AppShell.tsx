import React from "react";
import { Activity, Bell, BookOpen, Check, Command as CommandIcon, FileText, Gauge, Globe, KeyRound, LogOut, Moon, PanelTop, RadioTower, ScrollText, Settings, ShieldAlert, Sun, Ticket, Users, WalletCards, Waypoints } from "lucide-react";
import { NavLink, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

import { Button, Command, PopoverContent, PopoverRoot, PopoverTrigger, cn, useHotkeys } from "@lingshu/ui";
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

type NavGroupConfig = {
  titleKey: string;
  fallbackTitle: string;
  items: NavItemConfig[];
};

const userNavGroups: NavGroupConfig[] = [
  {
    titleKey: "overview",
    fallbackTitle: "Overview",
    items: [
      { to: "/dashboard", labelKey: "dashboard", fallbackLabel: "Dashboard", icon: Gauge, hint: "g+d" },
      { to: "/usage", labelKey: "usage", fallbackLabel: "Usage", icon: Activity, hint: "g+u" }
    ]
  },
  {
    titleKey: "resources",
    fallbackTitle: "Resources",
    items: [
      { to: "/api-keys", labelKey: "apiKeys", fallbackLabel: "API Keys", icon: KeyRound },
      { to: "/models", labelKey: "models", fallbackLabel: "Models", icon: PanelTop },
      { to: "/docs", labelKey: "docs", fallbackLabel: "Docs", icon: BookOpen }
    ]
  },
  {
    titleKey: "account",
    fallbackTitle: "Account",
    items: [
      { to: "/redeem", labelKey: "redeem", fallbackLabel: "Redeem", icon: Ticket },
      { to: "/announcements", labelKey: "announcements", fallbackLabel: "Announcements", icon: Bell },
      { to: "/settings", labelKey: "settings", fallbackLabel: "Settings", icon: Settings }
    ]
  }
];

const adminNavGroups: NavGroupConfig[] = [
  {
    titleKey: "operations",
    fallbackTitle: "Operations",
    items: [
      { to: "/admin/dashboard", labelKey: "dashboard", fallbackLabel: "Admin Dashboard", icon: Gauge },
      { to: "/admin/ops", labelKey: "ops", fallbackLabel: "Ops", icon: Activity },
      { to: "/admin/reports", labelKey: "reports", fallbackLabel: "Reports", icon: FileText }
    ]
  },
  {
    titleKey: "access",
    fallbackTitle: "Users & Access",
    items: [
      { to: "/admin/users", labelKey: "users", fallbackLabel: "Users", icon: Users },
      { to: "/admin/api-keys", labelKey: "apiKeys", fallbackLabel: "API Keys", icon: KeyRound },
      { to: "/admin/blacklist", labelKey: "blacklist", fallbackLabel: "Access Blacklist", icon: ShieldAlert }
    ]
  },
  {
    titleKey: "service",
    fallbackTitle: "Service Config",
    items: [
      { to: "/admin/models", labelKey: "models", fallbackLabel: "Models", icon: Waypoints },
      { to: "/admin/channels", labelKey: "channels", fallbackLabel: "Channels", icon: RadioTower },
      { to: "/admin/redeem", labelKey: "redeem", fallbackLabel: "Redeem Codes", icon: Ticket },
      { to: "/admin/announcements", labelKey: "announcements", fallbackLabel: "Announcements", icon: Bell }
    ]
  },
  {
    titleKey: "system",
    fallbackTitle: "System",
    items: [
      { to: "/admin/audit", labelKey: "audit", fallbackLabel: "Audit Logs", icon: ScrollText },
      { to: "/admin/settings", labelKey: "settings", fallbackLabel: "Settings", icon: Settings }
    ]
  }
];

function flattenGroups(groups: NavGroupConfig[]): NavItemConfig[] {
  return groups.flatMap((group) => group.items);
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const { user, isAdmin, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const { t } = useTranslation(["navigation", "common"]);
  const { siteInfo } = useSiteInfo();
  const navigate = useNavigate();
  const [commandOpen, setCommandOpen] = React.useState(false);

  const navItems = React.useMemo(() => {
    const source = isAdmin ? [...flattenGroups(userNavGroups), ...flattenGroups(adminNavGroups)] : flattenGroups(userNavGroups);
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
            <LanguageSwitcher />
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
    <nav className="grid gap-5" aria-label={t("appTitle")}>
      {userNavGroups.map((group) => (
        <NavGroup key={group.titleKey} group={group} section="userSection" />
      ))}
      {isAdmin ? (
        <div className="grid gap-5 border-t border-border pt-4">
          <p className="px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--clay)]">{t("adminSectionTitle")}</p>
          {adminNavGroups.map((group) => (
            <NavGroup key={group.titleKey} group={group} section="adminSection" />
          ))}
        </div>
      ) : null}
    </nav>
  );
}

function NavGroup({ group, section }: { group: NavGroupConfig; section: "userSection" | "adminSection" }) {
  const { t } = useTranslation("navigation");
  return (
    <div className="grid gap-1">
      <p className="px-3 pb-0.5 text-[10px] font-medium uppercase tracking-[0.14em] text-muted-foreground">
        {t(`groups.${group.titleKey}`, group.fallbackTitle)}
      </p>
      {group.items.map((item) => (
        <NavItem key={item.to} item={{ ...item, label: t(`${section}.${item.labelKey}`, item.fallbackLabel) }} />
      ))}
    </div>
  );
}

function LanguageSwitcher() {
  const { t, i18n } = useTranslation("common");
  const [open, setOpen] = React.useState(false);
  return (
    <PopoverRoot open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t("language")}>
          <Globe className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" sideOffset={8} className="z-50 w-44 overflow-hidden rounded-md border border-border bg-card p-1 text-foreground shadow-[var(--shadow-md)] outline-none">
        {availableLocales.map((locale) => {
          const active = i18n.resolvedLanguage === locale.code;
          return (
            <button
              key={locale.code}
              type="button"
              className={cn(
                "flex w-full items-center gap-2 rounded-sm px-2.5 py-2 text-left text-sm transition-colors hover:bg-[var(--bg-subtle)]",
                active ? "text-[var(--clay-hover)]" : "text-foreground"
              )}
              onClick={() => {
                void i18n.changeLanguage(locale.code);
                setOpen(false);
              }}
            >
              <span aria-hidden>{locale.flag}</span>
              <span className="flex-1">{locale.name}</span>
              {active ? <Check className="h-4 w-4" /> : null}
            </button>
          );
        })}
      </PopoverContent>
    </PopoverRoot>
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
