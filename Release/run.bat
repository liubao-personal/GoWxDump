@echo off
start /wait "" "GoWxDump.exe" -spy

cd decrypted 
start "" "my-app.exe"
