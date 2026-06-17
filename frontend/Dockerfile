# 灵枢前端生产镜像：构建 user + admin 两个 SPA，由 Nginx 静态托管并反代后端。
# 用户端是公开门面（端口 80），管理端走独立端口 8081（内部使用）。

FROM node:20-alpine AS build
WORKDIR /app

# 先拷 workspace 清单，利用缓存安装依赖
COPY package.json package-lock.json ./
COPY packages/shared/package.json packages/shared/
COPY user/package.json user/
COPY admin/package.json admin/
RUN npm ci

# 拷源码并构建。两端均以同源（空 base url）请求 /api、/v1，由 Nginx 反代到后端。
COPY . .
ENV VITE_API_BASE_URL=""
RUN npm --workspace @lingshu/user run build \
 && npm --workspace @lingshu/admin run build

FROM nginx:1.27-alpine AS runtime
# user 站点根目录、admin 站点根目录
COPY --from=build /app/user/dist /usr/share/nginx/html/user
COPY --from=build /app/admin/dist /usr/share/nginx/html/admin
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80 8081
CMD ["nginx", "-g", "daemon off;"]
