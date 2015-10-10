
Linx Server
======
[![Build Status](https://travis-ci.org/andreimarcu/linx-server.svg?branch=master)](https://travis-ci.org/andreimarcu/linx-server)  


Soon-to-be opensource replacement of Linx (media-sharing website)

**Consider it in pre-alpha development stages.**

Get release and run
-------------------
1. Grab the latest binary from the [releases](https://github.com/andreimarcu/linx-server/releases)
2. Run ```./linx-server...```
  
  
Command-line options
--------------------

- ```-bind 127.0.0.1:8080``` -- what to bind to  (default is 127.0.0.1:8080)
- ```-sitename myLinx``` -- the site name displayed on top (default is linx)
- ```-siteurl "http://mylinx.example.org/"``` -- the site url (for generating links)
- ```-filespath files/"``` -- Path to store uploads (default is files/)
- ```-metapath meta/``` -- Path to store information about uploads (default is meta/)
- ```-maxsize 4294967296``` maximum upload file size in bytes (default 4GB)
- ```-certfile path/to/your.crt``` -- Path to the ssl certificate (required if you want to use the https server)
- ```-keyfile path/to/your.key``` -- Path to the ssl key (required if you want to use the https server)
- ```-contentsecuritypolicy "..."``` -- Content-Security-Policy header for pages (default is "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; referrer none;")
- ```-filecontentsecuritypolicy "..."``` -- Content-Security-Policy header for files (default is "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; sandbox; referrer none;"")
- ```-xframeoptions "..." ``` -- X-Frame-Options header (default is "SAMEORIGIN")
- ```-remoteuploads``` -- (optionally) enable remote uploads (/upload?url=https://...) 
- ```-fastcgi``` -- (optionally) serve through fastcgi 
- ```-nologs``` -- (optionally) disable request logs in stdout

Deployment
----------

#### 1. Using fastcgi

A suggested deployment is running nginx in front of linx-server serving through fastcgi.
This allows you to have nginx handle the TLS termination for example.  
An example configuration:
```
server {
    ...
    server_name yourlinx.example.org;
    ...
    
    client_max_body_size 4096M;
    location / {
        fastcgi_pass 127.0.0.1:8080;
        include fastcgi_params;
    }
}
```
And run linx-server with the ```-fastcgi``` option.

#### 2. Using the built-in https server
Run linx-server with the ```-certfile path/to/cert.file``` and ```-keyfile path/to/key.file``` options.

#### 3. Using the built-in http server
Run linx-server normally.

Development
-----------
Any help is welcome, PRs will be reviewed and merged accordingly.  
The official IRC channel is #linx on irc.oftc.net  

1. ```go get -u github.com/andreimarcu/linx-server ```
2. ```cd $GOPATH/src/github.com/andreimarcu/linx-server ```
3. ```go build && ./linx-server```


TODO
----
Please refer to the [main TODO issue](https://github.com/andreimarcu/linx-server/issues/1) 


License
-------
Copyright (C) 2015 Andrei Marcu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

Author
-------
Andrei Marcu, http://andreim.net/
