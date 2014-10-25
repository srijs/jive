package main

import (
  "io"
  "bytes"
  "log"
  "net/http"
  "encoding/json"
  "archive/tar"
  "compress/gzip"
  "compress/bzip2"
)

const (
  typeUnknown  = iota
  typeTar      = iota
  typeGzip     = iota
  typeBzip2    = iota
)

type Jive struct {

}

func (j *Jive) ServeHTTP(w http.ResponseWriter, r *http.Request) {

  if r.Method != "PUT" && r.Method != "POST" {
    http.NotFound(w, r)
    return
  }

  reader := r.Body
  encoder := json.NewEncoder(w)

  handleReader, typ, err := j.GetType(reader)

  if err != nil {
    w.WriteHeader(500)
    return
  }

  switch typ {
    case typeTar:   j.HandleTar(handleReader, encoder)
    case typeGzip:  j.HandleGzip(handleReader, encoder)
    case typeBzip2: j.HandleBzip2(handleReader, encoder)
    default:        w.WriteHeader(415)
  }

}

func (j *Jive) GetType(r io.Reader) (io.Reader, int, error) {

  buf := make([]byte, 264)

  rd, err := r.Read(buf)

  if err != nil {
    return nil, typeUnknown, err
  }

  newR := io.MultiReader(bytes.NewReader(buf), r)

  if rd > 1 && buf[0] == 0x1f && buf[1] == 0x8b {
    return newR, typeGzip, nil
  }

  if rd > 1 && buf[0] == 0x42 && buf[1] == 0x5a {
    return newR, typeBzip2, nil
  }

  if rd > 261 && buf[257] == 0x75 && buf[258] == 0x73 &&
     buf[259] == 0x74 && buf[260] == 0x61 && buf[261] == 0x72 {
    return newR, typeTar, nil
  }

  return newR, typeUnknown, nil

}


func (j *Jive) HandleTar(r io.Reader, encoder *json.Encoder) {

  tarR := tar.NewReader(r)

  for {
    header, err := tarR.Next()
    if err != nil {
      msg := err.Error()
      if (msg != "EOF") {
        log.Println(msg)
        encoder.Encode(msg)
      }
      break
    }
    err = encoder.Encode(header)
    if err != nil {
      log.Println(err.Error())
      break
    }
  }

}

func (j *Jive) HandleGzip(r io.Reader, encoder *json.Encoder) {

  gzipR, err := gzip.NewReader(r)
  if err != nil {
    log.Println(err.Error())
    return
  }

  defer gzipR.Close()

  j.HandleTar(gzipR, encoder)

}

func (j *Jive) HandleBzip2(r io.Reader, encoder *json.Encoder) {

  bzip2R := bzip2.NewReader(r)
  j.HandleTar(bzip2R, encoder)

}
