-- snirp - a simple s-expression RPC server for SNI
-- Version 20230128
--
-- Copyright 2023 jsd1982
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
-- documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
-- rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
-- permit persons to whom the Software is furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
-- Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
-- WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
-- OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
-- OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

local socket = require("socket.core")
local brass = require("brass")
local state = "accept"
sni = {
    server = nil,
    port = 0,
    client = nil
}

-- handle a network request:
function handle(req_headers, req_body)
    -- copy request headers to response headers:
    local rsp_headers = {}
    for k,v in pairs(req_headers) do
        rsp_headers[k] = v
    end
    rsp_headers["frame_count"] = emu.framecount()

    -- validate expectations:
    local if_platform = req_headers["if_platform"]
    if if_platform ~= nil then
        local actual_platform = string.lower(emu.getsystemid())
        if if_platform ~= actual_platform then
            rsp_headers["error"] = "if_platform does not match actual `" .. actual_platform .. "`"
            return rsp_headers, nil
        end
    end
    local if_client_version = req_headers["if_client_version"]
    if if_client_version ~= nil then
        local actual_client_version = string.lower(client.getversion())
        if if_client_version ~= actual_client_version then
            rsp_headers["error"] = "if_client_version does not match actual `" .. actual_client_version .. "`"
            return rsp_headers, nil
        end
    end
    local if_rom_hash = req_headers["if_rom_hash"]
    if if_rom_hash ~= nil then
        local actual_rom_hash = gameinfo.getromhash()
        if if_rom_hash ~= actual_rom_hash then
            rsp_headers["error"] = "if_rom_hash does not match actual `" .. actual_rom_hash .. "`"
            return rsp_headers, nil
        end
    end

    -- process command:
    local cmd = req_headers["cmd"]
    if cmd == "info" then
        -- collect all current info about emulator, core, and game:

        -- get memory domain names and their sizes:
        local domainlist = memory.getmemorydomainlist()
        local domain_names = {}
        for i = 0,#domainlist do
            domain_names[#domain_names+1] = domainlist[i]
        end
        local domain_sizes = {}
        for i = 0,#domainlist do
            domain_sizes[#domain_sizes+1] = memory.getmemorydomainsize(domainlist[i])
        end

        -- response:
        return rsp_headers, {
            client_version = client.getversion(),
            platform = string.lower(emu.getsystemid()),
            rom_name = gameinfo.getromname(),
            rom_hash = gameinfo.getromhash(),
            domain_names = domain_names,
            domain_sizes = domain_sizes
        }
    elseif cmd == "read" then
        -- read memory:

        -- validate args:
        local domain, offset, size
        local domain = req_headers["domain"]
        local offset = req_headers["offset"]
        local size = req_headers["size"]
        if domain == nil or offset == nil or size == nil then
            rsp_headers["error"] = "missing required headers domain, offset, size"
            return rsp_headers, nil
        end
        offset = tonumber(offset, 16)
        size = tonumber(size, 16)

        -- read memory:
        local ok, data = pcall(memory.read_bytes_as_array, offset, size, domain)
        if not ok then
            rsp_headers["error"] = data
            return rsp_headers, nil
        end

        -- format data as hex:
        local sb = {}
        for i,v in ipairs(data) do
            sb[#sb+1] = string.format("%02x", v)
        end
        return rsp_headers, {
            data = table.concat(sb)
        }
    end

    rsp_headers["error"] = "unknown command"
    return req_headers, nil
end

function receive()
    if sni.client == nil then
        state = "accept"
        return
    end

    -- receive a message from the sni.client:
    local l, err = sni.client:receive('*l')
    if err == 'timeout' then
        return false
    elseif err == 'closed' then
        print("client:receive: Connection was closed")
        state = "accept"
        sni.client = nil
        return false
    elseif err ~= nil then
        print("client:receive: err=" .. err)
        state = "accept"
        sni.client = nil
        return false
    end

    if l == nil then
        print("client:receive: nil")
        return false
    end

    print("client:receive: `" .. l .. "`")

    local req, me
    me = 1
    while me <= #l do
        local ms = me
        -- (v20230128/req (seq 1) (check (platform snes) (rom-hash 00112233)) (read (snes/mem/console/wram ($10 $20)) (snes/mem/cart/sram (0 $300)))
        req, me, err = brass.decode(l:sub(ms))
        print("client:handle `" .. l:sub(ms,me) .. "`")
        if err ~= nil then
            print(err.err)
            return false
        end



        local req_headers, req_body = req[1], req[2]

        -- handle request:
        local rsp_headers, rsp_body = handle(req_headers, req_body)

        -- format response message:
        local rsp = brass.encode({ rsp_headers, rsp_body })

        --print("response: `" .. rsp .. "`")
        sni.client:send(rsp .. "\n")
    end
    return true
end

function main()
    local sock

    if sni.server == nil then
        local res, err = nil, nil
        for i = 0, 15 do
            sock, err = socket:tcp()
            if not sock then
                print(err)
                return
            end

            -- start a server:
            sni.server = sock

            -- DO NOT enable reuseaddr so that we get a clean error message and do not start overlapping with
            -- bound sockets created from a previous script if that script did not cleanly shut down.
            --sni.server:setoption("reuseaddr", true)

            res, err = sni.server:bind('localhost', 48896+i)
            if err == nil then
                sni.port = i
                print("server:bind(" .. (48896 + i) .. "): success")
                break
            end

            -- error, move on to the next port:
            print("server:bind(" .. (48896 + i) .. "): err=" .. err)
            sni.server:close()
            sni.server = nil
        end
        if err ~= nil then
            print("No open ports found to listen on. Please close this Lua Console window and re-open it to restart the server.")
            return
        end

        res, err = sni.server:listen()
        if err ~= nil then
            print("server:listen: err=" .. err)
            return
        end
    end

    -- main connection handling loop:
    while true do
        if state == "connected" then
            -- handle as many commands in a loop as possible before resuming the next frame:
            while receive() do end
        elseif state == "accept" then
            -- 1 frame of timeout worst case:
            sni.server:settimeout(0.015)
            local sock, err = sni.server:accept()
            if err == nil then
                print("server:accept: connection accepted")
                sock:settimeout(0)
                sock:setoption("tcp-nodelay", true)
                sni.client = sock
                state = "connected"
            elseif err ~= "timeout" then
                print("server:accept: error=`" .. err .. "`")
            end
        end

        emu.frameadvance()
    end
end

function shutdown()
    print("shutdown: shutting down")

    if sni.client ~= nil then
        print("shutdown: client close")
        sni.client:close()
        sni.client = nil
    end
    if sni.server ~= nil then
        print("shutdown: server close")
        sni.server:close()
        sni.server = nil
    end

    collectgarbage()
end

event.onexit(shutdown)
main()
