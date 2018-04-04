package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tywkeene/go-fsevents"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	inputFile *os.File

	prm_listenAddr   string
	prm_inputFile    string
	prm_scrapeIntvl  int
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
	kingpin.Flag("inputfile", "exporter input file (truncated!)").Short('i').Default("/var/log/radicale/radicale_exporter_input.log").StringVar(&prm_inputFile)
	kingpin.Flag("scrapeinterval", "Prometheus scrape interval").Short('s').Default("15").IntVar(&prm_scrapeIntvl)
	kingpin.Flag("radicale", "address:port to contact Radicale on").Short('r').Default(":5232").StringVar(&prm_radicaleAddr)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()
}

func openInputFile() {
	var err error
	inputFile, err = os.OpenFile(prm_inputFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func inotifyListener() {
	options := &fsevents.WatcherOptions{
		Recursive: false,
	}

	w, err := fsevents.NewWatcher(path.Dir(prm_inputFile), fsevents.FileRemovedEvent, options)
	if err != nil {
		log.Fatal(err)
	}

	w.StartAll()
	go w.Watch()

	for {
		event := <-w.Events

		if event.IsFileRemoved() && event.Path == prm_inputFile {
			openInputFile()
		}
	}
}

func main() {
	parseFlags()
	serveMetrics()
	go inotifyListener()

	openInputFile()
	defer inputFile.Close()

	for {
		if checkTCP() {
			scanner := bufio.NewScanner(inputFile)

			for scanner.Scan() {
				inspectLine(scanner.Text())
			}

			inputFile.Truncate(0)
			inputFile.Seek(0, 0)
		}

		time.Sleep(time.Duration(prm_scrapeIntvl) * time.Second)
	}
}
