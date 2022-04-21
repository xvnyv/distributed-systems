#!/bin/bash

matcher='^[^.]*'

for f in results_bin2/*
do
    name=$(basename $f)
    html=results_html2/$([[ $name =~ $matcher ]] && echo ${BASH_REMATCH[0]}).html
    vegeta plot $f > $html
done   