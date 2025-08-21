# Build stage
FROM node:18-alpine AS build
WORKDIR /app

# Install dependencies
COPY package.json package-lock.json* .npmrc* ./
RUN npm ci --no-audit --no-fund

# Copy source
COPY . .

# Build production bundle (browser build)
RUN npx ng build --configuration production --output-hashing=all --ssr=false

# Runtime stage
FROM nginx:1.27-alpine

# Default backend upstream; overridden by Helm via ConfigMap
ENV BACKEND_HOST=backend BACKEND_PORT=8080

# Copy built assets
COPY --from=build /app/dist/admin-portal/browser /usr/share/nginx/html

# Copy default NGINX config (can be overridden by mounted ConfigMap)
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
