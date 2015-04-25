FROM golang:1.4.2-onbuild
ENV MEDIA_FETCHER_TARGET /downloads
ENV MEDIA_FETCHER_LOG /downloads/media-fetcher.log
ENV REDIS_URL redis://redis:6379
ENV MEDIA_FETCHER_PROCESSES 1
VOLUME /downloads
