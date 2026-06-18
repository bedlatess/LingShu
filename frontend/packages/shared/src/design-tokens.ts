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
    warning: "#B5821F",
    danger: "#A84031",
    dangerSoft: "#F0DCD5"
  },
  radius: {
    sm: "4px",
    md: "6px",
    lg: "10px",
    card: "6px",
    button: "6px"
  },
  spacing: {
    page: "32px",
    section: "24px"
  },
  font: {
    serif: '"Tiempos Text", "Source Han Serif SC", "Songti SC", Georgia, "Times New Roman", serif',
    sans: 'Inter, "PingFang SC", "Microsoft YaHei", ui-sans-serif, system-ui, sans-serif',
    mono: '"JetBrains Mono", "SF Mono", Menlo, Consolas, monospace'
  },
  shadow: {
    sm: "0 1px 2px rgba(20,20,19,0.04)",
    md: "0 4px 16px rgba(20,20,19,0.08)"
  }
} as const;

