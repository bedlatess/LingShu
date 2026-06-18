# 灵枢 v1.6 UI 重构方案 —— Anthropic / Claude 极简学术风

> 本文档为 **agent 可直接执行** 的前端 UI 重构计划。目标：把用户端（frontend/user）与管理端（frontend/admin）统一到同一套「极简学术风」设计语言，达到生产级高质量，并彻底摒弃当前的科技感装饰。
>
> 执行者请严格按阶段顺序进行，每个阶段结束都要跑构建验证。不改任何后端代码，不改计费语义，不新增臃肿功能。

---

## 0. 设计目标与铁律

### 0.1 视觉目标（Anthropic / Claude Code 风格）

参考 Anthropic 官网、Claude.ai、Claude Code 的视觉语言，核心特征：

1. **暖色纸感底色**：以象牙白 / 米白（warm off-white）为主背景，而非纯白或深色。营造「纸张 / 学术论文」的阅读沉浸感。
2. **高对比墨黑文字**：正文近黑（不是纯黑 #000，用 `#141413` 这类带暖调的深墨色），降低视觉疲劳。
3. **赤陶橙点缀色（Clay / Rust）**：Anthropic 标志性的暖橙棕作为唯一强调色（按钮、链接、激活态、焦点环），克制使用，不铺满。
4. **衬线标题 + 无衬线正文**：标题用衬线字体（如 Tiempos / Georgia / 思源宋体）营造学术感，正文用高可读无衬线（Inter / system-ui）。
5. **极克制的描边与阴影**：用 1px 细边框 + 极淡阴影分隔区块，**禁止**毛玻璃、发光阴影、大模糊、彩色 radial-gradient、网格底纹、hover 抬升位移。
6. **大量留白**：宽松的行高、段距、内边距，信息密度低而清晰。

### 0.2 必须摒弃的现有元素（逐项清除）

- `body` 的青绿 + 紫色双 `radial-gradient` 背景（styles.css 65-68 行）
- `.glass` 毛玻璃类（styles.css 82-87 行）及所有页面里的 `className="glass"`
- `.soft-grid` 网格底纹（styles.css 94-99 行）及 app-layout 的 `soft-grid`
- `--shadow-glow` 青绿发光阴影（styles.css 25 行）及 button/site-nav 里的 `shadow-glow`
- 所有 `hover:-translate-y-0.5` / `hover:-translate-y-1` 抬升动效
- 所有 `border-white/10`、`bg-white/[0.04]`、`bg-white/[0.035]` 硬编码半透明白
- 青绿主色 `174 80% 52%` 和紫色强调 `270 75% 64%`
- 渐变图标背景、装饰方块（stat-card / empty-state 里的绝对定位发光块）

### 0.3 铁律（不可违反）

1. 计费语义不变：`charge = base_cost × rate_multiplier`，不碰后端、不碰钱路径。
2. 用户端绝不展示 `base_cost` / `rate_multiplier` / `gross_profit`（当前已合规，重构中不得引入）。
3. 不新增在线支付、分销、多级权限、2FA 等推后功能。
4. 重构是**视觉层**改造，不改业务逻辑、不改 API 调用、不改路由结构、不改数据流。
5. 每个页面改造后功能必须与改造前完全等价。

---

## 0.4 实施前：从 Anthropic 官网取色校准（必做前置步骤）

在写任何代码之前，先以 Anthropic 官方当前视觉为基准校准下面 1.1 节的颜色令牌。文档给出的色值是高质量近似，但 Anthropic 会迭代官网配色，**以官网实测为准**：

1. 用 WebFetch / 浏览器打开 `https://www.anthropic.com` 和 `https://claude.ai`，观察其主背景、卡片面、正文文字、强调色（按钮/链接）、边框的实际颜色。
2. 重点校准这几个值（Anthropic 官网近期主背景偏暖,常见在 `#F0EEE6`～`#F0EEE6` 一带,强调色为赤陶橙 `#C15F3C`～`#CC785C` 一带）：
   - `bg` 主背景、`surface` 卡片面
   - `ink` 主文字
   - `clay` 强调色（务必取 Anthropic 实际橙棕,不要自行偏红或偏黄）
   - `border` 边框
