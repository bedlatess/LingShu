import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import resourcesToBackend from "i18next-resources-to-backend";
import { initReactI18next } from "react-i18next";

export const LOCALE_STORAGE_KEY = "lingshu_locale";
export const availableLocales = [
  { code: "en", name: "English", flag: "US" },
  { code: "zh", name: "Chinese", flag: "CN" }
] as const;

export type LocaleCode = (typeof availableLocales)[number]["code"];

const resources = resourcesToBackend((language: string, namespace: string) => import(`./locales/${language}/${namespace}.ts`));

if (!i18n.isInitialized) {
  void i18n
    .use(resources)
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      fallbackLng: "en",
      supportedLngs: ["en", "zh"],
      defaultNS: "common",
      ns: ["common", "navigation", "auth", "dashboard", "docs", "keys", "usage", "pricing", "models", "redeem", "announcements", "settings", "admin"],
      interpolation: {
        escapeValue: false
      },
      detection: {
        order: ["localStorage", "navigator"],
        lookupLocalStorage: LOCALE_STORAGE_KEY,
        caches: ["localStorage"],
        convertDetectedLanguage: (lng: string) => (lng?.toLowerCase().startsWith("zh") ? "zh" : "en")
      }
    });
}

export async function ensureNamespaces(namespaces: string[]) {
  await i18n.loadNamespaces(namespaces);
}

export function setDocumentLanguage(language = i18n.resolvedLanguage ?? i18n.language) {
  document.documentElement.lang = language?.startsWith("zh") ? "zh-CN" : "en";
}

export function setDocumentTitle(title: string) {
  document.title = title;
}

export { i18n };
