#!/bin/bash

matcher='^[^.]*'

for f in results_bin/*
do
    name=$(basename $f)
    html=results_html/$([[ $name =~ $matcher ]] && echo ${BASH_REMATCH[0]}).html
    echo $html
    vegeta plot $f > $html
done   