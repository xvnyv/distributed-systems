#!/bin/bash 

# # ===QUICK RUN===
# ../go-system/start.sh
# sleep 10s
# python ../go-system/scripts/populate.py
# go run main.go -f="test" -t="read" -r=100 -n=5

# ===RUN MULTIPLE===
for n in 4 9 19
do
    ../go-system/start.sh $n
    sleep 10s
    python ../go-system/scripts/populate.py
    sleep 3s
    counter=0
    for r in 10 100 500
    do
        for t in "read" "write" "read-write"
        do
            # t="read-write"
            for i in {1..3}
            do
                echo "Running $t test $i for $n nodes with request rate $r/s"
                go run main.go -f="results-N${n}_$(echo $t | tr '[:lower:]' '[:upper:]')_R$r-$i" -t=$t -r=$r -n=$n -c=$counter
                sleep 10s
                let counter++
            done
        done
    done
    sleep 10s
    rm -r ../go-system/tmp
    # this kill will return 1 error bcos it will try to kill the grep process as well, but that's not an issue
    kill $(ps | grep "go run cmd/main.go -id=" | awk '{print $1}') # this does not free up the ports bcos ListenAndServe spins up a new process
    for i in $(seq 0 $n)
    do
        padI=`printf %03d $i`
        kill $(lsof -t -i:8$padI)   # kill ListenAndServe process which is occupying port
        sleep 1s
    done
    sleep 60s
done
