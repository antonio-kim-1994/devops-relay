FROM alpine:latest
RUN apk add --no-cache tzdata
WORKDIR /app
COPY ./devops-relay-gateway /app/devops-relay-gateway
ENTRYPOINT ["/app/devops-relay-gateway"]
EXPOSE 8080