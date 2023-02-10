-- original file found in a GPLv3 code repository, unclear if this is the intended license nor who the authors are
-- SNI modifications by Berserker, jsd1982; modifications licensed under MIT License
-- version 3 changes Read response from JSON to HEX

function readbyterange(addr, length, domain)
    local mtable;
    local mstart = 0;
    local mend = length - 1;

    -- jsd: format output in 2-char hex per byte:
    local toret = {};
    for i=mstart, mend do
        table.insert(toret, string.format("%02x", emu.read(addr + i, emu.memType.cpuDebug)))
    end
    return toret
end
function writebyte(addr, value, domain)
  emu.write(addr, value, emu.memType.cpuDebug)
end

local socket = require("socket.core")

local connection
local host = '127.0.0.1'
local port = 35398
local connected = false
local stopped = false
local name = "Unnamed"

local function onMessage(s)
    local parts = {}
    for part in string.gmatch(s, '([^|]+)') do
        parts[#parts + 1] = part
    end
    if parts[1] == "Read" then
        local adr = tonumber(parts[2])
        local length = tonumber(parts[3])
        local domain
        local byteRange = readbyterange(adr, length, domain)
        connection:send(table.concat(byteRange) .. "\n")
    elseif parts[1] == "Write" then
        local adr = tonumber(parts[2])
        local domain
        local offset = 2
        for k, v in pairs(parts) do
            if k > offset then
                writebyte(adr + k - offset - 1, tonumber(v), domain)
            end
        end
    elseif parts[1] == "SetName" then
        name = parts[2]
        emu.log("My name is " .. name .. "!")
    elseif parts[1] == "Message" then
        emu.log(parts[2])
    elseif parts[1] == "Exit" then
        emu.log("Lua script stopped, to restart the script press \"Run\"")
        stopped = true
    elseif parts[1] == "Version" then
        connection:send("Version|SNI Connector|3|Mesen-S\n")
    end
end


local main = function()
    if stopped then
        return nil
    end

    if not connected then
        emu.log('Connecting to SNI at ' .. host .. ':' .. port)
        connection, err = socket:tcp()
        if err ~= nil then
            emu.log(err)
            return
        end

        local returnCode, errorMessage = connection:connect(host, port)
        if (returnCode == nil) then
            emu.log("Error while connecting: " .. errorMessage)
            stopped = true
            connected = false
            emu.log("Please press \"Run\" to try to reconnect to SNI, make sure it's running.")
            return
        end

        connection:settimeout(0)
        connected = true
        emu.log('Connected to SNI')
        return
    end
    local s, status = connection:receive('*l')
    if s then
        onMessage(s)
    end
    if status == 'closed' then
        emu.log('Connection to SNI is closed')
        connection:close()
        connected = false
        return
    end
end

emu.addEventCallback(main, emu.eventType.startFrame)
