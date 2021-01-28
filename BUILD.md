# Building this project

## First time setup of SDL2 on Windows

1. Install choco.
2. `choco install msys2`
3. Lunch msys2.exe. In my case it's in c:\tools\msys64\.
4. In the opened msys2 window type `pacman -Syu` to update the package database and core system packages.
5. If it asks you to close the window, close it (X button in the corner, not ctrl+C).
6. Open msys2.exe again and type `pacman -Syu` to update the rest of the packages.Y
7. Type `pacman -S mingw64/mingw-w64-x86_64-SDL2{,_image,_mixer,_ttf,_gfx}` to install sdl2 for 64 bit system. If for some reason packages are not found, type pacman -Ss sdl2 to see the correct package names.
8. Add `c:\tools\msys64\mingw64\bin\` to your PATH environment variable. It should contain gcc. If not, you will need to install it in msys2 with `pacman -S mingw64/mingw-w64-x86_64-gcc`.

At this point, it should build using
    `go build`
You can also build it using the static tag:
    `go build -tags static`
To build it as a Windows executable:
    `go build -tags static -ldflags -H=windowsgui`