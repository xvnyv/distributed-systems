#!/bin/bash
nginx -s stop
nginx -c "$PWD"/nginx.conf