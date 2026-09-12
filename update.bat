@echo off
chcp 65001 >nul
setlocal
set "APP=%~dp0xhhRobot.exe"

if not exist "%APP%" (
  echo 未找到 xhhRobot.exe，请确认 update.bat 与程序放在同一目录。
  pause
  exit /b 1
)

"%APP%" -mode update
if errorlevel 1 (
  echo.
  echo 更新失败，请查看上方错误信息。
  pause
  exit /b 1
)

exit /b 0
