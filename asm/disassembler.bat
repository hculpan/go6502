@echo off

set FILE=%1
set OUTFILE=%2
if [%FILE%]== [] goto err_msg
if [%OUTFILE%]== [] goto err_msg

echo .
echo Building file %FILE%.a, output to %FILE%.bin
echo --------------------------------------------------------------------------------------
rem C:\Users\harry\Vasm\vasm6502_oldstyle_win32.exe %FILE%.a -dotdir -Fbin -o %FILE%.bin
C:\Users\harry\retroassembler\retroassembler -d -D=0000 %FILE%.bin %OUTFILE%.lst
echo --------------------------------------------------------------------------------------
goto end

:err_msg
echo Need to pass in the bin file sans the extension and an output file name sans extension

:end