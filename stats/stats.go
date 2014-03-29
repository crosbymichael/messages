package stats

import (
	"flag"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/influxdb"
	"log"
	"os"
	"time"
)

var (
	MessageCount metrics.Counter
)

var (
	toStdErr     bool
	duration     time.Duration
	influxConfig *influxdb.Config
)

func init() {
	registerFlags()

	MessageCount = metrics.NewCounter()
	metrics.Register("message-count", MessageCount)
}

func registerFlags() {
	flag.BoolVar(&toStdErr, "stats-to-stderr", false, "report stats periodically to stderr")
	flag.DurationVar(&duration, "stats-duration", 60*time.Second, "duration to flush stats")

	influxConfig = &influxdb.Config{}
	flag.StringVar(&influxConfig.Database, "stats-influx-db", "", "influxdb database name for metrics")
	flag.StringVar(&influxConfig.Username, "stats-influx-user", "", "influxdb username for metrics")
	flag.StringVar(&influxConfig.Password, "stats-influx-pass", "", "influxdb password for metrics")
}

func Collect() error {
	if toStdErr {
		go metrics.Log(metrics.DefaultRegistry, duration, log.New(os.Stderr, "stats: ", log.Lmicroseconds))
	}
	if influxConfig.Host != "" {
		go influxdb.Influxdb(metrics.DefaultRegistry, duration, influxConfig)
	}
	return nil
}
