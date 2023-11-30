FROM golang:1.20 as build



COPY . .

RUN go build main.go -o scaler

CMD './scaler'

