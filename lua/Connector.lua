-- original file found in a GPLv3 code repository, unclear if this is the intended license nor who the authors are
-- SNI modifications by Berserker, jsd1982; modifications licensed under MIT License
-- version 3 changes Read response from JSON to HEX

if not event then
    is_snes9x = true
    memory.usememorydomain = function()
        -- snes9x always uses "System Bus" domain, which cannot be switched
    end
else
    if emu.getsystemid() ~= "SNES" then
        print("Connector only for BSNES Core within Bizhawk, sorry.")
    end
    local current_engine = nil;
    -- client.get_lua_engine is new
    if client.get_lua_engine ~= nil then
        current_engine = client.get_lua_engine();
    -- emu.getluacore for old BizHawk
    elseif emu.getluacore ~= nil then
        current_engine = emu.getluacore();
    end
    if current_engine ~= nil and current_engine ~= "LuaInterface" then
        print("Wrong Lua Core. Found " .. current_engine .. ", was expecting LuaInterface. ")
        print("Please go to Config -> Customize -> Advanced and select Lua+LuaInterface.")
        print("Once set, restart Bizhawk.")
    end
end

function readbyterange(addr, length, domain)
    local mtable;
    local mstart = 0;
    local mend = length - 1;
    if is_snes9x then
        mtable = memory.readbyterange(addr, length);
        mstart = 1
        mend = length
    else
        -- jsd: wrap around address by domain size:
        local domainsize = memory.getmemorydomainsize(domain)
        while addr >= domainsize do
            addr = addr - domainsize
        end
        mtable = memory.readbyterange(addr, length, domain)
        mstart = 0;
        mend = length - 1;
    end

    -- jsd: format output in 2-char hex per byte:
    local toret = {};
    for i=mstart, mend do
        table.insert(toret, string.format("%02x", mtable[i]))
    end
    return toret
end
function writebyte(addr, value, domain)
  if is_snes9x then
    memory.writebyte(addr, value)
  else
    -- jsd: wrap around address by domain size:
    local domainsize = memory.getmemorydomainsize(domain)
    while addr >= domainsize do
        addr = addr - domainsize
    end
    memory.writebyte(addr, value, domain)
  end
end

local socket = require("socket.core")

local connection
local host = os.getenv("SNI_LUABRIDGE_LISTEN_HOST") or '127.0.0.1'
local port = os.getenv("SNI_LUABRIDGE_LISTEN_PORT") or 65398
local connected = false
local name = "Unnamed"

memory.usememorydomain("System Bus")

local function onMessage(s)
    local parts = {}
    for part in string.gmatch(s, '([^|]+)') do
        parts[#parts + 1] = part
    end
    if parts[1] == "Read" then
        local adr = tonumber(parts[2])
        local length = tonumber(parts[3])
        local domain
        if is_snes9x ~= true then
          domain = parts[4]
        end
        local byteRange = readbyterange(adr, length, domain)
        connection:send(table.concat(byteRange) .. "\n")
    elseif parts[1] == "Write" then
        local adr = tonumber(parts[2])
        local domain
        local offset = 2
        if is_snes9x ~= true then
          domain = parts[3]
          offset = 3
        end
        for k, v in pairs(parts) do
            if k > offset then
                writebyte(adr + k - offset - 1, tonumber(v), domain)
            end
        end
    elseif parts[1] == "SetName" then
        name = parts[2]
        print("My name is " .. name .. "!")
    elseif parts[1] == "Message" then
        print(parts[2])
    elseif parts[1] == "Version" then
        if is_snes9x then
            connection:send("Version|SNI Connector|3|Snes9x\n")
        else
            connection:send("Version|SNI Connector|3|Bizhawk\n")
        end
    elseif is_snes9x ~= true then
        if parts[1] == "Reset" then
            print("Rebooting core...")
            client.reboot_core()
        elseif parts[1] == "Pause" then
            print("Pausing...")
            client.pause()
        elseif parts[1] == "Unpause" then
            print("Unpausing...")
            client.unpause()
        elseif parts[1] == "PauseToggle" then
            print("Toggling pause...")
            client.togglepause()
        end
    end
end

local connectionBackOff = 0
local localIP, localPort, localFam

local function doclose()
    if connection ~= nil then
        print('Closing connection to SNI from ' .. localIP .. ':' .. localPort .. ' (' .. localFam .. ')')
        connection:close()
        print('Connection to SNI is closed from ' .. localIP .. ':' .. localPort .. ' (' .. localFam .. ')')
        localIP, localPort, localFam = nil, nil, nil
        connection = nil
    end
    connected = false
end

local function main()
    if not connected then
        if connectionBackOff > 0 then
            connectionBackOff = connectionBackOff - 1
            return
        end
        connectionBackOff = 60 * 10

        local err
        print('Connecting to SNI at ' .. host .. ':' .. port .. ' ...')
        connection, err = socket:tcp()
        if err ~= nil then
            print(err)
            print('Waiting 10 seconds...')
            return
        end

        connection:setoption('keepalive', true)
        connection:setoption('tcp-nodelay', true)
        connection:settimeout(0.01)
        local returnCode, err = connection:connect(host, port)
        if err ~= nil then
            print("Error while connecting: " .. err)
            print('Waiting 10 seconds...')
            connected = false
            return
        end

        connected = true

        localIP, localPort, localFam = connection:getsockname()
        if localFam == nil then
            localFam = 'inet'
        end
        print('Connected to SNI from ' .. localIP .. ':' .. localPort .. ' (' .. localFam .. ')')
        connection:settimeout(0)
        return
    end

    local s, status = connection:receive('*l')
    if s then
        onMessage(s)
    end
    if status == 'closed' then
        print('SNI closed the connection')
        doclose()
        print('Waiting 10 seconds...')
        return
    end
end

local function onexit()
    doclose()
end

if is_snes9x then
    -- snes9x-rr:
    emu.registerexit(onexit)
    emu.registerbefore(main)
else
    -- bizhawk:
    event.onexit(onexit)
    while true do
        main()
        emu.frameadvance()
    end
end
