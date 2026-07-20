FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml tsconfig.base.json ./
COPY packages ./packages
COPY apps/store-web ./apps/store-web
RUN pnpm install --frozen-lockfile || pnpm install
RUN pnpm --filter store-web build

FROM nginx:alpine
COPY apps/store-web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/store-web/dist /usr/share/nginx/html
EXPOSE 80
