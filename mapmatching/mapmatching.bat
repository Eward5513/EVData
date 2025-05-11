@echo off
setlocal enabledelayedexpansion

set loops=20
if not "%1"=="" set loops=%1

set delay=10
if not "%2"=="" set delay=%2

echo Starting mapmatching.exe, total: %loops% times, interval: %delay% seconds
echo StartTime: %time%

for /l %%i in (0,1,%loops%) do (
    echo.
    echo ======================================
    echo Execution %%i of %loops%
    echo Time: %time%
    echo ======================================

    mapmatching.exe -d=40 -wf=1 -gcf=2000 -bi=%%i -bs=500 -p

    if %%i neq %loops%-1 (
        echo Waiting %delay% seconds...
        timeout /t %delay% /nobreak > nul
    )
)

echo.
echo All executions completed!
echo Total executions: %loops%
echo EndTime: %time%

pause