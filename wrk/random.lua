request = function()
    local path = wrk.path .. math.random(1,990000)
    return wrk.format(wrk.method, path, wrk.headers, wrk.body)
end