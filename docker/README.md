# linx-server

A simple [Docker](http://docker.com) image for [linx-server](https://github.com/andreimarcu/linx-server).

[![](https://images.microbadger.com/badges/image/codehat/linx-server.svg)](http://microbadger.com/images/codehat/linx-server "Get your own image badge on microbadger.com")

## Getting Started

```sh
$ docker run -d -p 8080:8080 codehat/linx-server
```

Point your browser to `http://127.0.0.1:8080`.

## Usage

#### Default config.ini

The image contains a default config.ini.

```
bind = 0.0.0.0:8080
sitename = "My Linx Server"
# siteurl = "https://yoururl.example.com/"
filespath = "/srv/storage/files/"
metapath = "/srv/storage/meta/"
nologs = true
# certfile = "/srv/config/your.crt"
# keyfile = "/srv/config/your.key"
# authfile = "/srv/config/authfile"
```

#### Using local config.ini and keep storage folder

Replace `/path/to/storage` and `/path/to/config` accordingly.
Also make sure that a `config.ini` exists in your `/path/to/config` folder and it is properly configured!

```sh
$ docker run -d \
    -p 8080:8080 \
    -v /path/to/storage:/srv/storage \
    -v /path/to/config:/srv/config \
    codehat/linx-server
```

#### Enable API

Use your own `config.ini` and uncomment the following line:
`# authfile = "/srv/config/authfile"`
gets to:
`authfile = "/srv/config/authfile"`

Now generate an API key with the `linx-genkey` tool and put the key into the `/path/to/config/authfile`.
The next time you start your container, API will be enabled.