package imap

import (
   "bufio"
   "context"
   "github.com/emersion/go-imap"
   "github.com/newrelic/infra-integrations-sdk/v4/log"
   "io"
   "nri-mta/internal/constants"
   "nri-mta/internal/errors"
   "nri-mta/internal/mta"
   "regexp"
   "strings"
)

func init() {
   mta.Register(constants.IMAP, NewIMAPAgent, CastIMAPAgent)
}

type Agent struct {
   SMTPHost string `yaml:"SMTPHost"`
   SMTPPort string `yaml:"SMTPPort"`
   UserName string `yaml:"UserName"`
   Password string `yaml:"Password"`
   IMAPUrl  string `yaml:"IMAPUrl,omitempty"`
}

func NewIMAPAgent() interface{} {
   return &Agent{}
}

func CastIMAPAgent(i interface{}) mta.MTAgent {
   s := i.(mta.MTAgent)
   return s
}

func (a *Agent) Send(ctx context.Context, direction constants.Direction, id int64, to string) (err error) {
   log.Debug("processor.send: enter")
   ch := make(chan error)
   go func() {
      ch <- func() error {
         return a.SendMsg(id, direction, to)
      }()
   }()
   select {
   case <-ctx.Done():
      log.Debug("processor.send: exit ctx.Done")
      return ctx.Err()
   case err = <-ch:
      log.Debug("processor.send: exit err = <-ch")
      return err
   }
}

func (a *Agent) Receive(ctx context.Context, direction constants.Direction, id int64) (headers []string, err error) {
   log.Debug("processor.receive: enter")
   type retVal struct {
      msg *imap.Message
      err error
   }
   ch := make(chan retVal)
   go func() {
      ch <- func() retVal {
         v := retVal{}

         // Wait for the send propagation delay
         messageNotFound := true
         for messageNotFound {
            v.msg, v.err = a.ReadMsg(id, direction)
            _, messageNotFound = v.err.(*errors.MessageNotFound)
         }
         return v
      }()
   }()
   select {
   case <-ctx.Done():
      // Context times-out
      log.Debug("processor.receive: exit ctx.Done")
      return nil, ctx.Err()
   case v := <-ch:
      if v.err != nil {
         log.Debug("processor.receive: exit err ")
         return nil, v.err
      }
      // Get the headers from the message
      headers = a.getHeaders(v.msg)
      // log.Debug("processor.receive: headers: %+v", headers)

      // Generate timings from the headers
      log.Debug("processor.receive: exit ch ")
      return
   }

}

// receive reads the sent message via imap

// getHeaders extracts the headers we're interested in (via config) from the IMAP message
func (a *Agent) getHeaders(msg *imap.Message) (headers []string) {
   //   log.Debug("getHeaders: msg.Items: %+v", msg.Items)
   //  if _, ok := msg.Items[imap.FetchRFC822]; ok && len(msg.Items) == 1 {
   if len(msg.Body) == 1 {
      for _, lit := range msg.Body {
         return headersToArray(lit)
      }
   }
   // }
   return []string{}
}

var continuation = regexp.MustCompile(`^\s.*`)
var whitespaces = regexp.MustCompile(`[[:space:]]+`)

// headersToArray massages the text returned via IMAP into a list of header strings
func headersToArray(r io.Reader) (value []string) {
   value = make([]string, 0, 50)
   scanner := bufio.NewScanner(r)
   lastLine := ""
   currentLine := ""
   for scanner.Scan() {
      currentLine = scanner.Text()
      // Mime boundary marker, we're done
      if strings.HasPrefix(currentLine, "--") {
         break
      }
      // Header continuation lines begin with whitespace, add to the current header
      if continuation.MatchString(currentLine) {
         lastLine += " " + strings.TrimSpace(currentLine)
         currentLine = ""
         continue
      }
      // We've found a header
      if lastLine != "" {
         value = append(value, lastLine)
         lastLine = ""
      }
      lastLine = currentLine
   }
   lastLine = whitespaces.ReplaceAllString(lastLine, " ")
   value = append(value, lastLine)
   return
}

func (a *Agent) Username() string {
   return a.UserName
}
