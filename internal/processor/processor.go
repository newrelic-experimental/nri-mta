package processor

import (
   "context"
   "github.com/newrelic/infra-integrations-sdk/v4/integration"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "gopkg.in/yaml.v3"
   "nri-mta/internal/constants"
   "nri-mta/internal/mta"
   "os"
   "sync"
   "time"
)

type Processors struct {
   Processors []*Processor `yaml:"Processors"`
}

type Processor struct {
   MTA    *mta.Agent `yaml:"MTA"`
   Client *mta.Agent `yaml:"Client"`
   id     int64
}

// Process happens in pairs, Egress (MT -> Test) and Ingress (Test -> MT) asynchronously (in a waitgroup)
func (p *Processor) Process(ctx context.Context, wg *sync.WaitGroup, entity *integration.Entity) {
   log.Debug("p.Process: enter wg: %+v", wg)
   p.id = time.Now().UnixNano()
   // Egress & Ingress run concurrently
   // Egress
   go func() {
      defer wg.Done()
      headers, err := p.measure(ctx, constants.SEND, p.MTA, p.Client)
      timing, err := newTimings(constants.SEND, headers)
      if err != nil {
         log.Error("p.process.measure.send: host: %v", p.MTA, err)
         return
      }
      timing.ToMetrics(entity)
   }()

   // Ingress
   go func() {
      defer wg.Done()
      headers, err := p.measure(ctx, constants.RECEIVE, p.Client, p.MTA)
      timing, err := newTimings(constants.RECEIVE, headers)
      if err != nil {
         log.Error("p.process.measure.receive: host: %v", p.MTA, err)
         return
      }
      timing.ToMetrics(entity)
   }()
   log.Debug("p.Process: exit wg: %+v", wg)
   return
}

func (p *Processor) measure(ctx context.Context, direction constants.Direction, mta *mta.Agent, client *mta.Agent) (headers []string, err error) {
   log.Debug("processor.measure: enter")
   err = mta.MTAgent.Send(ctx, direction, p.id, client.MTAgent.Username())
   // If we get any sort of error (timetout, comm, ...) then return, nothing we can do
   if err != nil {
      return
   }
   log.Debug("processor.measure: exit")

   return client.MTAgent.Receive(ctx, direction, p.id)
}

func GetProcessors(configFile string) ([]*Processor, error) {
   configYaml, err := os.ReadFile(configFile)
   if err != nil {
      return nil, err
   }

   return parseTraceConfig(configYaml)
}

func parseTraceConfig(configYaml []byte) ([]*Processor, error) {
   var procesors Processors
   err := yaml.Unmarshal(configYaml, &procesors)
   if err != nil {
      return nil, err
   }
   return procesors.Processors, err
}
