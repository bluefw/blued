#!/bin/bash

cd ..
go build

  (./blued agent -node=ds1)&
  (./blued agent -node=ds2 -bind=:7948 -rpc-addr=:7374 -rest-addr=:8342)&
