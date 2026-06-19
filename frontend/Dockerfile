# 灵枢前端生产镜像：构建单一 app SPA，由 Nginx 静态托管并反代后端。
# 管理端挂在同一 app 的 /admin/* 路由下，统一 80 端口。

FROM node:20-alpine AS build
WORKDIR /app

# 先拷 workspace 清单安装依赖。
# 注意：锁文件是在 Windows 上生成的，optionalDependencies 里只锁了 win32 版
# Rollup 原生二进制。容器是 Linux（且本服务器为 arm64），必须删掉锁文件让 npm
# 完全按容器真实平台（linux-arm64-musl 等）重新解析，否则平台相关二进制漏装，
# 构建报 MODULE_NOT_FOUND: rollup/dist/native.js（npm/cli#4828）。
COPY package.json ./
COPY packages/shared/package.json packages/shared/
COPY packages/ui/package.json packages/ui/
COPY app/package.json app/
RUN npm install --no-audit --no-fund

# 拷源码并构建。前端以同源（空 base url）请求 /api、/v1，由 Nginx 反代到后端。
COPY . .
ENV VITE_API_BASE_URL=""
RUN npm --workspace @lingshu/app run build

FROM nginx:1.27-alpine AS runtime
# 单 app 站点根目录
COPY --from=build /app/app/dist /usr/share/nginx/html/app
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
