import React from "react";
import { createAPI, type PublicSiteInfo } from "@lingshu/shared";

type SiteInfoState = {
  siteInfo: PublicSiteInfo | null;
  loading: boolean;
  refreshSiteInfo: () => Promise<void>;
};

const SiteInfoContext = React.createContext<SiteInfoState | null>(null);

export function SiteInfoProvider({ children }: { children: React.ReactNode }) {
  const [siteInfo, setSiteInfo] = React.useState<PublicSiteInfo | null>(null);
  const [loading, setLoading] = React.useState(true);

  const refreshSiteInfo = React.useCallback(async () => {
    setLoading(true);
    try {
      const result = await createAPI().siteInfo();
      setSiteInfo(result);
      applyBrandColor(result.brand_primary_color);
      document.title = result.site_name || "LingShu";
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    void refreshSiteInfo();
  }, [refreshSiteInfo]);

  return (
    <SiteInfoContext.Provider value={{ siteInfo, loading, refreshSiteInfo }}>
      {children}
    </SiteInfoContext.Provider>
  );
}

export function useSiteInfo() {
  const context = React.useContext(SiteInfoContext);
  if (!context) throw new Error("useSiteInfo must be used inside SiteInfoProvider");
  return context;
}

export function displaySiteName(siteInfo: PublicSiteInfo | null) {
  return siteInfo?.site_name?.trim() || "LingShu";
}

function applyBrandColor(color?: string) {
  if (!color || !/^#[0-9a-fA-F]{6}$/.test(color.trim())) return;
  document.documentElement.style.setProperty("--clay", color.trim());
}
