package main

import "net/http"

func main() {
  j := &Jive{}
  http.ListenAndServe(":2000", j)
}