3. 若官网实测值与 1.1 节默认值有差异，**以官网为准**更新 1.1 节表格、1.4 节 `design-tokens.ts`、第 3 节 styles.css 的 `:root`、第 7 节 AntD 令牌，保证四处同步一致。
4. 字体：观察官网标题的衬线质感与正文无衬线字体，确认 1.2 节字体栈方向正确（标题衬线、正文 Inter 类）。

> 取色后在执行说明里记录最终采用的色值,后续所有阶段以校准后的令牌为唯一基准。如果无法访问官网,就用 1.1 节默认值,但 clay 强调色必须保持 Anthropic 赤陶橙调性,绝不可偏向蓝色或荧光色。

---

## 1. 设计令牌规范（Single Source of Truth）

两端必须共用同一套令牌。先在共享包定义，再分别注入到用户端 CSS 变量和管理端 AntD ConfigProvider。

> **注意**：下表色值为待校准的高质量近似,执行前请按 0.4 节以 Anthropic 官网实测值校准。

### 1.1 颜色（浅色学术主题，默认且唯一）

| 语义 | HEX | 用途 |
|---|---|---|
| `bg` 主背景 | `#F0EEE6` | 页面底色，暖象牙白 |
| `bgSubtle` 次级背景 | `#E8E6DC` | 区块分隔、侧栏、表头 |
| `surface` 卡片面 | `#FAF9F5` | 卡片 / 弹窗背景（比主背景略亮的纸感） |
| `ink` 主文字 | `#141413` | 标题与正文，暖墨黑 |
| `inkMuted` 次级文字 | `#5F5D57` | 描述、辅助说明、占位 |
| `inkFaint` 极弱文字 | `#87867F` | 时间戳、禁用态 |
| `border` 边框 | `#D8D4CA` | 默认 1px 描边 |
| `borderStrong` 强描边 | `#B0AEA5` | 输入框、激活分隔 |
| `clay` 强调色 | `#C6613F` | 主按钮、链接、激活、焦点环（Anthropic 赤陶橙） |
| `clayHover` | `#A94F31` | 强调色 hover |
| `claySoft` | `#E3DACC` | 强调色浅底（激活态背景、badge 底） |
| `success` | `#4F7A4D` | 成功（克制的橄榄绿，非荧光绿） |
| `warning` | `#B5821F` | 警告（暖琥珀） |
| `danger` | `#A84031` | 危险 / 删除（砖红，与 clay 区分但同色系） |
| `dangerSoft` | `#F0DCD5` | 危险浅底 |

> 说明：success/warning/danger 都选低饱和暖色调，避免传统 SaaS 的荧光绿 / 亮黄 / 大红，保持学术克制。

### 1.2 字体

- **标题字体栈** `--font-serif`：`"Tiempos Text", "Source Han Serif SC", "Songti SC", Georgia, "Times New Roman", serif`
  - 用于 PageHeader 主标题、CardTitle、登录页大标题、统计数字。
- **正文字体栈** `--font-sans`：`Inter, "PingFang SC", "Microsoft YaHei", ui-sans-serif, system-ui, sans-serif`
  - 用于所有正文、标签、按钮、表格、表单。
- **等宽字体栈** `--font-mono`：`"JetBrains Mono", "SF Mono", Menlo, Consolas, monospace`
  - 用于 API Key、base_url、request_id、代码片段。
- 通过 `@fontsource` 引入 Inter（必装）；衬线优先用系统已有的思源宋体 / Georgia 兜底，**不强制联网加载 Tiempos**（Anthropic 自有字体无法获取，用 Georgia + 思源宋体高质量替代）。
  - 安装：`npm --workspace @lingshu/user add @fontsource/inter @fontsource/jetbrains-mono`
  - 在 `frontend/user/src/main.tsx` 顶部 `import "@fontsource/inter/400.css"; import "@fontsource/inter/500.css"; import "@fontsource/inter/600.css"; import "@fontsource/jetbrains-mono/400.css";`
  - 管理端同样安装并在 main.tsx 引入。

### 1.3 圆角 / 间距 / 阴影

