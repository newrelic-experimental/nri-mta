package processor

import (
   "fmt"
   "github.com/newrelic/infra-integrations-sdk/v4/data/metric"
   "github.com/newrelic/infra-integrations-sdk/v4/integration"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "nri-mta/internal/arguments"
   "nri-mta/internal/constants"
   "strconv"
   "time"
)

type Timing struct {
   timings   []*timing
   direction constants.Direction
   id        string
}

type timing struct {
   fromHost string
   fromIP   string
   byHost   string
   byIP     string
   time     *time.Time
   id       int
}

// TODO add an option for ToEvents

func (T Timing) ToMetrics(entity *integration.Entity) {

   log.Debug("ToMetrics: T.timings: %+v", T.timings)
   for i, e := range T.timings {
      var delta int64 = 0
      if i > 0 {
         if e.time == nil {
            log.Error("ToMetrics: timing e is missing time: %+v", e)
            continue
         }
         if T.timings[i-1].time == nil {
            log.Error("ToMetrics: timing [i-1] is missing time: %+v", T.timings[i-1])
            continue
         }
         delta = (e.time.UnixMilli() - T.timings[i-1].time.UnixMilli()) / 1000
      }
      m, err := metric.NewGauge(*e.time, "SMTP", float64(delta))
      if err != nil {
         log.Error("NewGauge: %v", err)
         continue
      }

      m.AddDimension("messageId", T.id)
      m.AddDimension("receivedBy", e.byHost)
      m.AddDimension("receivedFrom", e.fromHost)
      m.AddDimension("receivedAt", e.time.String())
      m.AddDimension("sequenceId", strconv.Itoa(e.id))
      m.AddDimension("direction", string(T.direction))
      entity.AddMetric(m)
   }
}

func newTimings(direction constants.Direction, headers []string) (*Timing, error) {
   times := &Timing{direction: direction}
   times.timings = make([]*timing, 0, len(headers))
   // Iterate backwards over the headers so the timings are in first -> last order
   for i, id := len(headers)-1, 0; i >= 0; i-- {
      h := headers[i]
      m := arguments.MatchHeader(h)
      // No match, not interesting
      if m == nil {
         continue
      }
      log.Debug("match: %+v\n", m)
      id++
      // FIXME if fromhost is nil infer it
      t := &timing{
         fromHost: m["fromhost"],
         fromIP:   m["fromip"],
         byHost:   m["byhost"],
         byIP:     m["byip"],
         time:     arguments.GetTime(m),
         id:       id,
      }
      times.timings = append(times.timings, t)
   }

   if len(times.timings) <= 0 {
      return nil, fmt.Errorf("headers do not contain any timings")
   }

   log.Debug("times.timings[0] %+v\n", times.timings[0])
   times.id = times.timings[0].fromHost + ":" + strconv.FormatInt(times.timings[0].time.UnixMilli(), 10)
   return times, nil
}
