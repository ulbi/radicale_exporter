# radicale_exporter

A very simple Prometheus exporter for the [Radicale CalDAV/CardDAV server](http://radicale.org)

Events are read from Radicale's logfile and a metric `radicale_requests` is exported providing counters about the various HTTP request types (`PROPFIND, PUT, DELETE, ...`) received by the server. Additionally, a metric `radicale_up` tells if the exporter could reach Radicale via TCP.

Metrics can be scraped via the `/metrics` endpoint.

## Usage / default values
```
usage: radicale_exporter [<flags>]

Flags:
  -l, --listen=":9191"                              address:port to serve /metrics on
  -i, --inputfile="/var/log/radicale/radicale.log"  radicale log file
  -s, --scrapeinterval=15                           Prometheus scrape interval
  -r, --radicale=":5232"                            address:port to contact Radicale on
```

## Known Issues
- The exporter will get confused when Radicale's logfile is moved or truncated (i.e. rotated). Make sure to restart the exporter when using logrotate or the like.

## Supported versions
The exporter was tested with Radicale 1.1.1 and 1.1.6. It _should_ work with any version that logs to a file with format:  
`2018-03-28 13:42:57,286 - INFO: PROPFIND request at /user/MyCalendar/ received`
