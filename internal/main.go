package internal

import (
   "context"
   "fmt"
   "github.com/newrelic/infra-integrations-sdk/v4/integration"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "nri-mta/internal/arguments"
   "nri-mta/internal/errors"
   "nri-mta/internal/processor"
   "os"
   "runtime"
   "strings"
   "sync"
   "time"
)

const (
   integrationName = "com.newrelic.mta-monitor"
)

var (
   args               arguments.ArgumentList
   integrationVersion = "0.0.1"
   gitCommit          = ""
   buildDate          = ""
)

func Main() {
   log.SetupLogging(args.Verbose)
   // Create Integration
   i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
   exitOnErr(err, "integration.New")

   // e, err := i.NewEntity("", "", "")
   // exitOnErr(err)

   if args.ShowVersion {
      fmt.Printf(
         "New Relic %s integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
         strings.Replace(integrationName, "com.newrelic.", "", 1),
         integrationVersion,
         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
         runtime.Version(),
         gitCommit,
         buildDate)
      os.Exit(0)
   }

   err = args.Validate()
   exitOnErr(err, "args.Validate")

   processors, err := processor.GetProcessors(args.TraceConfig)
   if err != nil {
      exitOnErr(err, "processor.GetProcessors")
   }
   if len(processors) <= 0 {
      exitOnErr(&errors.NoValidTestsError{}, "processor.GetProcessor")
   }

   // FIXME tweak the timeout so we finish _before_ the process is killed
   ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.Timeout)*time.Second)
   defer cancel()

   var wg sync.WaitGroup
   for _, proc := range processors {
      // Run each test (ingress + egress) asynchronously
      wg.Add(2)
      go func(p *processor.Processor) {
         p.Process(ctx, &wg, i.HostEntity)
         log.Debug("p.Process complete wg: %+v", wg)
      }(proc)
   }

   // Hack so we can select on the waitGroup
   waitCh := make(chan struct{})
   go func() {
      wg.Wait()
      close(waitCh)
   }()

   // Wait for either the timeout or complete
   select {
   case <-waitCh:
   case <-ctx.Done():
      if ctx.Err() != nil {
         log.Error("context error: %v", ctx.Err())
      }
   }

   exitOnErr(i.Publish(), "i.Publish")
}

func exitOnErr(err error, prefix string) {
   if err != nil {
      log.Error("Encountered fatal at: %s error: %v", prefix, err)
      os.Exit(1)
   }
}
