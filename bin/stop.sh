#!/bin/bash

ps -ef | grep blued | grep -v grep | cut -c 10-15 | xargs kill -9
