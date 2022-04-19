#!/bin/bash

# sorry shawn, only works for mac
n=4
if [ ! -z "$var" ]; then
    n=$1
fi

dir=$(echo $PWD | sed 's! !\\\\ !g')/../go-system
echo $dir
osascript -e "
    tell application \"Terminal\" 
        do script \"cd $dir && go run cmd/main.go -id=0 -port=8000 -first=true\"
    end tell
    "
sleep 1s
for i in $(seq 1 $n)
do
padI=`printf %03d $i`
echo $padI
osascript -e "
    tell application \"Terminal\" 
        do script \"cd $dir && go run cmd/main.go -id=$i -port=8$padI\"
    end tell
    "
sleep 2s
done

#  do script \"cd $dir && go run cmd/main.go -id=1 -port=8001\"
# do script \"cd $dir && go run cmd/main.go -id=2 -port=8002\"
# do script \"cd $dir && go run cmd/main.go -id=3 -port=8003\"
# do script \"cd $dir && go run cmd/main.go -id=4 -port=8004\"