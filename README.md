# Media Fetcher

A simple service that monitors a Redis queue, for files to download, and
downloads them.

## Usage

```shell
$ docker pull redis
$ docker pull cwninja/media-fetcher

$ docker run -d --name redis redis
$ docker run -d --name media-fetcher --link redis:redis -v ~/Downloads/:/downloads/ cwninja/media-fetcher
```

Then push a JSON object like the following onto the redis `download-queue` list:

```shell
$ docker run --rm -it --link redis:r redis redis-cli -h r
r:6379> rpush download-queue '{"url":"http://example.com","filename":"example.com.html"}'
```

The media fetcher will pull down the file, and store it in
`~/Downloads/example.com.html`
