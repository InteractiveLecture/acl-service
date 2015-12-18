FROM alpine:3.1 
ADD out/main /main
cmd ["/main"]
EXPOSE 8080 