- 圆角：`--radius-sm: 4px`、`--radius-md: 6px`、`--radius-lg: 10px`。整体偏小，学术克制（不要 16px+ 的大圆角）。
- 阴影（极淡，仅用于弹窗 / 下拉浮层）：
  - `--shadow-sm: 0 1px 2px rgba(20,20,19,0.04)`
  - `--shadow-md: 0 4px 16px rgba(20,20,19,0.08)`
  - 卡片默认**不用阴影**，靠 1px border 分隔。
- 焦点环：`--ring: clay`，统一 `box-shadow: 0 0 0 2px var(--bg), 0 0 0 4px clay` 风格或 Tailwind `ring-2 ring-clay`。

### 1.4 落地：新建共享令牌文件

**文件**：`frontend/packages/shared/src/design-tokens.ts`（重写现有文件）

```ts
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
  radius: { sm: "4px", md: "6px", lg: "10px" },
  spacing: { page: "32px", section: "24px" },
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
```

> 保留导出名 `designTokens`，因为 admin-shell.tsx 已引用。确保 `frontend/packages/shared/src/index.ts` 仍导出它。

---

## 2. 执行阶段总览

| 阶段 | 范围 | 产出 |
|---|---|---|
| **P0** | 设计令牌 + 用户端全局 styles.css 重写 + 字体引入 | 视觉地基切换为浅色学术风 |
| **P1** | 用户端基础组件（button/card/input/badge/tabs/skeleton/sonner） | shadcn 组件适配新令牌 |
| **P2** | 用户端外壳（app-layout / site-nav / page-header）+ 辅助组件（stat-card/empty-state） | 控制台与公共页外壳统一 |
| **P3** | 用户端各业务页面去 glass / 去装饰 / 套用新组件 | 全部用户页面落地 |
| **P4** | 管理端 AntD 主题令牌注入 + 全局 css + 字体 + 外壳布局 | 管理端切换为同源浅色风 |
| **P5** | 管理端各页面细节统一（表格 / 表单 / 弹窗 / 标签色） | 管理端落地 |
| **P6** | 跨端一致性走查 + 响应式 + 可达性 + 构建验证 | 收口 |

---

## 3. P0：视觉地基（用户端 styles.css 重写）

**文件**：`frontend/user/src/styles.css`（整体重写为以下内容）

```css
@import "tailwindcss";
@plugin "@tailwindcss/typography";

@theme {
  --color-background: var(--bg);
  --color-foreground: var(--ink);
  --color-card: var(--surface);
  --color-card-foreground: var(--ink);
  --color-primary: var(--clay);
  --color-primary-foreground: #FAF9F5;
  --color-secondary: var(--bg-subtle);
  --color-secondary-foreground: var(--ink);
  --color-muted: var(--bg-subtle);
  --color-muted-foreground: var(--ink-muted);
  --color-accent: var(--clay-soft);
  --color-accent-foreground: var(--ink);
  --color-destructive: var(--danger);
  --color-destructive-foreground: #FAF9F5;
  --color-border: var(--border-c);
  --color-input: var(--border-strong);
  --color-ring: var(--clay);
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 10px;
  --font-sans: Inter, "PingFang SC", "Microsoft YaHei", ui-sans-serif, system-ui, sans-serif;
  --font-serif: "Tiempos Text", "Source Han Serif SC", "Songti SC", Georgia, "Times New Roman", serif;
  --font-mono: "JetBrains Mono", "SF Mono", Menlo, Consolas, monospace;
}

@layer base {
  :root {
    --bg: #F0EEE6;
    --bg-subtle: #E8E6DC;
    --surface: #FAF9F5;
    --ink: #141413;
    --ink-muted: #5F5D57;
    --ink-faint: #87867F;
    --border-c: #D8D4CA;
    --border-strong: #B0AEA5;
    --clay: #C6613F;
    --clay-hover: #A94F31;
    --clay-soft: #E3DACC;
    --radius: 6px;
  }

  * { border-color: var(--border-c); }

  html { color-scheme: light; }

  body {
    min-height: 100vh;
    background-color: var(--bg);
    color: var(--ink);
    font-family: var(--font-sans);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    font-synthesis: none;
    letter-spacing: 0;
    text-rendering: optimizeLegibility;
    margin: 0;
  }

  h1, h2, h3, .font-serif { font-family: var(--font-serif); }

  button, input { font: inherit; }
}

@layer components {
  .page-grid { display: grid; gap: 1.5rem; }

  /* 学术卡片：纸感面 + 1px 细边，无阴影无毛玻璃 */
  .paper {
    border: 1px solid var(--border-c);
    background: var(--surface);
    border-radius: var(--radius-lg);
  }

  .prose { color: var(--ink-muted); }
  .prose h1, .prose h2, .prose h3, .prose strong { color: var(--ink); font-family: var(--font-serif); }
  .prose a { color: var(--clay); text-decoration: underline; text-underline-offset: 2px; }
  .prose code, .prose pre {
    border-radius: var(--radius-md);
    background: var(--bg-subtle);
    font-family: var(--font-mono);
  }
}
```

