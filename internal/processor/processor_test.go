package processor

import (
   "reflect"
   "testing"
)

func Test_parseTraceConfig(t *testing.T) {
   tests := []struct {
      name    string
      args    string
      want    []*Processor
      wantErr bool
   }{
      {
         name:    "Mimimum passing",
         want:    nil,
         wantErr: false,
         args: `
Processors:
  - MTA:
      Kind: "IMAP"
    Client:
      Kind: "MSGRAPH"
`,
      },
   }
   for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
         got, err := parseTraceConfig([]byte(tt.args))
         if (err != nil) != tt.wantErr {
            t.Errorf("parseTraceConfig() error = %v, wantErr %v", err, tt.wantErr)
            return
         }
         if !reflect.DeepEqual(got, tt.want) {
            t.Errorf("parseTraceConfig() got = %v, want %v", got, tt.want)
         }
      })
   }
}
