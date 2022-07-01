sysbench.cmdline.options = {
    reconnect = {"Reconnect after every N events. The default (0) is to not reconnect", 0},
}

cursize=0

-- 10 groups, 119 characters
local c_value_template = "###########-###########-###########-" ..
    "###########-###########-###########-" ..
    "###########-###########-###########-" ..
    "###########"

function get_c_value()
    return sysbench.rand.string(c_value_template)
end

function thread_init()
   drv = sysbench.sql.driver()
   con = drv:connect()
end

function thread_done()
    con:bulk_insert_done()
    con:disconnect()
end

function prepare()
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    con:query(string.format([[
        CREATE TABLE IF NOT EXISTS sbtest (
            id INTEGER NOT NULL AUTO_INCREMENT,
            c1 CHAR(120) DEFAULT '' NOT NULL,
            c2 CHAR(120) DEFAULT '' NOT NULL,
            c3 CHAR(120) DEFAULT '' NOT NULL,
            c4 CHAR(120) DEFAULT '' NOT NULL,
            PRIMARY KEY (id), KEY (c1), KEY (c2), KEY (c3), KEY (c4))]]))
end

function cleanup()
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    con:query("DROP TABLE IF EXISTS sbtest")
end

function event()
    local c = get_c_value()
    if (cursize == 0) then
        con:bulk_insert_init("INSERT INTO sbtest VALUES ")
    end

    cursize = cursize + 1

    con:bulk_insert_next("(0,'" .. c .. "','" .. c .. "','" .. c .. "','" .. c .. "')")
    if sysbench.opt.reconnect > 0 and cursize % sysbench.opt.reconnect == 0 then
        con:bulk_insert_done()
        con:reconnect()
        cursize = 0
    end
end
