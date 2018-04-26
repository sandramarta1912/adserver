FROM iron/go

WORKDIR /app

ADD server /app/

ENTRYPOINT ["./server"]