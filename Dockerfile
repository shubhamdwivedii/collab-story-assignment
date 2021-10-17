# Build-Stage 
FROM golang:alpine AS build-stage
COPY . /app 
WORKDIR /app 
RUN go build -o bin/server main.go 


# Production-Stage 
FROM alpine 
COPY --from=build-stage /app/bin /collab/
COPY --from=build-stage /app/docker-entrypoint.sh /collab/
COPY --from=build-stage /app/wait-for /collab/
EXPOSE 8080 
CMD /collab/server
# CMD will be overwritten by docker-entrypoint.sh 