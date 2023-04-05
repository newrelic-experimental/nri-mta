package arguments

import (
   "fmt"
   sdkArgs "github.com/newrelic/infra-integrations-sdk/v4/args"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "gopkg.in/yaml.v3"
   "os"
   "regexp"
   "strings"
   "time"
)

var Args ArgumentList

type ArgumentList struct {
   sdkArgs.DefaultArgumentList
   Timeout               int    `default:"30" help:"The number of seconds to wait before a request times out."`
   CABundleFile          string `default:"" help:"Alternative Certificate Authority bundle file"`
   CABundleDir           string `default:"" help:"Alternative Certificate Authority bundle directory"`
   ShowVersion           bool   `default:"false" help:"Print build information and exit"`
   TLSInsecureSkipVerify bool   `default:"false" help:"Skip verification of the certificate sent by the host."`
   TraceConfig           string `default:"trace-config.yml" help:""`
   ReceivePatterns       string `default:"receive.patterns" help:""`
   Traces                bool   `default:"false" help:"Publish Trace data (Spans)"`
   NewRelicIngestKey     string `default:"" help:"New Relic Ingest Key"`
}

// Patterns all come from the config file
var Patterns = map[string]*regexp.Regexp{
   // "Date":     regexp.MustCompile(`^Date: (?P<DateTime>.*)`),
   // "Received": regexp.MustCompile(`^Received:( from (?P<fromhost>.*?)( \((?P<fromip>.*?)\))?)?( by (?P<byhost>.*?)( \((?P<byip>(.*?))?\))?)?( via (?P<via>.*?))?( with (?P<with>.*?))?( id (?P<id>.*?))?(; (?P<timestamp>.*))`),
   // "Received": regexp.MustCompile(`^Received:( from (?P<fromhost>.*?)( \((?P<fromip>.*?)\))?)?( by (?P<byhost>.*?)( \((?P<byip>(.*?))?\))?)?( via (?P<via>.*?))?( with (?P<with>.*?))?( id (?P<id>.*?))?( for (?P<for>.*?))?(; (?P<datetime>.*))`),
   // "Received": regexp.MustCompile(`^Received:( from (?P<from>.*?))?( by (?P<by>.*?))?( via (?P<via>.*?))?( with (?P<with>.*?))?( id (?P<id>.*?))?(; (?P<timestamp>.*))`),
   // "ReceivedBy":      regexp.MustCompile(`^Received: by (?P<Hostname>[0-9:].*) with SMTP id (?P<SMTPId>.*);(?P<DateTime>.*)`),
   // "ReceivedByEx":    regexp.MustCompile(`^Received: from (?P<Hostname>.*)? \((?P<HostIP>.*)\)? by (.*)? with (.*)? id (?P<SMTPId>.*);(?P<DateTime>.*)`),
   // "ReceivedFrom":    regexp.MustCompile(`^Received: from (?P<Hostname>.*) \(.*\) by (.*) with SMTPS id (?P<SMTPId>.*) for (.*);(?P<DateTime>.*)`),
   // "XReceivedGoogle": regexp.MustCompile(`^X-Received: by (?P<Hostname>[0-9:].*) with SMTP id (?P<SMTPId>.*)\.(.*)\.(?P<Timestamp>.*);(?P<DateTime>.*)`),
}

// TimeLayouts https://pkg.go.dev/time#pkg-constants
// TODO Time layouts should come from a config file
var TimeLayouts = []string{
   time.RFC1123Z,
   time.RFC1123,
   time.RFC822,
   time.RFC822Z,
   time.RFC850,
   time.RFC3339,
   time.RFC3339Nano,
   // 13 Feb 2023 19:41:09 -0800
   `_2 Jan 2006 15:04:05 -0700`,
   // Google: Mon, 06 Feb 2023 08:33:18 -0800 (PST)
   `Mon, _2 Jan 2006 15:04:05 -0700 (MST)`,
   // Tue, 28 Feb 2023 09:21:05 -0800 (PST)
   `Mon, 2 Jan 2006 15:04:05 -0700 (MST)`,
   // Exchange: Thu, 9 Feb 2023 14:54:17 +0000
   `Mon, _2 Jan 2006 15:04:05 -0700`,
   // Mon, 13 Feb 2023 07:50:45 -0800
   // Date: header Mon, 6 Feb 2023 10:32:42 -0600
   `Mon, _2 Jan 2006 15:04:05 -0700`,
}

var Verbose = false

// Validate validates the input arguments
func (args *ArgumentList) Validate() (err error) {
   Verbose = args.Verbose
   err = loadPatterns(args.ReceivePatterns)
   if err != nil {
      return
   }
   return
}

func loadPatterns(receivePatterns string) (err error) {
   patternYaml, err := os.ReadFile(receivePatterns)
   if err != nil {
      return
   }
   return parsePatterns(patternYaml)
}

func parsePatterns(patternYaml []byte) (err error) {
   patterns := make(map[string]string)
   err = yaml.Unmarshal(patternYaml, &patterns)
   if err != nil {
      log.Error("Error loading receive patterns: %v", err)
      log.Fatal(err)
      return
   }
   for k, v := range patterns {
      Patterns[k] = regexp.MustCompile(v)
   }
   return
}

// MatchHeader tests the header h against all patterns.
// Return the capture group as a map or nil if no match
func MatchHeader(h string) map[string]string {
   for _, p := range Patterns {
      if p.MatchString(h) {
         return namedCaptureToMap(p, h)
      }
   }
   return nil
}

// GetTime return a valid parsed Time or nil
func GetTime(m map[string]string) *time.Time {
   var err error
   // TODO there should be a smarter, type-safe, extensible way to do this
   value := m["timestamp"]
   if value == "" {
      value = m["datetime"]
   }

   if value == "" {
      log.Error("GetTime: time not found: %+v", m)
      return nil
   }

   value = strings.TrimSpace(value)
   // FIXME this doesn't work for Timestamp
   for _, layout := range TimeLayouts {
      t, e := time.Parse(layout, value)
      if e == nil {
         return &t
      }
      err = e
   }
   log.Error("GetTime: unable to parse: %s, error: %v", value, err)
   fmt.Printf("GetTime: unable to parse: %s\n", value)
   return nil
}

func namedCaptureToMap(p *regexp.Regexp, h string) map[string]string {
   log.Debug("namedCaptureToMap: enter: %s", h)
   result := make(map[string]string, 5)

   names := p.SubexpNames()
   // log.Debug("names: %+v\n", names)
   // dumpArray(names)

   capture := p.FindAllStringSubmatch(h, -1)[0]
   // log.Debug("capture: %+v\n", capture)
   // dumpArray(capture)

   for i, c := range capture {
      if names[i] == "" {
         continue
      }
      result[names[i]] = c
   }
   log.Debug("namedCaptureToMap: result: %+v", result)
   return result
}

func dumpArray(a []string) {
   for i, s := range a {
      if s == "" {
         continue
      }
      log.Debug("%d: %s\n", i, s)
   }
}
