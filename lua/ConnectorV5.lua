-- SNI BizHawk Connector V5
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
local state = "accept"
sni = {
	server = nil,
	port = 0,
	client = nil
}

-- escapes '\', '|', and ',' to '\xNN' where NN is the hex representation of the ASCII code of the char escaped:
local function escape(s)
    return string.gsub(s, '[\\|,]', function (m)
        return string.format('\\x%02x', string.byte(m))
    end)
end

-- unescapes '\xNN' where NN is the hex representation of the ASCII code of the char to unescape:
local function unescape(s)
    return string.gsub(s, '\\x[0-9a-f][0-9a-f]', function (m)
        return string.char(tonumber(string.sub(m,3,4),16))
    end)
end

-- encodes strings into escaped strings and encodes array-tables into comma-delimited encoded strings:
local function encode(s)
    if type(s) == "string" then
        return escape(s)
    elseif type(s) == "table" then
		local d = {}
		for i,v in ipairs(s) do
		    d[i] = encode(v)
		end
		return table.concat(d, ',')
    else
        return tostring(s)
    end
end

-- decodes a comma-delimited string into an array-table and unescapes each string element:
local function decode_list(s)
    local t = {}
    for item in string.gmatch("[^,]+") do
        t[#t+1] = unescape(item)
    end
    return t
end

function handle(req_headers, req_body)
	-- command:
	local cmd = req_headers["cmd"]

	if cmd == "info" then
		-- collect all current info about emulator, core, and game:

		-- get memory domain names and their sizes:
		local domain_names = memory.getmemorydomainlist()
		local domain_sizes = {}
		for i,k in ipairs(domain_names) do
			domain_sizes[i] = memory.getmemorydomainsize(k)
		end

        -- response:
		return req_headers, {
		    client_version = client.getversion(),
		    platform = string.lower(emu.getsystemid()),
            rom_name = gameinfo.getromname(),
            rom_hash = gameinfo.getromhash(),
		    domain_names = domain_names,
		    domain_sizes = domain_sizes
        }
	end

    req_headers["error"] = "unknown command"
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

	-- separate request headers from body by first instance of "||"
	-- decode a delimited key-value map e.g. "key=value|key=value|key=value|...|"
	local req_headers, req_body = {}, {}
	local reqpart = req_headers
	if l[-1] ~= '|' then l = l .. '|' end
    for m in string.gmatch(l, '([^|]*)|') do
        if #m == 0 then
            -- an empty header denotes the transition between headers and body:
            reqpart = req_body
        else
            local i = string.find(m, '=', 1, true)
            local k,v
            if i == nil then
                k,v = m, ""
            else
                k,v = string.sub(m, 1, i-1), string.sub(m, i+1)
            end
            reqpart[k] = v
        end
    end

    -- handle request:
	local rsp_headers, rsp_body = handle(req_headers, req_body)

	-- format response message:
	local sb = {}
	for k,v in pairs(rsp_headers) do
		sb[#sb+1] = k
		sb[#sb+1] = "="
		sb[#sb+1] = encode(v)
        sb[#sb+1] = "|"
	end
	if rsp_body ~= nil then
        sb[#sb+1] = "|"
        for k,v in pairs(rsp_body) do
            sb[#sb+1] = k
            sb[#sb+1] = "="
            sb[#sb+1] = encode(v)
            sb[#sb+1] = "|"
        end
	end
	local rsp = table.concat(sb)

	print("response: `" .. rsp .. "`")
	sni.client:send(rsp .. "\n")
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
				print("server:accept: " .. err)
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
