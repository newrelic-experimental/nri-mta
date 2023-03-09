package constants

type Direction string

const (
   SEND    Direction = "SEND"
   RECEIVE Direction = "RECEIVE"
)

type Kind string

const (
   IMAP    Kind = "IMAP"
   MSGRAPH Kind = "MSGRAPH"
)
