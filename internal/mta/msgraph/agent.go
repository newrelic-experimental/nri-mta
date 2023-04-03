package msgraph

import (
   "context"
   "fmt"
   "nri-mta/internal/constants"
   "nri-mta/internal/mta"
)

func init() {
   mta.Register(constants.MSGRAPH, NewMSGraphAgent, CastMSGraphAgent)
}

type Agent struct {
   ClientID     string   `yaml:"ClientID"`
   ClientSecret string   `yaml:"ClientSecret"`
   TenantID     string   `yaml:"TenantID"`
   Scopes       []string `yaml:"Scopes"`
   UserName     string   `yaml:"UserName"`
}

func NewMSGraphAgent() interface{} {
   return &Agent{}
}

func CastMSGraphAgent(i interface{}) mta.MTAgent {
   s := i.(mta.MTAgent)
   return s
}

func (a *Agent) Send(ctx context.Context, direction constants.Direction, id int64, to string) (err error) {
   gh, err := NewGraphHelper(a.ClientID, a.TenantID, a.ClientSecret)
   if err != nil {
      fmt.Println("Returning error from NewGraphHelp")
      return
   }
   err = gh.SendMessage(to)
   if err != nil {
      fmt.Println("Returning error from g.SendMessage")
   }
   return
}
func (a *Agent) Receive(ctx context.Context, direction constants.Direction, id int64) (headers []string, err error) {
   return
}

func (a *Agent) Username() string {
   return a.UserName
}

func (a *Agent) AgentHost() string {
   return "MSGraph"
}
