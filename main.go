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
	targetPort = flag.String("p", "5201", "(Optional) Server port to connect to")
)

type iperf3Collector struct {
	EndSumSentBitsPerSecond     *prometheus.Desc
	EndSumReceivedBitsPerSecond *prometheus.Desc
	EndSumSentBytes             *prometheus.Desc
	EndSumReceivedBytes         *prometheus.Desc
	EndSumSentRetransmits       *prometheus.Desc
	EndStreamsSenderMaxRtt      *prometheus.Desc
	EndStreamsSenderMinRtt      *prometheus.Desc
	EndStreamsSenderMeanRtt     *prometheus.Desc
	EndStreamsSenderMaxSndCwnd  *prometheus.Desc
	EndStreamsSenderMaxSndWnd   *prometheus.Desc
}

func (collector *iperf3Collector) Describe(ch chan<- *prometheus.Desc) {
	//Update this section with the each metric you create for a given collector
	ch <- collector.EndSumSentBitsPerSecond
	ch <- collector.EndSumReceivedBitsPerSecond
	ch <- collector.EndSumSentBytes
	ch <- collector.EndSumReceivedBytes
	ch <- collector.EndSumSentRetransmits
	ch <- collector.EndStreamsSenderMaxRtt
	ch <- collector.EndStreamsSenderMinRtt
	ch <- collector.EndStreamsSenderMeanRtt
	ch <- collector.EndStreamsSenderMaxSndCwnd
	ch <- collector.EndStreamsSenderMaxSndWnd
}

// Collect implements required collect function for all promehteus collectors
func (collector *iperf3Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/bin/iperf3", "-c", *targetHost, "-J", "-p", *targetPort).Output()
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
	EndSumSentBitsPerSecond := prometheus.MustNewConstMetric(collector.EndSumSentBitsPerSecond, prometheus.GaugeValue, stats.End.SumSent.BitsPerSecond)
	ch <- EndSumSentBitsPerSecond
	EndSumReceivedBitsPerSecond := prometheus.MustNewConstMetric(collector.EndSumReceivedBitsPerSecond, prometheus.GaugeValue, stats.End.SumReceived.BitsPerSecond)
	ch <- EndSumReceivedBitsPerSecond
	EndSumSentBytes := prometheus.MustNewConstMetric(collector.EndSumSentBytes, prometheus.GaugeValue, float64(stats.End.SumSent.Bytes))
	ch <- EndSumSentBytes
	EndSumReceivedBytes := prometheus.MustNewConstMetric(collector.EndSumReceivedBytes, prometheus.GaugeValue, float64(stats.End.SumReceived.Bytes))
	ch <- EndSumReceivedBytes
	EndSumSentRetransmits := prometheus.MustNewConstMetric(collector.EndSumSentRetransmits, prometheus.GaugeValue, float64(stats.End.SumSent.Retransmits))
	ch <- EndSumSentRetransmits
	EndStreamsSenderMaxRtt := prometheus.MustNewConstMetric(collector.EndStreamsSenderMaxRtt, prometheus.GaugeValue, float64(stats.End.Streams[0].Sender.MaxRtt))
	ch <- EndStreamsSenderMaxRtt
	EndStreamsSenderMinRtt := prometheus.MustNewConstMetric(collector.EndStreamsSenderMinRtt, prometheus.GaugeValue, float64(stats.End.Streams[0].Sender.MinRtt))
	ch <- EndStreamsSenderMinRtt
	EndStreamsSenderMeanRtt := prometheus.MustNewConstMetric(collector.EndStreamsSenderMeanRtt, prometheus.GaugeValue, float64(stats.End.Streams[0].Sender.MeanRtt))
	ch <- EndStreamsSenderMeanRtt
	EndStreamsSenderMaxSndCwnd := prometheus.MustNewConstMetric(collector.EndStreamsSenderMaxSndCwnd, prometheus.GaugeValue, float64(stats.End.Streams[0].Sender.MaxSndCwnd))
	ch <- EndStreamsSenderMaxSndCwnd
	EndStreamsSenderMaxSndWnd := prometheus.MustNewConstMetric(collector.EndStreamsSenderMaxSndWnd, prometheus.GaugeValue, float64(stats.End.Streams[0].Sender.MaxSndWnd))
	ch <- EndStreamsSenderMaxSndWnd
}

func newMetricsCollector() *iperf3Collector {
	return &iperf3Collector{
		EndSumSentBitsPerSecond: prometheus.NewDesc("iperf3_end_sum_sent_bits_per_second",
			"Shows TBD",
			nil, nil,
		),
		EndSumReceivedBitsPerSecond: prometheus.NewDesc("iperf3_end_sum_received_bits_per_second",
			"Shows TBD",
			nil, nil,
		),
		EndSumSentBytes: prometheus.NewDesc("iperf3_end_sum_sent_bytes",
			"Shows TBD",
			nil, nil,
		),
		EndSumReceivedBytes: prometheus.NewDesc("iperf3_end_sum_received_bytes",
			"Shows TBD",
			nil, nil,
		),
		EndSumSentRetransmits: prometheus.NewDesc("iperf3_end_sum_sent_retransmits",
			"Shows TBD",
			nil, nil,
		),
		EndStreamsSenderMaxRtt: prometheus.NewDesc("iperf3_end_streams_sender_max_rtt",
			"Shows TBD",
			nil, nil,
		),
		EndStreamsSenderMinRtt: prometheus.NewDesc("iperf3_end_streams_sender_min_rtt",
			"Shows TBD",
			nil, nil,
		),
		EndStreamsSenderMeanRtt: prometheus.NewDesc("iperf3_end_streams_sender_mean_rtt",
			"Shows TBD",
			nil, nil,
		),
		EndStreamsSenderMaxSndCwnd: prometheus.NewDesc("iperf3_end_streams_sender_max_snd_cwnd",
			"Shows TBD",
			nil, nil,
		),
		EndStreamsSenderMaxSndWnd: prometheus.NewDesc("iperf3_end_streams_sender_max_snd_wnd",
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