**关键变更说明给 agent**：
- 删除 `--shadow-glow`、`.glass`、`.soft-grid`、`body` 的 radial-gradient。
- `color-scheme` 从 dark 改 light。
- 新增 `.paper` 类替代 `.glass`（后续页面把 `glass` 全替换成 `paper`）。
- 标题元素自动应用衬线字体。

**字体引入**：执行 1.2 节的 npm 安装 + main.tsx import。

**验收 P0**：`npm --workspace @lingshu/user run build` 通过；页面背景变为暖白、文字变墨黑（此时组件还没改完，允许局部错位）。

---

## 4. P1：用户端基础组件改造

路径均在 `frontend/user/src/components/ui/`。

### 4.1 button.tsx

把 `buttonVariants` 改为（去 shadow-glow、去 translate）：

```ts
const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-[var(--clay-hover)]",
        secondary: "border border-border bg-secondary text-secondary-foreground hover:bg-[var(--bg-subtle)]",
        ghost: "text-foreground hover:bg-[var(--bg-subtle)]",
        outline: "border border-input bg-transparent text-foreground hover:bg-[var(--bg-subtle)]",
        destructive: "bg-destructive text-destructive-foreground hover:bg-[#8f3328]"
      },
      size: { default: "h-9 px-4", sm: "h-8 px-3 text-[13px]", lg: "h-10 px-5", icon: "h-9 w-9" }
    },
    defaultVariants: { variant: "default", size: "default" }
  }
);
```

要点：`transition-colors`（不再 `transition-all`），移除 `hover:-translate-y-0.5`，移除 `shadow-glow`，高度从 h-10 收到 h-9 更精致。

### 4.2 card.tsx

```tsx
export function Card({ className, ...props }) {
  return <div className={cn("rounded-lg border border-border bg-card text-card-foreground", className)} {...props} />;
}
// CardHeader: "flex flex-col gap-1.5 p-6"
// CardTitle:  "text-lg font-semibold leading-tight font-serif"
// CardDescription: "text-sm text-muted-foreground leading-6"
// CardContent: "p-6 pt-0"
```

要点：去掉 `shadow-sm`（靠 border 分隔）；CardTitle 加 `font-serif`；内边距 p-5 → p-6 增加留白。

### 4.3 input.tsx

```tsx
"flex h-9 w-full rounded-md border border-input bg-surface px-3 py-2 text-sm text-foreground placeholder:text-[var(--ink-faint)] focus-visible:outline-none focus-visible:border-[var(--clay)] focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50"
```

要点：去 `shadow-sm`；焦点用 clay 描边 + 淡环。

### 4.4 badge.tsx

加 variant 支持（学术克制配色）：

```tsx
// 用 cva：
// default: "border-transparent bg-[var(--bg-subtle)] text-foreground"
// clay:    "border-transparent bg-[var(--clay-soft)] text-[var(--clay-hover)]"
// success: "border-transparent bg-[#E6EDE5] text-[#3D6B3B]"
// danger:  "border-transparent bg-[var(--clay-soft)] text-[var(--danger)]"
// outline: "border-border text-muted-foreground"
// 基类: "inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium"
```

> 当前 badge 不支持 variant，改造时保留无 variant 调用的默认行为（default）。

### 4.5 tabs.tsx

