package processor

import (
   "context"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
   "nri-mta/internal/arguments"
   "time"
)

var harvester *telemetry.Harvester

func RecordSpan(id string, traceId string, timestamp time.Time, duration time.Duration, name string, parentId string) {
   if harvester == nil {
      var err error
      harvester, err = telemetry.NewHarvester(telemetry.ConfigAPIKey(arguments.Args.NewRelicIngestKey))
      if err != nil {
         log.Error("NewHarvester: %v", err)
      }

   }

   harvester.RecordSpan(telemetry.Span{
      Duration:    duration,
      ID:          id,
      Name:        name,
      ParentID:    parentId,
      ServiceName: "MTA",
      Timestamp:   timestamp,
      TraceID:     traceId,
      Attributes:  map[string]interface{}{},
   })
}

func SendSpans() {
   harvester.HarvestNow(context.Background())
}
