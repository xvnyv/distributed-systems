#1. open powershell as administrator and run: "Set-ExecutionPolicy RemoteSigned"
#2. Run: "& filelocationofscript\widowsPowerShellScript.ps1"

Write-Host "Congratulations! Your first script executed successfully"

$numberOfNodes = 5

Start-Process .\nginx $PWD/nginx.conf

for($i = 1; $i -le $numberOfNodes; $i++){
    Start-Process powerShell "%{ cd $PWD/go-system/cmd; go run main.go -id=$i -port=800$i -first=true }"
}
