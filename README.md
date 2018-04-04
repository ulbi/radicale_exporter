# radicale_exporter

A very simple Prometheus exporter for the [Radicale CalDAV/CardDAV server](http://radicale.org)

The exporter presents a metric `radicale_requests` which provides counters about the various HTTP request types (`PROPFIND, PUT, DELETE, ...`). Additionally, a metric `radicale_up` tells if the exporter was able to contact the Radicale server via its TCP socket.

Events are read from a duplicated Radicale logfile. See the `logging` file for how to configure logging to two targets. Metrics can be scraped via the `/metrics` endpoint. 

Note that **the input file will be truncated** with each scrape to avoid duplicate events.

## Usage / default values
```
usage: radicale_exporter [<flags>]

Flags:
  -l, --listen=":9191"                                             address:port to serve /metrics on
  -i, --inputfile="/var/log/radicale/radicale_exporter_input.log"  exporter input file (truncated!)
  -s, --scrapeinterval=15                                          Prometheus scrape interval
  -r, --radicale=":5232"                                           address:port to contact Radicale on
  -d, --debug                                                      more output and skip TCP socket check
```

## Supported versions
The exporter was tested with Radicale 1.1.1 and 1.1.6. It *should* work with any version that logs to a file with format:  
`2018-03-28 13:42:57,286 - INFO: PROPFIND request at /user/MyCalendar/ received`
