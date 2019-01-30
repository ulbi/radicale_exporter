package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/danielb42/handlekeeper"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	prmListenAddr   string
	prmInputFile    string
	prmScrapeIntvl  int
	prmRadicaleAddr string
	prmDebug        bool

	mtrUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "radicale",
			Name:      "up",
			Help:      "shows if radicale could be reached via tcp",
		},
	)

	mtrRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "radicale",
			Name:      "requests",
			Help:      "number of requests to ressources",
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

	if prmDebug {
		println("parsed: ", line)
	}

	if result != nil {
		mtrRequests.WithLabelValues(result[1], result[2]).Inc()
	}
}

func checkTCP() bool {
	conn, err := net.Dial("tcp", prmRadicaleAddr)

	if err != nil {
		mtrUp.Set(0)
		return false
	}

	conn.Close()
	mtrUp.Set(1)
	return true
}

func serveMetrics() {
	prometheus.MustRegister(mtrUp)
	prometheus.MustRegister(mtrRequests)

	http.Handle("/metrics", prometheus.Handler())
	go func() {
		log.Fatalf("Metrics server crashed: %s", http.ListenAndServe(prmListenAddr, nil))
	}()
}

func parseFlags() {
	kingpin.Flag("listen", "address:port to serve /metrics on").Short('l').Default(":9191").StringVar(&prmListenAddr)
	kingpin.Flag("inputfile", "exporter input file (truncated!)").Short('i').Default("/var/log/radicale/radicale_exporter_input.log").StringVar(&prmInputFile)
	kingpin.Flag("scrapeinterval", "Prometheus scrape interval").Short('s').Default("15").IntVar(&prmScrapeIntvl)
	kingpin.Flag("radicale", "address:port to contact Radicale on").Short('r').Default(":5232").StringVar(&prmRadicaleAddr)
	kingpin.Flag("debug", "more output and skip TCP socket check").Short('d').Default("false").BoolVar(&prmDebug)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()
}

func main() {
	parseFlags()
	serveMetrics()

	hk, err := handlekeeper.NewHandlekeeper(prmInputFile)
	if err != nil {
		log.Fatalln("could not open input file")
	}
	defer hk.Close()

	for {
		if prmDebug || checkTCP() {
			scanner := bufio.NewScanner(hk.Handle)

			for scanner.Scan() {
				inspectLine(scanner.Text())
			}

			hk.Handle.Truncate(0)
			hk.Handle.Seek(0, 0)
		}

		time.Sleep(time.Duration(prmScrapeIntvl) * time.Second)
	}
}
