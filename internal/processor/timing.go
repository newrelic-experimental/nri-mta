package processor

import (
   "fmt"
   "github.com/newrelic/infra-integrations-sdk/v4/data/event"
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

func (T Timing) ToTraces() {
   if !arguments.Args.Traces {
      return
   }
   // spans := make([]*Span, 0, len(T.timings))
   parentId := ""
   var duration time.Duration = 0

   for i, t := range T.timings {
      if i == 0 {
         // No special case for the first index
      } else {
         duration = t.time.Sub(*T.timings[i].time)
      }
      id := T.id + ":" + strconv.Itoa(t.id)
      name := t.byHost
      if name == "" {
         name = t.byIP
      }
      RecordSpan(id, T.id+string(T.direction), *t.time, duration, name, parentId)
      // spans = append(spans, NewSpan(id, T.id, t.time.UnixMilli(), float64(duration), name, parentId))
      parentId = id
   }
   SendSpans()
}

func (T Timing) ToEvent(entity *integration.Entity, src string, dst string) {
   if !arguments.Args.Events {
      return
   }
   evt, err := event.New(time.Now(), "MTA response time", "MTA")
   if err != nil {
      log.Error(" error creating new Event: %v", err)
      return
   }

   evt.Attributes["direction"] = string(T.direction)
   evt.Attributes["messageId"] = T.id
   evt.Attributes["source"] = src
   evt.Attributes["destination"] = dst
   evt.Attributes["transitTime"] = T.timing()
   entity.AddEvent(evt)
}

func (T Timing) source() (src string) {
   if len(T.timings) < 1 {
      log.Warn("source: no timings available")
      return
   }

   // last from
   idx := len(T.timings) - 1
   src = T.timings[idx].fromHost
   if src == "" {
      src = T.timings[idx].fromIP
   }
   return
}

func (T Timing) destination() (dst string) {
   if len(T.timings) < 1 {
      log.Warn("destination: no timings available")
      return
   }

   // first by
   dst = T.timings[0].byHost
   if dst == "" {
      dst = T.timings[0].byIP
   }
   return
}

func (T Timing) timing() (delta int64) {
   if len(T.timings) < 2 {
      log.Warn("timing: only one timing available, returning 0")
      return 0
   }
   // first time - last time
   idx := len(T.timings) - 1
   return T.timings[idx].time.Sub(*T.timings[0].time).Milliseconds()
}

func (T Timing) ToMetrics(entity *integration.Entity) {
   if !arguments.Args.Metrics {
      return
   }

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
      m, err := metric.NewGauge(*e.time, "MTA", float64(delta))
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
