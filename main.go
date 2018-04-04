package main

import (
	"bufio"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/danielb42/handlekeeper"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	prm_listenAddr   string
	prm_inputFile    string
	prm_scrapeIntvl  int
	prm_radicaleAddr string
	prm_debug        bool

	mtr_up = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "radicale",
			Name:      "up",
			Help:      "shows if radicale could be reached via tcp",
		},
	)

	mtr_requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "radicale",
			Name:      "requests",
			Help:      "number of requests to ressouces",
		},
		[]string{
			"type",
			"ressource",
		},
	)
)

func inspectLine(line string) {
	pattern := `INFO: (.+) request at (.+/.+)/(.+)? received`
	result := regexp.MustCompile(pattern).FindStringSubmatch(line)

	if prm_debug {
		println("parsed: ", line)
	}

	if result != nil {
		mtr_requests.WithLabelValues(result[1], result[2]).Inc()
	}
}

func checkTCP() bool {
	conn, err := net.Dial("tcp", prm_radicaleAddr)

	if err != nil {
		mtr_up.Set(0)
		return false
	} else {
		conn.Close()
		mtr_up.Set(1)
		return true
	}
}

func serveMetrics() {
	prometheus.MustRegister(mtr_up)
	prometheus.MustRegister(mtr_requests)

	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(prm_listenAddr, nil)
}

func parseFlags() {
	kingpin.Flag("listen", "address:port to serve /metrics on").Short('l').Default(":9191").StringVar(&prm_listenAddr)
	kingpin.Flag("inputfile", "exporter input file (truncated!)").Short('i').Default("/var/log/radicale/radicale_exporter_input.log").StringVar(&prm_inputFile)
	kingpin.Flag("scrapeinterval", "Prometheus scrape interval").Short('s').Default("15").IntVar(&prm_scrapeIntvl)
	kingpin.Flag("radicale", "address:port to contact Radicale on").Short('r').Default(":5232").StringVar(&prm_radicaleAddr)
	kingpin.Flag("debug", "more output and skip TCP socket check").Short('d').Default("false").BoolVar(&prm_debug)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()
}

func main() {
	parseFlags()
	serveMetrics()

	hk := handlekeeper.NewHandlekeeper(prm_inputFile)
	defer hk.Close()

	for {
		if prm_debug || checkTCP() {
			scanner := bufio.NewScanner(hk.Handle)

			for scanner.Scan() {
				inspectLine(scanner.Text())
			}

			hk.Handle.Truncate(0)
			hk.Handle.Seek(0, 0)
		}

		time.Sleep(time.Duration(prm_scrapeIntvl) * time.Second)
	}
}