换成 Radix（更专业、可达性好）：
- 安装：`npm --workspace @lingshu/user add @radix-ui/react-tabs`
- 用 `@radix-ui/react-tabs` 重写，`TabsList` 用 `inline-flex rounded-md border border-border bg-secondary p-1`，激活态 `data-[state=active]:bg-surface data-[state=active]:text-foreground data-[state=active]:shadow-sm`。
- **注意**：保持现有调用 API（TabsList/TabsTrigger/TabsContent 的 props）兼容，逐个 import 处验证。若改 Radix 影响调用面过大，可保留手撸实现但更新配色（二选一，优先 Radix）。

### 4.6 skeleton.tsx

`"animate-pulse rounded-md bg-[var(--bg-subtle)]"`（颜色适配浅底）。

### 4.7 sonner.tsx

去掉 `next-themes` 依赖（无 Provider），固定 light：
```tsx
<Sonner theme="light" className="toaster group" toastOptions={{
  classNames: {
    toast: "group toast bg-[var(--surface)] text-[var(--ink)] border border-[var(--border-c)] rounded-md",
    description: "text-[var(--ink-muted)]",
    actionButton: "bg-[var(--clay)] text-white",
  }
}} ... />
```
移除 `import { useTheme } from "next-themes"`。

**验收 P1**：build 通过；按钮 / 卡片 / 输入框呈现纸面 + clay 强调，无发光无位移。

---

## 5. P2：用户端外壳与辅助组件

### 5.1 page-header.tsx

```tsx
export function PageHeader({ eyebrow, title, description }) {
  return (
    <div className="mb-8 flex flex-col gap-3 border-b border-border pb-6">
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--clay)]">{eyebrow}</p>
      <h1 className="max-w-3xl font-serif text-3xl font-semibold tracking-tight text-foreground sm:text-[2.5rem] leading-[1.1]">{title}</h1>
      <p className="max-w-2xl text-[15px] leading-7 text-muted-foreground">{description}</p>
    </div>
  );
}
```

要点：eyebrow 仍用 clay 但收敛字间距；标题用衬线；底部加细分隔线强化「章节标题」学术感。

### 5.2 app-layout.tsx（控制台外壳）

整体改造：
- 外层 `min-h-screen soft-grid` → `min-h-screen bg-background`（去网格）。
- Header：`border-b border-white/10 bg-background/70 backdrop-blur-2xl` → `border-b border-border bg-surface`（去毛玻璃，纸面顶栏）。
- Logo 方块：`bg-primary ... shadow-glow` → `bg-[var(--ink)] text-[var(--bg)]`（墨黑底白字，去发光）或 `border border-border bg-surface text-ink`。去掉 `Sparkles` 装饰图标。
- 余额/用户名 pill：`border-white/10 bg-white/[0.04]` → `border-border bg-surface`。
- 侧栏 `aside`：`glass` → `paper`（或直接 `border-r border-border bg-bg-subtle`，做成贴边纸感侧栏）。
- 侧栏 NavLink 激活态：`bg-primary/12 text-primary` → `bg-[var(--clay-soft)] text-[var(--clay-hover)] font-medium`；hover `hover:bg-white/[0.06]` → `hover:bg-[var(--bg-subtle)]`。

### 5.3 site-nav.tsx（公共顶栏）

- `border-white/10 bg-background/75 backdrop-blur-xl` → `border-border bg-surface`。
- Logo `bg-primary ... shadow-glow` → `bg-[var(--ink)] text-[var(--bg)]`。
- 链接 hover 配色沿用 muted→foreground。

### 5.4 stat-card.tsx

- 去掉 4 个 tone 的渐变图标背景 / 发光，统一为：图标用 `text-[var(--clay)]`，图标容器 `border border-border bg-[var(--bg-subtle)]`。
- 外层 Card 去 `glass` / 去 `hover:-translate-y-1 hover:border-primary/35`，改 `paper`，hover 仅 `hover:border-[var(--border-strong)]`。
- 统计大数字用 `font-serif`。

### 5.5 empty-state.tsx

删除 4 个绝对定位的装饰发光方块，改为简洁居中：图标（`text-[var(--ink-faint)]`）+ 标题（serif）+ 描述（muted）。

