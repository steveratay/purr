package main

import (
  "encoding/json"

  "github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
)

func Handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
  main()
  return nil, nil
}
