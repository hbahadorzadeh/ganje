# Development Dockerfile for Angular application
FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including dev dependencies for ng serve)
RUN npm ci

# Copy source code
COPY . .

# Expose port
EXPOSE 4200

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4200/ || exit 1

# Default command (can be overridden by docker-compose)
CMD ["npm", "start", "--", "--host", "0.0.0.0", "--disable-host-check"]
