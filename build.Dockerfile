FROM node:18 as node

COPY . /repo
WORKDIR /repo
RUN cd frontend && npm install && npm run build

FROM golang:1.19.1-bullseye as go

COPY . /repo
COPY --from=node /repo/frontend/build /repo/backend/build
WORKDIR /repo/backend
ENV CGO_ENABLED=1

RUN go build
