#!/bin/bash

# ===QUICK RUN===
# ../go-system/start.sh
# sleep 10s
# python ../go-system/scripts/populate.py
# go run main.go -f="test" -t="read-write" -r=100 -n=5

# ===RUN MULTIPLE===
for n in 4 9 19
do
    ../go-system/start.sh $n
    sleep 10s
    python ../go-system/scripts/populate.py
    sleep 3s
    counter=0
    # for r in 1 2 3
    # for r in 100
    for r in 10 100 500
    do
        # for t in "read" "write" "read-write"
        # do
        t="read-write"
        for i in {1..3}
        do
            echo "Running $t test $i for $n nodes with request rate $r/s"
            go run main.go -f="results-N${n}_$(echo $t | tr '[:lower:]' '[:upper:]')_R$r-$i" -t=$t -r=$r -n=$n -c=$counter
            sleep 10s
            # sleep 5s
            let counter++
        done
        # done
    done
    kill $(ps | grep "go run cmd/main.go -id=" | awk '{print $1}')
    rm -r ../go-system/tmp
    sleep 10s
done
