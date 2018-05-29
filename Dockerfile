FROM iron/go

WORKDIR /app

RUN mkdir /app/tpl

COPY tpl/* /app/tpl/
COPY server /app/

ENTRYPOINT ["./server"]