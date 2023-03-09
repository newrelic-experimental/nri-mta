package mta

import (
   "context"
   "fmt"
   "gopkg.in/yaml.v3"
   "nri-mta/internal/constants"
)

type Agent struct {
   Kind    constants.Kind `yaml:"Kind"`
   MTAgent MTAgent
}

type MTAgent interface {
   Send(ctx context.Context, direction constants.Direction, id int64, to string) (err error)
   Receive(ctx context.Context, direction constants.Direction, id int64) (headers []string, err error)
   Username() string
}

var constructors = make(map[constants.Kind]func() interface{}, 10)
var casts = make(map[constants.Kind]func(interface{}) MTAgent, 10)

func init() {
}

func Register(kind constants.Kind, constructor func() interface{}, cast func(interface{}) MTAgent) {
   constructors[kind] = constructor
   casts[kind] = cast
}

// TypeDef stops infinite recursion when trying to decode on Agent
type TypeDef struct {
   Kind constants.Kind `yaml:"Kind"`
}

// UnmarshalYAML gives us the hook into the YAML process
func (a *Agent) UnmarshalYAML(value *yaml.Node) (err error) {
   //  Decode into a temp struct to get the Kind
   td := &TypeDef{}
   err = value.Decode(td)
   if err != nil {
      return
   }

   if td.Kind == "" {
      err = fmt.Errorf("missing Kind")
      return
   }

   // Get the constructor for the Kind
   constructor, ok := constructors[td.Kind]
   if !ok {
      err = fmt.Errorf("missing constructor: %s", td.Kind)
      return
   }

   // Get the cast for the Kind
   cast, ok := casts[td.Kind]
   if !ok {
      err = fmt.Errorf("missing cast: %s", td.Kind)
      return
   }

   // Get an instance of the Kind
   tp := constructor()
   // Decode into the Kind's instance, this is an actual underlying struct
   err = value.Decode(tp)
   if err != nil {
      return
   }

   // CastIMAPAgent the decoded Kind instance to a Processor
   mp := cast(tp)
   a.MTAgent = mp
   return
}
