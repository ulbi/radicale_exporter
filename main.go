package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	prm_listenAddr   string
	prm_inputFile    string
	prm_scrapeint    int
	prm_radicaleAddr string

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
	kingpin.Flag("inputfile", "radicale log file").Short('i').Default("/var/log/radicale/radicale.log").StringVar(&prm_inputFile)
	kingpin.Flag("scrapeinterval", "Prometheus scrape interval").Short('s').Default("15").IntVar(&prm_scrapeint)
	kingpin.Flag("radicale", "address:port to contact Radicale on").Short('r').Default(":6379").StringVar(&prm_radicaleAddr)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()
}

func main() {
	parseFlags()
	serveMetrics()

	f, err := os.OpenFile(prm_inputFile, os.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for {
		if checkTCP() {
			scanner := bufio.NewScanner(f)

			for scanner.Scan() {
				inspectLine(scanner.Text())
			}
		}

		time.Sleep(time.Duration(prm_scrapeint) * time.Second)
	}
}
