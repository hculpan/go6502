@echo off

set FILE=%1
if [%FILE%]== [] goto err_msg

echo .
echo Building file %FILE%.a, output to %FILE%.bin
echo --------------------------------------------------------------------------------------
rem C:\Users\harry\Vasm\vasm6502_oldstyle_win32.exe %FILE%.a -dotdir -Fbin -o %FILE%.bin
C:\Users\harry\retroassembler\retroassembler -x %FILE%.a %FILE%.bin
echo --------------------------------------------------------------------------------------
goto end

:err_msg
echo Need to pass in the assembly file sans the extension

:end