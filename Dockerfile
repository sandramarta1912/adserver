FROM iron/go

WORKDIR /app

RUN mkdir /app/tpl

COPY tpl/* /app/tpl/
COPY adserver /app/

ENTRYPOINT ["./adserver"]