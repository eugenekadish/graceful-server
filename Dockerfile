FROM golang:alpine

COPY . .

EXPOSE 80
CMD ["go", "run", "main.go"]