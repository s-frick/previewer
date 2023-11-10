FROM alpine:3.9 
RUN apk add ca-certificates

WORKDIR /app

COPY app .

EXPOSE 5149

CMD ["/app/app"]
