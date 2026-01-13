FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED 1

# Update build timestamps (prebuild script will also run, but this ensures it happens)
RUN node -e "const fs=require('fs');const now=new Date();const thaiTime=new Date(now.getTime()+7*60*60*1000);const buildDate=thaiTime.toISOString().slice(0,19).replace('T',' ');const buildTime=thaiTime.toISOString().slice(0,19)+'+07:00';const updateVersion=(file)=>{if(fs.existsSync(file)){const data=JSON.parse(fs.readFileSync(file,'utf8'));data.buildDate=buildDate;data.buildTime=buildTime;fs.writeFileSync(file,JSON.stringify(data,null,2)+'\n');}};updateVersion('VERSION.json');updateVersion('public/VERSION.json');" || true

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 4000

ENV PORT 4000

CMD ["node", "server.js"]


