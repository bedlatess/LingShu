export const designTokens = {
  colors: {
    bg: "#F0EEE6",
    bgSubtle: "#E8E6DC",
    surface: "#FAF9F5",
    ink: "#141413",
    inkMuted: "#5F5D57",
    inkFaint: "#87867F",
    border: "#D8D4CA",
    borderStrong: "#B0AEA5",
    clay: "#C6613F",
    clayHover: "#A94F31",
    claySoft: "#E3DACC",
    success: "#4F7A4D",
    successSoft: "#E2E9DB",
    warning: "#B5821F",
    warningSoft: "#F1E5C7",
    danger: "#A84031",
    dangerSoft: "#F0DCD5",
    info: "#516E88",
    infoSoft: "#DDE6EB"
  },
  radius: {
    sm: "4px",
    md: "6px",
    lg: "10px",
    card: "6px",
    button: "6px"
  },
  spacing: {
    xs: "4px",
    sm: "8px",
    md: "12px",
    lg: "16px",
    xl: "24px",
    "2xl": "32px",
    "3xl": "48px",
    page: "32px",
    section: "24px"
  },
  typography: {
    display: { size: "44px", lineHeight: "1.06", letterSpacing: "-0.035em", weight: 600 },
    h1: { size: "36px", lineHeight: "1.1", letterSpacing: "-0.02em", weight: 600 },
    h2: { size: "26px", lineHeight: "1.18", letterSpacing: "-0.018em", weight: 600 },
    h3: { size: "20px", lineHeight: "1.25", letterSpacing: "-0.01em", weight: 600 },
    body: { size: "15px", lineHeight: "1.65", letterSpacing: "0", weight: 400 },
    small: { size: "13px", lineHeight: "1.5", letterSpacing: "0", weight: 400 },
    caption: { size: "12px", lineHeight: "1.45", letterSpacing: "0.04em", weight: 500 }
  },
  font: {
    serif: '"Tiempos Text", "Source Han Serif SC", "Songti SC", Georgia, "Times New Roman", serif',
    sans: 'Inter, "PingFang SC", "Microsoft YaHei", ui-sans-serif, system-ui, sans-serif',
    mono: '"JetBrains Mono", "SF Mono", Menlo, Consolas, monospace'
  },
  shadow: {
    xs: "0 1px 2px rgba(20,20,19,0.04)",
    sm: "0 2px 8px rgba(20,20,19,0.05)",
    md: "0 8px 24px rgba(20,20,19,0.08)",
    lg: "0 18px 48px rgba(20,20,19,0.10)",
    xl: "0 28px 70px rgba(20,20,19,0.12)"
  },
  status: {
    success: { bg: "#E2E9DB", border: "rgba(79,122,77,0.30)", text: "#4F7A4D" },
    warning: { bg: "#F1E5C7", border: "rgba(181,130,31,0.30)", text: "#B5821F" },
    danger: { bg: "#F0DCD5", border: "rgba(168,64,49,0.30)", text: "#A84031" },
    info: { bg: "#DDE6EB", border: "rgba(81,110,136,0.30)", text: "#516E88" }
  },
  motion: {
    fast: "120ms",
    base: "160ms",
    slow: "240ms",
    ease: "cubic-bezier(0.32,0.72,0,1)"
  },
  z: {
    dropdown: 40,
    sticky: 30,
    overlay: 50,
    modal: 60,
    toast: 70
  }
} as const;
