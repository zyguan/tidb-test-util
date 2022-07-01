sysbench.cmdline.options = {
    truncate   = {"Truncate after every N events. The default (0) is to not truncate", 0},
    batch_size = {"Number of rows for each insert", 100},
}

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
    con:disconnect()
end

function prepare()
    local i
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    for i = 1, sysbench.opt.threads do
        print("Creating table 'truncate" .. i .. "'...")
        con:query(string.format([[
            CREATE TABLE IF NOT EXISTS truncate%d (
                id INTEGER NOT NULL AUTO_INCREMENT,
                c1 CHAR(120) DEFAULT '' NOT NULL,
                c2 CHAR(120) DEFAULT '' NOT NULL,
                c3 CHAR(120) DEFAULT '' NOT NULL,
                c4 CHAR(120) DEFAULT '' NOT NULL,
                PRIMARY KEY (id), KEY (c1), KEY (c2), KEY (c3), KEY (c4))]], i))
    end
end

function cleanup()
    local i
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    for i = 1, sysbench.opt.threads do
        print("Dropping table 'truncate" .. i .. "'...")
        con:query("DROP TABLE IF EXISTS truncate" .. i)
    end
end

events = 0

function event()
    local c = get_c_value()

    con:bulk_insert_init("INSERT INTO truncate" .. sysbench.tid+1 .. " VALUES ")
    for i = 1, sysbench.opt.batch_size do
        con:bulk_insert_next("(0,'" .. c .. "','" .. c .. "','" .. c .. "','" .. c .. "')")
    end
    con:bulk_insert_done()

    events = events + 1

    if sysbench.opt.truncate > 0 and events % sysbench.opt.truncate == 0 then
        con:query("TRUNCATE TABLE truncate" .. sysbench.tid+1)
    end
end
