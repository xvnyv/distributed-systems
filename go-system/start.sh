#!/bin/bash

# sorry shawn, only works for mac

dir=$(echo $PWD | sed 's! !\\\\ !g')
echo $dir
osascript -e "
    tell application \"Terminal\" 
        do script \"cd $dir && go run cmd/main.go -id=0 -port=8000 -first=true\"
        do script \"cd $dir && go run cmd/main.go -id=1 -port=8001\"
        do script \"cd $dir && go run cmd/main.go -id=2 -port=8002\"
        do script \"cd $dir && go run cmd/main.go -id=3 -port=8003\"
        do script \"cd $dir && go run cmd/main.go -id=4 -port=8004\"
    end tell
    "