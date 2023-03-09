package main

import (
   "nri-mta/internal"
   _ "nri-mta/internal/mta/imap"
   _ "nri-mta/internal/mta/msgraph"
)

func main() {
   internal.Main()
}