**验收 P2**：build 通过；登录态控制台和公共页外壳都呈现统一暖白纸感，无任何毛玻璃/网格/发光。

---

## 6. P3：用户端业务页面落地

逐页处理。**通用替换规则**（对 routes/ 下所有页面执行）：

1. 所有 `className="glass"` → `className="paper"`（或并入 Card 后删除）。
2. 所有 `border-white/10`、`bg-white/[0.04]`、`bg-white/[0.035]` → `border-border`、`bg-[var(--bg-subtle)]`。
3. 所有 `shadow-glow` → 删除。
4. 所有 `hover:-translate-y-*` → 删除，必要时换 `hover:border-[var(--border-strong)]`。
5. 所有 `text-primary` 图标点缀保留（现在 primary=clay，语义正确）。
6. 价格 / 金额 / 大数字加 `font-serif` 强化。

逐页要点（路径 `frontend/user/src/routes/`）：

- **login.tsx**：登录卡用 `paper`，大标题用衬线；背景纯 `bg-background`。
- **pricing.tsx**：价格表用 `paper` 包裹，表头 `bg-[var(--bg-subtle)]`，单价用 `font-mono`。继续只展示用户侧价格字段（input_price_per_1m / output_price_per_1m / price_per_call），**不得引入敏感字段**。
- **dashboard.tsx**：StatCard 已在 P2 改；最近调用列表行分隔用 `border-border`。
- **usage.tsx**：图表（recharts）配色改为 clay 主色 + 中性灰网格线（`stroke="#D8D4CA"`），去掉青绿。
- **api-keys.tsx**：key mask 用 `font-mono`；按钮已在 P1 改；列表卡 `paper`。
- **models.tsx**：分组标题 + Badge（用新 badge default variant）；卡片 `paper`。
- **redeem.tsx**：兑换卡 `paper`，右侧说明区 `bg-[var(--bg-subtle)]`。
- **announcements.tsx**：用 prose（已适配）渲染 markdown，pinned 标记用 clay badge。
- **settings.tsx**：账户信息行 `bg-[var(--bg-subtle)]`（当前是 `bg-white/[0.035]`），修改密码表单沿用新 input/button。
- **measured-chart.tsx / loading-grid.tsx**：图表配色与骨架色适配。

> 注意：recharts 图表里若硬编码了青绿/紫色，统一改为 `var(--clay)` 主色和中性网格。把这些颜色集中到一个常量便于复用。

**验收 P3**：build 通过；逐页人工核对功能不回退、无残留深色/毛玻璃。可用 `git grep "glass\|shadow-glow\|translate-y\|border-white/10\|bg-white/\[" frontend/user/src` 确认无残留。

---

## 7. P4：管理端切换为同源浅色学术风

管理端基于 AntD，策略是**通过 ConfigProvider 注入令牌 + 一份全局 css 覆盖**，而非逐组件重写。

### 7.1 新建管理端全局 css

**文件**：`frontend/admin/src/styles.css`（新建），在 main.tsx 中 `import "./styles.css"`（放在 `antd/dist/reset.css` 之后）。

```css
:root {
  --bg: #F0EEE6;
  --bg-subtle: #E8E6DC;
  --surface: #FAF9F5;
  --ink: #141413;
  --ink-muted: #5F5D57;
  --border-c: #D8D4CA;
  --clay: #C6613F;
}

body {
  background: var(--bg);
  font-family: Inter, "PingFang SC", "Microsoft YaHei", ui-sans-serif, system-ui, sans-serif;
  color: var(--ink);
}

/* 标题衬线 */
h1, h2, h3, .ant-typography h1, .ant-typography h2, .ant-typography h3 {
  font-family: "Tiempos Text", "Source Han Serif SC", "Songti SC", Georgia, serif;
}
```

### 7.2 重写 admin-shell.tsx 的 Theme

把 AntD 主题令牌全面对齐设计令牌（用 `cssVar` + 完整 token）：

