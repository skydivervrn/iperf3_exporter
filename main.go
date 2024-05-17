package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var (
	targetHost = flag.String("target", "", "(Required) Server address to connect to")
)

type iperf3Collector struct {
	intervalsStreamsBitsPerSecond *prometheus.Desc
	intervalsStreamsRetransmits   *prometheus.Desc
	intervalsStreamsRtt           *prometheus.Desc
	intervalsStreamsRttvar        *prometheus.Desc
	intervalsStreamsPmtu          *prometheus.Desc
	endSumSentSeconds             *prometheus.Desc
	endSumSentBytes               *prometheus.Desc
}

func newMetricsCollector() *iperf3Collector {
	return &iperf3Collector{
		intervalsStreamsBitsPerSecond: prometheus.NewDesc("iperf3_intervals_streams_bits_per_second",
			"Shows TBD",
			nil, nil,
		),
		intervalsStreamsRetransmits: prometheus.NewDesc("iperf3_intervals_streams_retransmits",
			"Shows TBD",
			nil, nil,
		),
		intervalsStreamsRtt: prometheus.NewDesc("iperf3_intervals_streams_rtt",
			"Shows TBD",
			nil, nil,
		),
		intervalsStreamsRttvar: prometheus.NewDesc("iperf3_intervals_streams_rttvar",
			"Shows TBD",
			nil, nil,
		),
		intervalsStreamsPmtu: prometheus.NewDesc("iperf3_intervals_streams_pmtu",
			"Shows TBD",
			nil, nil,
		),
		endSumSentSeconds: prometheus.NewDesc("iperf3_end_sum_sent_seconds",
			"Shows TBD",
			nil, nil,
		),
		endSumSentBytes: prometheus.NewDesc("iperf3_end_sum_sent_bytes",
			"Shows TBD",
			nil, nil,
		),
	}
}

func main() {
	log.Println("Program started")

	flag.Parse()
	collector := newMetricsCollector()
	prometheus.MustRegister(collector)
	prometheus.Unregister(collectors.NewGoCollector())                                       //https://stackoverflow.com/questions/35117993/how-to-disable-go-collector-metrics-in-prometheus-client-golang
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})) //https://stackoverflow.com/questions/35117993/how-to-disable-go-collector-metrics-in-prometheus-client-golang
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(err)
	}
}

func (collector *iperf3Collector) Describe(ch chan<- *prometheus.Desc) {
	//Update this section with the each metric you create for a given collector
	ch <- collector.intervalsStreamsBitsPerSecond
	ch <- collector.intervalsStreamsRetransmits
	ch <- collector.intervalsStreamsRtt
	ch <- collector.intervalsStreamsRttvar
	ch <- collector.intervalsStreamsPmtu
	ch <- collector.endSumSentSeconds
	ch <- collector.endSumSentBytes
}

// Collect implements required collect function for all promehteus collectors
func (collector *iperf3Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/bin/iperf3", "-c", *targetHost, "-J").Output()
	if err != nil {
		log.Printf("Failed to run iperf3: %s", err)
		return
	}
	//log.Println(string(out))
	log.Println("Collecting metrics")
	stats := iperf3Result{}
	if err := json.Unmarshal(out, &stats); err != nil {
		log.Printf("Failed to parse iperf3 result: %s", err)
		return
	}
	intervalsStreamsBitsPerSecond := prometheus.MustNewConstMetric(collector.intervalsStreamsBitsPerSecond, prometheus.GaugeValue, stats.Intervals[0].Streams[0].BitsPerSecond)
	ch <- intervalsStreamsBitsPerSecond
	intervalsStreamsRetransmits := prometheus.MustNewConstMetric(collector.intervalsStreamsRetransmits, prometheus.GaugeValue, float64(stats.Intervals[0].Streams[0].Retransmits))
	ch <- intervalsStreamsRetransmits
	intervalsStreamsRtt := prometheus.MustNewConstMetric(collector.intervalsStreamsRtt, prometheus.GaugeValue, float64(stats.Intervals[0].Streams[0].Rtt))
	ch <- intervalsStreamsRtt
	intervalsStreamsRttvar := prometheus.MustNewConstMetric(collector.intervalsStreamsRttvar, prometheus.GaugeValue, float64(stats.Intervals[0].Streams[0].Rttvar))
	ch <- intervalsStreamsRttvar
	intervalsStreamsPmtu := prometheus.MustNewConstMetric(collector.intervalsStreamsPmtu, prometheus.GaugeValue, float64(stats.Intervals[0].Streams[0].Pmtu))
	ch <- intervalsStreamsPmtu
	endSumSentSeconds := prometheus.MustNewConstMetric(collector.endSumSentSeconds, prometheus.GaugeValue, stats.End.SumSent.Seconds)
	ch <- endSumSentSeconds
	endSumSentBytes := prometheus.MustNewConstMetric(collector.endSumSentBytes, prometheus.GaugeValue, stats.End.SumSent.BitsPerSecond)
	ch <- endSumSentBytes
}
