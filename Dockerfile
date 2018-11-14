FROM golang:1.10 as build
WORKDIR /go/src/github.com/antimoth/estimator
COPY ./estimator /go/src/github.com/antimoth/estimator/estimator
COPY ./selfconf /go/src/selfconf
COPY ./main.go /go/src/github.com/antimoth/estimator
RUN go build -a -o runEstimator main.go
CMD ["./runEstimator"]

FROM alpine as prod
COPY --from=build /go/src/github.com/antimoth/estimator/runEstimator .
COPY --from=build /lib/x86_64-linux-gnu/libpthread-2.24.so /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=build /lib/x86_64-linux-gnu/libc-2.24.so  /lib/x86_64-linux-gnu/libc.so.6
COPY --from=build /lib/x86_64-linux-gnu/ld-2.24.so /lib64/ld-linux-x86-64.so.2
CMD ["./runEstimator"]
