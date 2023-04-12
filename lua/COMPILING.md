Compiling luasockets for a new lua release:
1. get luarocks https://github.com/luarocks/luarocks/wiki/Download along with the lua distribution of your choice
2. Get the proper version of the lua lib - for the last release I used "lua-5.4.2_Win64_dll16_lib" at https://sourceforge.net/projects/luabinaries/files/5.4.2/Windows%20Libraries/Dynamic/ - it's likely that the important thing here is choosing the correct lua version and selecting "Dynamic". CRT version shouldn't matter so much.
3. configure luarocks as per their instructions
4. `luarocks install luasocket`
5. The dll you need will likely be at "%APPDATA%\luarocks\lib\lua\5.4\socket\core.dll" - rename it to socket-windows-$LUA_MAJOR-$LUA_MINOR.dll