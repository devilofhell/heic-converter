FROM dpokidov/imagemagick
WORKDIR /opt
COPY heif-converter /opt/heif-converter
CMD [ "mkdir data" ]
ENV WATCH="/opt/data/"
ENV TIME_BETWEEN="1h"
ENV KEEP_ORIGINAL="false"
ENV KEEP_LIVE_PHOTO="false"
ENTRYPOINT ["/opt/heif-converter"]