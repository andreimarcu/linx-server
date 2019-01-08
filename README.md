
linx-server
======
[![Build Status](https://travis-ci.org/andreimarcu/linx-server.svg?branch=master)](https://travis-ci.org/andreimarcu/linx-server)  

Self-hosted file/media sharing website.  


### Features

- Display common filetypes (image, video, audio, markdown, pdf)  
- Display syntax-highlighted code with in-place editing
- Documented API with keys if need to restrict uploads (can use [linx-client](https://github.com/andreimarcu/linx-client) for uploading through command-line)
- Torrent download of files using web seeding
- File expiry, deletion key, and random filename options


### Screenshots
<img width="230" src="https://cloud.githubusercontent.com/assets/4650950/10530123/4211e946-7372-11e5-9cb5-9956c5c49d95.png" /> <img width="230" src="https://cloud.githubusercontent.com/assets/4650950/10530124/4217db8a-7372-11e5-957d-b3abb873dc80.png" />  
<img width="230" src="https://cloud.githubusercontent.com/assets/4650950/10530844/48d6d4e2-7379-11e5-8886-d4c32c416cbc.png" /> <img width="230" src="https://cloud.githubusercontent.com/assets/4650950/10530845/48dc9ae4-7379-11e5-9e59-959f7c40a573.png" /> <img width="230" src="https://cloud.githubusercontent.com/assets/4650950/10530846/48df08ec-7379-11e5-89f6-5c3f6372384d.png" />   


Get release and run
-------------------
1. Grab the latest binary from the [releases](https://github.com/andreimarcu/linx-server/releases)
2. Run ```./linx-server```

  
Usage
-----

#### Configuration
All configuration options are accepted either as arguments or can be placed in an ini-style file as such:  
```ini
maxsize = 4294967296
allowhotlink = true
# etc
```  
...and then invoke ```linx-server -config path/to/config.ini```  

#### Options
- ```-bind 127.0.0.1:8080``` -- what to bind to  (default is 127.0.0.1:8080)
- ```-sitename myLinx``` -- the site name displayed on top (default is inferred from Host header)
- ```-siteurl "http://mylinx.example.org/"``` -- the site url (default is inferred from execution context)
- ```-filespath files/``` -- Path to store uploads (default is files/)
- ```-metapath meta/``` -- Path to store information about uploads (default is meta/)
- ```-maxsize 4294967296``` -- maximum upload file size in bytes (default 4GB)
- ```-maxexpiry 86400``` -- maximum expiration time in seconds (default is 0, which is no expiry)
- ```-allowhotlink``` -- Allow file hotlinking
- ```-contentsecuritypolicy "..."``` -- Content-Security-Policy header for pages (default is "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';")
- ```-filecontentsecuritypolicy "..."``` -- Content-Security-Policy header for files (default is "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';")
- ```-refererpolicy "..."``` -- Referrer-Policy header for pages (default is "same-origin")
- ```-filereferrerpolicy "..."``` -- Referrer-Policy header for files (default is "same-origin")
- ```-xframeoptions "..." ``` -- X-Frame-Options header (default is "SAMEORIGIN")
- ```-remoteuploads``` -- (optionally) enable remote uploads (/upload?url=https://...) 
- ```-nologs``` -- (optionally) disable request logs in stdout

#### SSL with built-in server 
- ```-certfile path/to/your.crt``` -- Path to the ssl certificate (required if you want to use the https server)
- ```-keyfile path/to/your.key``` -- Path to the ssl key (required if you want to use the https server)

#### Use with http proxy 
- ```-realip``` -- let linx-server know you (nginx, etc) are providing the X-Real-IP and/or X-Forwarded-For headers.

#### Use with fastcgi
- ```-fastcgi``` -- serve through fastcgi 

#### Require API Keys for uploads
- ```-authfile path/to/authfile``` -- (optionally) require authorization for upload/delete by providing a newline-separated file of scrypted auth keys
- ```-remoteauthfile path/to/remoteauthfile``` -- (optionally) require authorization for remote uploads by providing a newline-separated file of scrypted auth keys

A helper utility ```linx-genkey``` is provided which hashes keys to the format required in the auth files.


Cleaning up expired files
-------------------------
When files expire, access is disabled immediately, but the files and metadata
will persist on disk until someone attempts to access them. If you'd like to
automatically clean up files that have expired, you can use the included
`linx-cleanup` utility. To run it automatically, use a cronjob or similar type
of scheduled task.

You should be careful to ensure that only one instance of `linx-client` runs at
a time to avoid unexpected behavior. It does not implement any type of locking.

#### Options
- ```-filespath files/``` -- Path to stored uploads (default is files/)
- ```-metapath meta/``` -- Path to stored information about uploads (default is meta/)
- ```-nologs``` -- (optionally) disable deletion logs in stdout


Deployment
----------
Linx-server supports being deployed in a subdirectory (ie. example.com/mylinx/) as well as on its own (example.com/).


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

#### 4. Using Docker with the built-in http server
First, build the image:
```docker build -t linx-server .```

You'll need some directories for the persistent storage. For the purposes of this example, we will use `/media/meta` and `/media/files`.

Then, run it:
```docker run -p 8080:8080 -v /media/meta:/data/meta -v /media/files:/data/files linx-server```


Development
-----------
Any help is welcome, PRs will be reviewed and merged accordingly.  
The official IRC channel is #linx on irc.oftc.net  

1. ```go get -u github.com/andreimarcu/linx-server ```
2. ```cd $GOPATH/src/github.com/andreimarcu/linx-server ```
3. ```go build && ./linx-server```


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
