-- original file found in a GPLv3 code repository, unclear if this is the intended license nor who the authors are
-- SNI modifications by Berserker, jsd1982; modifications licensed under MIT License
-- version 3 changes Read response from JSON to HEX
-- version 4 introduces memory domain support for non-snes systems

if not event then
    is_snes9x = true
    memory.usememorydomain = function()
        -- snes9x always uses "System Bus" domain, which cannot be switched
    end
end

local readbytearray
local writebytearray

if is_snes9x then
    -- domain is ignored for snes9x
    readbytearray = function(domain, addr, length)
        local mtable = memory.readbyterange(addr, length);

        -- jsd: format output in 2-char hex per byte:
        local t = {};
        for i=1,length do
            t[#t+1] = string.format("%02x", mtable[i])
        end
        return t
    end
    writebytearray = function(domain, addr, values)
        for i, value in ipairs(values) do
            memory.writebyte(addr + i - 1, value)
        end
    end
else
    -- bizhawk
    readbytearray = function(domain, addr, length)
        local mtable;

        -- jsd: wrap around address by domain size:
        local domainsize = memory.getmemorydomainsize(domain)
        while addr >= domainsize do
            addr = addr - domainsize
        end

        mtable = memory.read_bytes_as_array(addr, length, domain)

        -- jsd: format output in 2-char hex per byte:
        local t = {};
        for i=0, length-1 do
            t[#t+1] = string.format("%02x", mtable[i])
        end
        return t
    end
    writebytearray = function(domain, addr, values)
        -- jsd: wrap around address by domain size:
        local domainsize = memory.getmemorydomainsize(domain)
        while addr >= domainsize do
            addr = addr - domainsize
        end

        memory.write_bytes_as_array(addr, values, domain)
    end
end

local socket = require("socket.core")

local connection
local host = os.getenv("SNI_LUABRIDGE_LISTEN_HOST") or '127.0.0.1'
local port = os.getenv("SNI_LUABRIDGE_LISTEN_PORT") or 35398
local connected = false
local expectedState = {}

local function getState()
    local state = {
        { "client_version", client.getversion() },
        { "platform", string.lower(emu.getsystemid()) },
        { "rom_name", gameinfo.getromname() },
        { "rom_hash", gameinfo.getromhash() },
    }
    return state
end

-- v4 protocol:
local function onMessageV4(s)
    local parts = {}
    for part in string.gmatch(s, '([^|]+)') do
        parts[#parts + 1] = part
    end

    local cmd = parts[1]
    local rsp = {cmd}

    if cmd == "domains" then
        rsp[#rsp+1] = "ok"
        -- bizhawk 2.8 pcall fails to protect against native methods throwing exceptions.
        local status, err = pcall(memory.getmemorydomainlist)
        if status ~= true then
            print(err)
            rsp[#rsp+1] = "0"
        else
            local domains = err
            rsp[#rsp+1] = string.format("%x", #domains+1)
            for i=0,#domains do
                local name = domains[i]
                local size
                status, err = pcall(memory.getmemorydomainsize, name)
                if status ~= true then
                    print(err)
                    size = 0
                else
                    size = err
                end
                rsp[#rsp+1] = name .. ";" .. string.format("%x", size)
            end
        end
    elseif cmd == "state" then
        rsp[#rsp+1] = "ok"
        local state = getState()
        rsp[#rsp+1] = #state
        for i=1,#state do
            rsp[#rsp+1] = state[i][1] .. ";" .. state[i][2]
        end
    elseif cmd == "expect" then
        rsp[#rsp+1] = "ok"
        -- set expected state and notify if changed:
        expectedState = getState()
        rsp[#rsp+1] = #expectedState
        for i=1,#expectedState do
            rsp[#rsp+1] = expectedState[i][1] .. ";" .. expectedState[i][2]
        end
    else
        rsp[#rsp+1] = "unknown"
    end

    connection:send(table.concat(rsp, "|") .. "\n")
end

-- v3 protocol for compatibility:
local function onMessageV3(s)
    local parts = {}
    for part in string.gmatch(s, '([^|]+)') do
        parts[#parts + 1] = part
    end
    if parts[1] == "Read" then
        -- Read|address|size|domain\n
        local adr = tonumber(parts[2])
        local length = tonumber(parts[3])
        local domain
        if is_snes9x ~= true then
            domain = parts[4]
        end

        local byteRange = readbyterange(domain, adr, length)
        connection:send(table.concat(byteRange) .. "\n")
    elseif parts[1] == "Write" then
        -- Write|address|domain|..data..\n
        local adr = tonumber(parts[2])
        local domain
        local offset = 2
        if is_snes9x ~= true then
            domain = parts[3]
            offset = 3
        end

        local values = {}
        for k = offset, #parts do
            values[#values+1] = tonumber(parts[k])
        end
        writebytearray(domain, adr, values)
    elseif parts[1] == "Message" then
        print(parts[2])
    elseif parts[1] == "Version" then
        if is_snes9x then
            connection:send("Version|SNI Connector|4|Snes9x\n")
        else
            connection:send("Version|SNI Connector|4|Bizhawk\n")
        end
    elseif parts[1] == "UpgradeV4" then
        print("Upgraded to V4 protocol")
        connection:send("UpgradeV4|upgraded\n")
        onMessage = onMessageV4
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

        pcall(memory.usememorydomain, "System Bus")

        -- default to v3 protocol:
        onMessage = onMessageV3

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
