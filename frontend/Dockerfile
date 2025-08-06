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

# Accept build argument and set as environment variable
ARG PUBLIC_API_BASE_URL=https://api.wili.me
ENV PUBLIC_API_BASE_URL=$PUBLIC_API_BASE_URL

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