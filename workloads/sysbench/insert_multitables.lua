sysbench.cmdline.options = {
    tables   = {"Number of tables", 1000},
    batch_size = {"Number of rows per insert", 10},
}

function thread_init()
    drv = sysbench.sql.driver()
    con = drv:connect()
end

function thread_done()
    con:disconnect()
end

function event()
    local i = sysbench.rand.uniform(1, sysbench.opt.tables)
    con:bulk_insert_init("insert into t"..i.." values ")
    for i = 1, sysbench.opt.batch_size do
        con:bulk_insert_next("(0,"..sysbench.rand.default(1, sysbench.opt.tables)..")")
    end
    con:bulk_insert_done()
end

function prepare()
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    for i = sysbench.tid % sysbench.opt.threads + 1, sysbench.opt.tables, sysbench.opt.threads do
        print("creating table t"..i.." ...")
        con:query("create table if not exists t"..i.."(id int not null auto_increment, k int, primary key (id), key (k))")
    end
end

function cleanup()
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    for i = sysbench.tid % sysbench.opt.threads + 1, sysbench.opt.tables, sysbench.opt.threads do
        print("dropping table t"..i.." ...")
        con:query("drop table if exists t"..i)
    end
end

sysbench.cmdline.commands = {
    prepare = {prepare, sysbench.cmdline.PARALLEL_COMMAND},
    cleanup = {cleanup, sysbench.cmdline.PARALLEL_COMMAND},
}
