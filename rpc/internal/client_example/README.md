# Client example

This a client that simulates a gRPC consumer. We're using this for the time being to document the interaction with the
gRPC interface.

To use it run `arduino-cli daemon` and then `client_example`.

To test the proxy settings first run:

```
docker run --name squid -d --restart=always \
  --publish 3128:3128 \
  --volume /path/to/squid.conf:/etc/squid/squid.conf \
  --volume /srv/docker/squid/cache:/var/spool/squid \
  sameersbn/squid:3.5.27-2
```

The `squid.conf` file to use is in this directory so change the volume path to that.

To verify that requests are passing through the local proxy run:

```
docker exec -it squid tail -f /var/log/squid/access.log
```

If it works you should see logs similar to this:

```
1612176447.893 400234 172.17.0.1 TCP_TUNNEL/200 116430 CONNECT downloads.arduino.cc:443 - HIER_DIRECT/104.18.28.45 -
1612176448.197 400245 172.17.0.1 TCP_TUNNEL/200 1621708 CONNECT downloads.arduino.cc:443 - HIER_DIRECT/104.18.28.45 -
1612176448.946 400256 172.17.0.1 TCP_TUNNEL/200 354882 CONNECT downloads.arduino.cc:443 - HIER_DIRECT/104.18.28.45 -
```
