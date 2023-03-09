package imap

import (
   "crypto/tls"
   "errors"
   "fmt"
   "github.com/newrelic/infra-integrations-sdk/log"
   "net"
   "net/smtp"
   "nri-mta/internal/constants"
)

type loginAuth struct {
   username, password string
}

func LoginAuth(username, password string) smtp.Auth {
   return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
   return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
   if more {
      switch string(fromServer) {
      case "Username:":
         return []byte(a.username), nil
      case "Password:":
         return []byte(a.password), nil
      default:
         return nil, errors.New("Unknown from server")
      }
   }
   return nil, nil
}

func (a *Agent) SendMsg(id int64, direction constants.Direction, to string) (err error) {

   conn, err := net.Dial("tcp", a.SMTPHost+":"+a.SMTPPort)
   if err != nil {
      log.Error("net.Dial: err: %v", err)
      return
   }

   c, err := smtp.NewClient(conn, a.SMTPHost)
   if err != nil {
      log.Error("smtp.NewClient: err: %v", err)
      return
   }

   tlsconfig := &tls.Config{
      ServerName: a.SMTPHost,
   }

   if err = c.StartTLS(tlsconfig); err != nil {
      log.Error("c.StartTLS: err: %v", err)
      return
   }

   auth := LoginAuth(a.UserName, a.Password)

   if err = c.Auth(auth); err != nil {
      log.Error("c.Auth: err: %v", err)
      return
   }
   content := string(direction) + ":" + fmt.Sprint(id)
   msg := []byte(
      "To: " + to + "\r\n" +
         "Subject: " + content + "\r\n" +
         "\r\n" +
         content + ".\r\n")
   err = smtp.SendMail(a.SMTPHost+":"+a.SMTPPort, auth, a.UserName, []string{to}, msg)
   if err != nil {
      log.Error("smtp.SendMail: host: %s port: %s auth: %+v from: %s to: %s body: %s err: %v", a.SMTPHost, a.SMTPPort, auth, a.UserName, to, msg, err)
      return
   }
   return
}
