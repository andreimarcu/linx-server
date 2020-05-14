
linx-cleanup
-------------------------
When files expire, access is disabled immediately, but the files and metadata
will persist on disk until someone attempts to access them. 

If you'd like to automatically clean up files that have expired, you can use the included `linx-cleanup` utility. To run it automatically, use a cronjob or similar type
of scheduled task.

You should be careful to ensure that only one instance of `linx-cleanup` runs at
a time to avoid unexpected behavior. It does not implement any type of locking.


|Option|Description
|------|-----------
| ```-filespath files/``` | Path to stored uploads (default is files/)
| ```-nologs``` | (optionally) disable deletion logs in stdout
| ```-metapath meta/``` | Path to stored information about uploads (default is meta/)

