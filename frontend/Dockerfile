# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Install pnpm globally
RUN npm install -g pnpm

# Copy package files
COPY package*.json ./
COPY pnpm-lock.yaml ./

# Install all dependencies (including dev dependencies for build)
RUN pnpm install --frozen-lockfile --shamefully-hoist

# Copy source code
COPY . .

RUN chmod -R 755 node_modules/.bin

# Accept build arguments and set as environment variables
ARG PUBLIC_AUTH_API_BASE_URL=https://api.wili.me
ARG PUBLIC_USER_API_BASE_URL=https://api.wili.me
ARG PUBLIC_WISHLIST_API_BASE_URL=https://api.wili.me/wishlists
ARG PUBLIC_TELEGRAM_BOT_USERNAME
ARG PUBLIC_TELEGRAM_WEBAPP_URL
ARG VITE_YANDEX_CLIENT_ID
ENV PUBLIC_AUTH_API_BASE_URL=$PUBLIC_AUTH_API_BASE_URL
ENV PUBLIC_USER_API_BASE_URL=$PUBLIC_USER_API_BASE_URL
ENV PUBLIC_WISHLIST_API_BASE_URL=$PUBLIC_WISHLIST_API_BASE_URL
ENV PUBLIC_TELEGRAM_BOT_USERNAME=$PUBLIC_TELEGRAM_BOT_USERNAME
ENV PUBLIC_TELEGRAM_WEBAPP_URL=$PUBLIC_TELEGRAM_WEBAPP_URL
ENV VITE_YANDEX_CLIENT_ID=$VITE_YANDEX_CLIENT_ID

# Build the application
RUN pnpm run build

# Remove dev dependencies after build
RUN echo "y" | pnpm prune --prod

# Production stage
FROM node:18-alpine AS runner

WORKDIR /app

# Copy package files and production dependencies from builder
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules

# Copy built application from builder stage
COPY --from=builder /app/build ./build

# Expose port
EXPOSE 3000

# Set environment to production
ENV NODE_ENV=production

# Command to run the SvelteKit server
CMD ["node", "build/index.js"]