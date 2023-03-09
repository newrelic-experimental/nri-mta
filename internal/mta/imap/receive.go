package imap

import (
   "fmt"
   "github.com/emersion/go-imap"
   "github.com/emersion/go-imap/client"
   "github.com/newrelic/infra-integrations-sdk/log"
   "net/textproto"
   "nri-mta/internal/arguments"
   "nri-mta/internal/constants"
   "nri-mta/internal/errors"
   "os"
)

func (a *Agent) ReadMsg(id int64, direction constants.Direction) (msg *imap.Message, err error) {
   log.Debug("ReadMsg: enter from host: %s id: %d direction: %s", a.IMAPUrl, id, direction)
   c, err := client.DialTLS(a.IMAPUrl, nil)
   if err != nil {
      log.Error("%v", err)
      return
   }
   if arguments.Verbose {
      c.SetDebug(os.Stderr)
   }

   if err = c.Login(a.UserName, a.Password); err != nil {
      log.Error("%v", err)
      return
   }
   defer c.Logout()

   _, err = c.Select("INBOX", false)
   if err != nil {
      log.Error("%v", err)
      return
   }

   criteria := &imap.SearchCriteria{
      Header: textproto.MIMEHeader{"SUBJECT": {string(direction) + ":" + fmt.Sprint(id)}},
   }
   log.Debug("ReadMsg: search criteria: %+v", criteria)

   ids, err := c.Search(criteria)
   if err != nil {
      log.Error("%v", err)
      return
   }
   if len(ids) == 0 {
      err = &errors.MessageNotFound{Msg: fmt.Sprintf("ReadMsg: message not found %s:%d", direction, id)}
      log.Debug("c.Search err: %v", err)
      return
   }
   log.Debug("ReadMsg: search returned: %d", len(ids))

   seqset := new(imap.SeqSet)

   //   if len(ids) == 1 {
   seqset.AddNum(ids...)
   //   } else {
   //      seqset.AddNum(ids[0:10]...)
   //   }

   messages := make(chan *imap.Message, len(ids))
   // done := make(chan error, 1)
   // go func() {
   //   done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchBody, imap.FetchBodyStructure, imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchRFC822, imap.FetchRFC822Header, imap.FetchRFC822Size, imap.FetchRFC822Text, imap.FetchUid}, messages)
   // err = c.Fetch(seqset, []imap.FetchItem{imap.FetchBody, imap.FetchBodyStructure, imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchRFC822, imap.FetchRFC822Header, imap.FetchRFC822Size, imap.FetchRFC822Text, imap.FetchUid}, messages)
   err = c.Fetch(seqset, []imap.FetchItem{imap.FetchRFC822}, messages)
   // }()
   if err != nil {
      log.Error("ReadMsg: c.fetch: host: %s err: %v", a.IMAPUrl, err)
      return
   }

   if len(messages) <= 0 {
      log.Error("ReadMsg: no messages")
      return nil, fmt.Errorf("%s:%d not found", direction, id)
   }
   if len(messages) > 1 {
      log.Warn("Multiple (%d) %s:%d messages found", len(messages), direction, id)
   }

   for msg = range messages {
      // headerSection, _ := imap.ParseBodySectionName(imap.FetchRFC822Header)
      // msgHeader := msg.GetBody(headerSection)
      // if msgHeader == nil {
      //   fmt.Println("msg header retrieve failed..")
      // } // else {
      //    headerBody, _ := ioutil.ReadAll(msgHeader)
      //    fmt.Printf("headerBody %s \n", string(headerBody))
      // }
      //      log.Debug("Subject: " + msg.Envelope.Subject)
      // fmt.Printf("Body: %+v\n ", msg.Body)

      err = deleteMsg(c, msg.SeqNum)
      if err != nil {
         log.Error("Error deleting message: %v", err)
      }
   }

   // if err := <-done; err != nil {
   //    log.Fatal(err)
   // }

   //   log.Debug("ReadMsg: results: ids: %+v msg: %+v", ids, msg)
   return
}

func deleteMsg(c *client.Client, num uint32) (err error) {
   // TODO make deletion a configurable option
   seqset := new(imap.SeqSet)
   seqset.AddNum(num)
   // First mark the message as deleted
   item := imap.FormatFlagsOp(imap.AddFlags, true)
   flags := []interface{}{imap.DeletedFlag}
   if err = c.Store(seqset, item, flags, nil); err != nil {
      return
   }

   // Then expunge it
   err = c.Expunge(nil)
   return
}
