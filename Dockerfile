FROM alpine:3.9 
RUN apk add ca-certificates

WORKDIR /app
RUN mkdir src
RUN mkdir templates

COPY app .
COPY entrypoint.sh .

EXPOSE 5149

# CMD ["/app/app", "-src \"/app/src\"", "-templates \"/app/templates\""]
ENTRYPOINT ["/app/entrypoint.sh"]
