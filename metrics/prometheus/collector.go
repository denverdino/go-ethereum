package prometheus

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/metrics"
)

var (
	countersKey        = []byte("counters")
	gauagesKey         = []byte("gauages")
	metersKey          = []byte("meters")
	histogramsKey      = []byte("histograms")
	timersKey          = []byte("timers")
	resettingTimersKey = []byte("resettingTimers")

	countersHeader       = []byte("# HELP counters is set of geth counters\n# TYPE counters gauage\n")
	gauagesHeader        = []byte("# HELP gauages is set of geth gauages\n# TYPE gauages gauage\n")
	meterHeader          = []byte("# HELP meters is set of geth meters\n# TYPE meters gauage\n")
	histogramHeader      = []byte("# HELP histograms is set of geth histograms\n# TYPE histograms summary\n")
	timerHeader          = []byte("# HELP timers is set of geth timers\n# TYPE timers summary\n")
	resettingTimerHeader = []byte("# HELP resetting_timers is set of geth resetting timers\n# TYPE resetting_timers summary\n")

	tagsValueTemplate = "{name=\"%s\",aggr=\"%s\"} %v\n"
)

var bufPool sync.Pool

func getBuf() *bytes.Buffer {
	buf := bufPool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

func giveBuf(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}

type collector struct {
	counters        *bytes.Buffer
	gauages         *bytes.Buffer
	histograms      *bytes.Buffer
	meters          *bytes.Buffer
	timers          *bytes.Buffer
	resettingTimers *bytes.Buffer
}

func newCollector() *collector {
	return &collector{
		counters:        getBuf(),
		gauages:         getBuf(),
		histograms:      getBuf(),
		meters:          getBuf(),
		timers:          getBuf(),
		resettingTimers: getBuf(),
	}
}

func (c *collector) reset() {
	giveBuf(c.counters)
	giveBuf(c.gauages)
	giveBuf(c.histograms)
	giveBuf(c.meters)
	giveBuf(c.timers)
	giveBuf(c.resettingTimers)
}

func (c *collector) result() *bytes.Buffer {
	buf := getBuf()
	if c.counters.Len() > 0 {
		buf.Write(countersHeader)
		buf.Write(c.counters.Bytes())
	}

	if c.gauages.Len() > 0 {
		buf.Write(gauagesHeader)
		buf.Write(c.gauages.Bytes())
	}

	if c.meters.Len() > 0 {
		buf.Write(meterHeader)
		buf.Write(c.meters.Bytes())
	}

	if c.histograms.Len() > 0 {
		buf.Write(histogramHeader)
		buf.Write(c.histograms.Bytes())
	}

	if c.timers.Len() > 0 {
		buf.Write(timerHeader)
		buf.Write(c.timers.Bytes())
	}

	if c.resettingTimers.Len() > 0 {
		buf.Write(resettingTimerHeader)
		buf.Write(c.resettingTimers.Bytes())
	}

	return buf
}

func (c *collector) addCounter(name string, m metrics.Counter) {
	writeMetric(c.counters, countersKey, name, "value", m.Count())
}

func (c *collector) addGuage(name string, m metrics.Gauge) {
	writeMetric(c.gauages, gauagesKey, name, "value", m.Value())
}

func (c *collector) addGuageFloat64(name string, m metrics.GaugeFloat64) {
	writeMetric(c.gauages, gauagesKey, name, "value", m.Value())
}

func (c *collector) addHistogram(name string, m metrics.Histogram) {
	ps := m.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
	writeMetric(c.histograms, histogramsKey, name, "count", m.Count())
	writeMetric(c.histograms, histogramsKey, name, "count", m.Count())
	writeMetric(c.histograms, histogramsKey, name, "max", m.Max())
	writeMetric(c.histograms, histogramsKey, name, "mean", m.Mean())
	writeMetric(c.histograms, histogramsKey, name, "min", m.Min())
	writeMetric(c.histograms, histogramsKey, name, "stddev", m.StdDev())
	writeMetric(c.histograms, histogramsKey, name, "variance", m.Variance())
	writeMetric(c.histograms, histogramsKey, name, "p50", ps[0])
	writeMetric(c.histograms, histogramsKey, name, "p75", ps[1])
	writeMetric(c.histograms, histogramsKey, name, "p95", ps[2])
	writeMetric(c.histograms, histogramsKey, name, "p99", ps[3])
	writeMetric(c.histograms, histogramsKey, name, "p999", ps[4])
	writeMetric(c.histograms, histogramsKey, name, "p9999", ps[5])
}

func (c *collector) addMeter(name string, m metrics.Meter) {
	writeMetric(c.meters, metersKey, name, "count", m.Count())
	writeMetric(c.meters, metersKey, name, "m1", m.Rate1())
	writeMetric(c.meters, metersKey, name, "m5", m.Rate5())
	writeMetric(c.meters, metersKey, name, "m15", m.Rate15())
	writeMetric(c.meters, metersKey, name, "mean", m.RateMean())
}

func (c *collector) addTimer(name string, m metrics.Timer) {
	ps := m.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
	writeMetric(c.timers, timersKey, name, "count", m.Count())
	writeMetric(c.timers, timersKey, name, "max", m.Max())
	writeMetric(c.timers, timersKey, name, "mean", m.Mean())
	writeMetric(c.timers, timersKey, name, "min", m.Min())
	writeMetric(c.timers, timersKey, name, "stddev", m.StdDev())
	writeMetric(c.timers, timersKey, name, "variance", m.Variance())
	writeMetric(c.timers, timersKey, name, "p50", ps[0])
	writeMetric(c.timers, timersKey, name, "p75", ps[1])
	writeMetric(c.timers, timersKey, name, "p95", ps[2])
	writeMetric(c.timers, timersKey, name, "p99", ps[3])
	writeMetric(c.timers, timersKey, name, "p999", ps[4])
	writeMetric(c.timers, timersKey, name, "p9999", ps[5])
	writeMetric(c.timers, timersKey, name, "m1", m.Rate1())
	writeMetric(c.timers, timersKey, name, "m5", m.Rate5())
	writeMetric(c.timers, timersKey, name, "m15", m.Rate15())
	writeMetric(c.timers, timersKey, name, "meanrate", m.RateMean())
}

func (c *collector) addResettingTimer(name string, m metrics.ResettingTimer) {
	if len(m.Values()) <= 0 {
		return
	}
	ps := m.Percentiles([]float64{50, 95, 99})
	val := m.Values()
	writeMetric(c.resettingTimers, resettingTimersKey, name, "count", len(val))
	writeMetric(c.resettingTimers, resettingTimersKey, name, "max", val[len(val)-1])
	writeMetric(c.resettingTimers, resettingTimersKey, name, "mean", m.Mean())
	writeMetric(c.resettingTimers, resettingTimersKey, name, "min", val[0])
	writeMetric(c.resettingTimers, resettingTimersKey, name, "p50", ps[0])
	writeMetric(c.resettingTimers, resettingTimersKey, name, "p95", ps[1])
	writeMetric(c.resettingTimers, resettingTimersKey, name, "p99", ps[2])

}

func writeMetric(buf *bytes.Buffer, key []byte, name, aggr string, value interface{}) {
	buf.Write(key)
	buf.WriteString(fmt.Sprintf(tagsValueTemplate, name, aggr, value))
}