```tsx
import { ConfigProvider, theme as antdTheme } from "antd";
import { designTokens } from "@lingshu/shared";

export function Theme({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider
      theme={{
        cssVar: true,
        token: {
          colorPrimary: designTokens.colors.clay,
          colorInfo: designTokens.colors.clay,
          colorSuccess: designTokens.colors.success,
          colorWarning: designTokens.colors.warning,
          colorError: designTokens.colors.danger,
          colorBgLayout: designTokens.colors.bg,
          colorBgContainer: designTokens.colors.surface,
          colorBgElevated: designTokens.colors.surface,
          colorBorder: designTokens.colors.border,
          colorBorderSecondary: designTokens.colors.border,
          colorText: designTokens.colors.ink,
          colorTextSecondary: designTokens.colors.inkMuted,
          borderRadius: 6,
          fontFamily: designTokens.font.sans,
          boxShadow: designTokens.shadow.md,
          boxShadowSecondary: designTokens.shadow.sm
        },
        components: {
          Layout: {
            headerBg: designTokens.colors.surface,
            siderBg: designTokens.colors.bgSubtle,
            bodyBg: designTokens.colors.bg
          },
          Menu: {
            itemBg: "transparent",
            itemSelectedBg: designTokens.colors.claySoft,
            itemSelectedColor: designTokens.colors.clayHover,
            itemColor: designTokens.colors.inkMuted
          },
          Table: { headerBg: designTokens.colors.bgSubtle, headerColor: designTokens.colors.ink },
          Card: { colorBorderSecondary: designTokens.colors.border }
        }
      }}
    >
      {children}
    </ConfigProvider>
  );
}
```

### 7.3 AdminMenu 改浅色

`<Menu theme="dark" ...>` → `<Menu theme="light" mode="inline" ...>`（Sider 现在是 bgSubtle 浅底，菜单必须改 light）。

### 7.4 main.tsx 布局硬编码色清理

- 登录页 `background: "#f6f8fb"` → `background: designTokens.colors.bg`（或用 css 变量 `var(--bg)`）。
- 登录卡标题用衬线（可在 Card title 包一个 `<span style={{fontFamily:...}}>` 或靠全局 h css）。
- `Header style={{ background: "#fff", borderBottom: "1px solid #f0f0f0" }}` → `background: var(--surface); borderBottom: 1px solid var(--border-c)`。
- Sider 里 logo `<div style={{ color: "white", ... }}>LingShu Admin</div>`：现在 Sider 是浅底，改 `color: var(--ink)`，并把字体改衬线，弱化为品牌字标。
- `Content style={{ padding: 24 }}` 可加 `background: var(--bg)`。
- 「正在校验登录状态」加载页背景 `#f6f8fb` 同样改 `var(--bg)`。

> 提醒：admin 字体安装同 1.2，npm 安装到 admin workspace 并在 main.tsx import。

**验收 P4**：build 通过；管理端整体变为暖白底 + 浅色侧栏 + clay 主色按钮/激活，与用户端同源。

---

## 8. P5：管理端页面细节统一

逐页（`frontend/admin/src/pages/`）核对，多数靠 P4 的全局令牌自动生效，重点处理硬编码：

- **admin-dashboard.tsx**：统计卡（含「累计毛利」gross_profit —— 管理端可见，保留）数字用衬线；MiniBars 柱状色改 clay。
- **reports.tsx**：MiniBars / 图表配色改 clay + 中性；金额 fmtMoney 展示用衬线。
- **channels.tsx**：Tag 颜色 `color="green"/"red"` → 改用 token 色（healthy 用 success、异常用 danger）；检测结果 Alert 沿用 AntD info（已是浅色）。
- **users.tsx / api-keys.tsx / models.tsx / redeem.tsx / audit.tsx / announcements.tsx / settings.tsx**：核对是否有手写颜色（如 Tag color、style 内联色），统一到 token 语义色。
- **model-form.tsx**：表单沿用 AntD，自动适配。
- 所有页面中 `base_cost / rate_multiplier / gross_profit` 列**保留**（管理端应可见），仅调整呈现样式。

**MiniBars 改色**（admin-page-utils.tsx）：把柱状条颜色从当前色改为 `var(--clay)` 或 `designTokens.colors.clay`。

**验收 P5**：build 通过；管理端各表格 / 表单 / 标签视觉统一。

