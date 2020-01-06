log = require('log')
fiber = require('fiber')

box.cfg {
    listen = 3301;
    memtx_memory = 128 * 1024 * 1024; -- 128Mb
    memtx_min_tuple_size = 16;
    memtx_max_tuple_size = 128 * 1024 * 1024; -- 128Mb
    vinyl_memory = 128 * 1024 * 1024; -- 128Mb
    vinyl_cache = 128 * 1024 * 1024; -- 128Mb
    vinyl_max_tuple_size = 128 * 1024 * 1024; -- 128Mb
    vinyl_write_threads = 2;
    wal_mode = "none";
    wal_max_size = 256 * 1024 * 1024;
    checkpoint_interval = 60 * 60; -- one hour
    checkpoint_count = 6;
    force_recovery = true;
    net_msg_max = 2048;

     -- 1 – SYSERROR
     -- 2 – ERROR
     -- 3 – CRITICAL
     -- 4 – WARNING
     -- 5 – INFO
     -- 6 – VERBOSE
     -- 7 – DEBUG
     log_level = 4;
--      log_nonblock = true;
     too_long_threshold = 0.5;
}

function find_users_sql(prefix, minId, limit)
    return box.execute('select * from "mysql_users" where ("name" like \'' .. prefix .. '%\' or "last_name" like \'' .. prefix .. '%\') and "id">' .. minId ..  ' order by "id" asc limit ' .. limit)
end

-- function find_users(prefix, minId, limit)
--     return box.space.mysql_users.index.primary:select({}, {limit=1000})
-- end

function find_users(prefix, minId, limit)
    local result = {}
	local count = 0
    local lprefix = utf8.lower(prefix)
    for _, user in box.space.mysql_users.index.primary:pairs(minId, { iterator = 'GT' }) do
        local sub1 = utf8.lower(utf8.sub(user[2], 1, utf8.len(prefix)))
        local sub2 = utf8.lower(utf8.sub(user[3], 1, utf8.len(prefix)))
--         box.session.push(string.format("lprefix=%s sub1=%s sub2=%s s1=%d s2=%d", lprefix, sub1, sub2, s1, s2))
        if (utf8.cmp(sub1, lprefix) == 0) or (utf8.cmp(sub2, lprefix) == 0) then
            table.insert(result, user)
			count = count + 1
        end
		if count == limit then
			break
		end
        if count % 50 == 0 then
            fiber.yield()
        end
    end
    return result
end

local function bootstrap()
    if not box.space.mysqldaemon then
        s = box.schema.space.create('mysqldaemon',
		{id = 512})
        s:create_index('primary',
        {type = 'tree', parts = {1, 'unsigned'}})
    end

    if not box.space.mysql_users then
        t = box.schema.space.create('mysql_users',
		{id = 513})
        t:create_index('primary',
        {type = 'tree', parts = {1, 'integer'}})
        t:create_index('name',
        {type = 'tree', parts = {2, 'string', collation = 'unicode_ci'}, unique = false})
        t:create_index('last_name',
        {type = 'tree', parts = {3, 'string', collation = 'unicode_ci'}, unique = false})
        t:format({{'id','integer'},{'name','string'},{'last_name','string'}})
    end
    box.schema.user.grant('guest', 'read,write,execute', 'universe', nil, {if_not_exists=true})
end

bootstrap()
