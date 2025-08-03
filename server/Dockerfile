FROM alpine:latest
RUN apk add --no-cache tzdata
WORKDIR /app
COPY ./devops-relay-server /app/devops-relay-server
ENTRYPOINT ["/app/devops-relay-server"]
EXPOSE 8080