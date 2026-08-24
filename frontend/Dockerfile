FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
FROM node:20-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
# Railway's private network resolves service hostnames to IPv6 only; prefer AAAA
# so the /api rewrite can reach the backend over the internal domain.
ENV NODE_OPTIONS=--dns-result-order=ipv6first
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/package.json ./package.json
COPY --from=build /app/node_modules ./node_modules
EXPOSE 3000
CMD ["npm", "start"]
