FROM dpokidov/imagemagick
WORKDIR /opt
COPY heif-converter /opt/heif-converter
RUN "mkdir data"
ENV WATCH="/opt/data/"
ENTRYPOINT ["/opt/heif-converter"]