---

## 9. P6：收口（一致性 / 响应式 / 可达性）

### 9.1 跨端一致性走查
- 两端主色都是 clay `#C6613F`；背景都是 `#F0EEE6`；标题都是衬线；圆角都是 6px。
- 截图对比用户端 dashboard 与管理端 dashboard，确认视觉同源。

### 9.2 响应式
- 用户端 app-layout 已有 `lg:grid-cols-[220px_1fr]` + 移动端横向滚动导航，核对断点在新配色下正常。
- 管理端 AntD Layout 在窄屏下 Sider 可收起（可选：加 `breakpoint="lg" collapsedWidth="0"` 到 Sider，提升移动端体验，属轻量增强）。
- 所有页面在 375px / 768px / 1280px 三个宽度下不溢出、不错位。

### 9.3 可达性
- 文字对比度：ink `#141413` on bg `#F0EEE6` ≈ 14:1（达 AAA）；inkMuted on bg ≈ 4.8:1（达 AA）；clay 按钮白字对比度需核对（clay `#C6613F` + 白字 ≈ 3.4:1，**达 AA Large 但正文级偏低**，按钮文字用 14px+ medium 可接受，必要时把按钮文字加粗或用 `#FAF9F5`）。
- 焦点环：所有可交互元素 `focus-visible` 有 clay 环（P1 已统一）。
- Tabs 换 Radix 后键盘可达。

> 完整 WCAG 合规需人工用辅助技术测试与专家评审，本阶段只做可量化的对比度与焦点检查。

### 9.4 最终构建验证（全套）

```bash
cd D:/code/LingShu/backend && go build ./... && go vet ./... && go test ./...
cd D:/code/LingShu/frontend/packages/shared && npm run build 2>/dev/null || echo "shared 无 build 脚本，跳过"
cd D:/code/LingShu/frontend/user && npm run build
cd D:/code/LingShu/frontend/admin && npm run build
```

> 注意：两个前端 workspace 必须**分别**构建，不要并行（并行会触发误报 ENOENT）。

---

## 10. 最终验收清单

- [ ] 用户端、管理端共用同一套 designTokens（colors/font/radius）。
- [ ] 背景为暖象牙白 `#F0EEE6`，文字为暖墨黑 `#141413`，强调色为赤陶橙 `#C6613F`。
- [ ] 标题使用衬线字体，正文使用 Inter，代码使用 JetBrains Mono。
- [ ] 彻底无毛玻璃（glass）、无网格底纹（soft-grid）、无发光阴影（shadow-glow）、无 radial-gradient 背景、无 hover 位移。
- [ ] 用户端绝不展示 `base_cost / rate_multiplier / gross_profit`（重构后再次确认）。
- [ ] 管理端侧栏改为浅色，登录页 / Header 硬编码色全部清理。
- [ ] 所有页面功能与改造前等价，无业务回退。
- [ ] 三个断点（375/768/1280）响应式正常。
- [ ] 可交互元素有 clay 焦点环；正文对比度达 AA。
- [ ] 后端 build/vet/test + 两个前端 build 全部通过。

---

## 11. 与现状的取舍说明（给你决策）

1. **浅色 vs 深色**：Anthropic / Claude 主界面是**暖浅色**学术风，故本方案把用户端从强制深色翻转为浅色。如果你更想要「深色学术风」（暖深褐底 + 米色字 + clay），我可以另出一版深色令牌，但 Anthropic 官方主基调是浅色，这里按其风格走浅色。**若你想保留深色，请告知，我改 P0 令牌即可，后续阶段不变。**
2. **衬线标题**：Anthropic 用自有衬线字体 Tiempos，无法商用获取，方案用 Georgia + 思源宋体高质量替代。若你有 Tiempos 授权字体文件，可放入并替换字体栈。
3. **管理端不重写为 Tailwind**：保留 AntD，仅通过 ConfigProvider 令牌 + 全局 css 对齐风格，成本最低、回退风险最小。彻底 Tailwind 化工作量极大且无必要。
4. **暂不做深浅主题切换**：当前定为单一浅色主题，保持克制。若未来要切换再加 ThemeProvider。
