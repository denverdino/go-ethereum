package prometheus

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

// Run prometheus http server, returns metrics for any addr path
func Run(r metrics.Registry, addr string) {
	s := http.Server{
		Addr:         addr,
		Handler:      handler(r),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Info("Starting prometheus http server", "addr", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Warn("Unable to start prometheus metrics server", "addr", addr, "err", err)
	}
}

func handler(reg metrics.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := newCollector()
		defer c.reset()

		reg.Each(func(name string, i interface{}) {
			switch m := i.(type) {
			case metrics.Counter:
				ms := m.Snapshot()
				c.addCounter(name, ms)
			case metrics.Gauge:
				ms := m.Snapshot()
				c.addGuage(name, ms)
			case metrics.GaugeFloat64:
				ms := m.Snapshot()
				c.addGuageFloat64(name, ms)
			case metrics.Histogram:
				ms := m.Snapshot()
				c.addHistogram(name, ms)
			case metrics.Meter:
				ms := m.Snapshot()
				c.addMeter(name, ms)
			case metrics.Timer:
				ms := m.Snapshot()
				c.addTimer(name, ms)
			case metrics.ResettingTimer:
				ms := m.Snapshot()
				c.addResettingTimer(name, ms)
			}
		})

		res := c.result()
		defer giveBuf(res)

		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Content-Length", fmt.Sprint(res.Len()))
		w.Write(res.Bytes())
	})
}
