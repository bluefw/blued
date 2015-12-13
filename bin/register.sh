#!/bin/bash

curl -H "Content-Type: application/json" -X PUT -d \
    '{"addr":"http://127.0.0.1:80/rs","providers":["a.b","a.c"],"consumers":["a.b"]}' \
    http://127.0.0.1:8341/msd/register
