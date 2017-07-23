# passunix: a library for passing sockets over unix domain sockets

People often conflate load balancers, request routers, and proxies because most
software implements load balancing and routing by proxying connections. This is
a powerful and general way to do the job.

But. If all you want to do is decide which process should handle a given connection, you can
skip the proxy part and just pass off the connection. This is vastly more efficient, leaving
your servers free to do more important things.
