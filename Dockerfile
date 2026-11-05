FROM golang:1.26-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cv-search .

FROM alpine:3.22
RUN adduser -D -H app
WORKDIR /app
COPY --from=build /out/cv-search /app/cv-search
COPY frontend /app/frontend
ENV FRONTEND_DIR=/app/frontend
EXPOSE 8080
USER app
ENTRYPOINT ["/app/cv-search"]
