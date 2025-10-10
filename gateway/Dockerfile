# Build stage
FROM node:24-alpine AS builder

# RUN apt update && apt install -y protobuf-compiler make
RUN apk add --no-cache protobuf protobuf-dev make

WORKDIR /app

COPY gateway/package.json gateway/yarn.lock ./

RUN yarn install

COPY gateway/Makefile ./
COPY protos/ ../protos
RUN mkdir protos && make proto

COPY gateway/ .

RUN yarn build
RUN yarn install --production --frozen-lockfile --offline

# Production stage
FROM node:24-alpine AS production

WORKDIR /app

COPY --from=builder /app/package.json /app/yarn.lock ./

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

CMD ["yarn", "start:prod"]