
Linx Server
======
[![Build Status](https://travis-ci.org/andreimarcu/linx-server.svg?branch=master)](https://travis-ci.org/andreimarcu/linx-server)  


Soon-to-be opensource replacement of Linx (media-sharing website)

**Consider it in pre-alpha development stages.**

Get release and run
-------------------
1. Grab the latest binary from the [releases](https://github.com/andreimarcu/linx-server/releases)
2. Run ```./linx-server-v...```
  
  
Command-line options
--------------------

- Specify what to bind to ```-b 0.0.0.0:8080```
- Specify the sitename ```-sitename myLinx```
- Specify the siteurl (for generating link) ```-siteurl "http://mylinx.example.org/"```
- Specify the filespath (where files are uploaded to) ```-filespath files/"```
- Specify the metapath (where metadata for files are stored) ```-metapath meta/```
- Optionally: Specify to disable request logs in stdout ```-nologs```


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
