FROM golang:alpine AS builder

WORKDIR /app

# Installer les dépendances nécessaires
RUN apk add --no-cache git

# Copier les sources dans l'image
COPY ./go ./

# Build de l'exécutable
RUN go build -o pocketbase ./cmd/pocketbase


# === Image finale minimale ===
FROM denoland/deno:alpine

WORKDIR /app

# Pour exécuter les binaires Go
RUN apk add --no-cache ca-certificates ffmpeg

# Copier le binaire buildé et le start script
COPY --from=builder /app/pocketbase /app/pocketbase

# Ajouter permissions d'exécution
RUN chmod +x /app/pocketbase

EXPOSE 8080

CMD ["/app/pocketbase", "serve", "--http=0.0.0.0:8090"]
