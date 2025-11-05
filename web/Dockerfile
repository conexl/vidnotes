# syntax=docker/dockerfile:1

FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# Ensure API URL can be set at build time (fallback to /api/v1 for reverse proxy)
ENV NEXT_PUBLIC_API_URL=/api/v1
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=builder /app/.next ./.next
COPY package.json ./package.json
# Minimal deps for prod runtime
RUN npm pkg delete scripts.prepare || true
RUN npm i --omit=dev
EXPOSE 3000
CMD ["npm", "run", "start"]