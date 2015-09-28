Linx Server
======
[![Build Status](https://travis-ci.org/andreimarcu/linx-server.svg?branch=master)](https://travis-ci.org/andreimarcu/linx-server)  


Soon-to-be opensource replacement of Linx  

**Please do not use yet! Consider it in pre-alpha development stages.**

Build & Run
----------------

1. ```go get -u github.com/andreimarcu/linx-server ```
2. ```cd $GOPATH/src/github.com/andreimarcu/linx-server ```
3. ```go build && ./linx-server```

By default linx will bind to ```http://127.0.0.1:8080/```, use the "files/" files directory and the "linx" sitename.  
Configurable flags can be found in ```server.go```.



Development
-----------
Any help is welcome, PRs will be reviewed and merged accordingly.  
The official IRC channel is #linx on irc.oftc.net  


TODO
----
Please refer to the [main TODO issue](https://github.com/andreimarcu/linx-server/issues/1) 

Author
-------
Andrei Marcu, http://andreim.net/
