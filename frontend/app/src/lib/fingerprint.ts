export const DEVICE_ID_KEY = "lingshu_device_id";
export const DEVICE_SECRET_KEY = "lingshu_device_secret_key";

export function getDeviceID() {
  if (typeof window === "undefined") return "";
  const existing = window.localStorage.getItem(DEVICE_ID_KEY);
  if (existing) return existing;
  const id = `dev_${randomID()}_${browserHint()}`;
  window.localStorage.setItem(DEVICE_ID_KEY, id);
  return id;
}

export function setDeviceSecret(secret?: string) {
  if (typeof window === "undefined") return;
  const value = secret?.trim() ?? "";
  if (value) {
    window.localStorage.setItem(DEVICE_SECRET_KEY, value);
  } else {
    window.localStorage.removeItem(DEVICE_SECRET_KEY);
  }
}

function randomID() {
  const webCrypto = globalThis.crypto;
  if (webCrypto?.randomUUID) {
    return webCrypto.randomUUID();
  }
  const bytes = new Uint8Array(16);
  if (webCrypto?.getRandomValues) {
    webCrypto.getRandomValues(bytes);
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function browserHint() {
  if (typeof navigator === "undefined" || typeof screen === "undefined") return "browser";
  const raw = [navigator.language, screen.width, screen.height, screen.colorDepth].filter(Boolean).join("-");
  return raw.replace(/[^a-zA-Z0-9_-]/g, "").slice(0, 24);
}
