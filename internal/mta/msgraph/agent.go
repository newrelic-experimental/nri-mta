package msgraph

import (
   "context"
   "nri-mta/internal/constants"
   "nri-mta/internal/mta"
)

func init() {
   mta.Register(constants.MSGRAPH, NewMSGraphAgent, CastMSGraphAgent)
}

type Agent struct {
   UserName string
}

func NewMSGraphAgent() interface{} {
   return &Agent{}
}

func CastMSGraphAgent(i interface{}) mta.MTAgent {
   s := i.(mta.MTAgent)
   return s
}

func (a *Agent) Send(ctx context.Context, direction constants.Direction, id int64, to string) (err error) {
   return
}
func (a *Agent) Receive(ctx context.Context, direction constants.Direction, id int64) (headers []string, err error) {
   return
}

func (a *Agent) Username() string {
   return a.UserName
}
