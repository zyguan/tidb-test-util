require("oltp_common")

function prepare_statements()
end

function event()
end

function create_table(drv, con, table_num)
    local id_index_def, id_def
    local engine_def = ""
    local extra_table_options = ""
    local query

    if sysbench.opt.secondary then
        id_index_def = "KEY xid"
    else
        id_index_def = "PRIMARY KEY"
    end

    if drv:name() == "mysql"
    then
        if sysbench.opt.auto_inc then
            id_def = "BIGINT UNSIGNED NOT NULL AUTO_INCREMENT"
        else
            id_def = "BIGINT UNSIGNED NOT NULL"
        end
        engine_def = "/*! ENGINE = " .. sysbench.opt.mysql_storage_engine .. " */"
    elseif drv:name() == "pgsql"
    then
        if not sysbench.opt.auto_inc then
            id_def = "INTEGER NOT NULL"
        elseif pgsql_variant == 'redshift' then
            id_def = "INTEGER IDENTITY(1,1)"
        else
            id_def = "SERIAL"
        end
    else
        error("Unsupported database driver:" .. drv:name())
    end

    print(string.format("Creating table 'sbtest%d'...", table_num))

    query = string.format([[
CREATE TABLE sbtest%d(
  id %s,
  k INTEGER DEFAULT '0' NOT NULL,
  c CHAR(120) DEFAULT '' NOT NULL,
  pad CHAR(60) DEFAULT '' NOT NULL,
  %s (id)
) %s %s]],
        table_num, id_def, id_index_def, engine_def,
        sysbench.opt.create_table_options or '')

    con:query(query)

    if sysbench.opt.create_secondary then
        print(string.format("Creating a secondary index on 'sbtest%d'...",
                            table_num))
        con:query(string.format("CREATE INDEX k_%d ON sbtest%d(k)",
                                table_num, table_num))
    end
end

function insert_data(con, drv)
    local n = math.floor(sysbench.opt.table_size / sysbench.opt.threads)
    local k = sysbench.opt.table_size % sysbench.opt.threads
    local off = sysbench.tid * n + math.min(sysbench.tid, k)
    local len = sysbench.tid < k and n+1 or n

    for table_num = 1, sysbench.opt.tables do
        print(string.format("Inserting %d records into 'sbtest%d' by thread-%d",
                            len, table_num, sysbench.tid))

        if sysbench.opt.auto_inc then
            query = "INSERT INTO sbtest" .. table_num .. "(k, c, pad) VALUES"
        else
            query = "INSERT INTO sbtest" .. table_num .. "(id, k, c, pad) VALUES"
        end

        con:bulk_insert_init(query)

        local c_val
        local pad_val

        for i = off+1, off+len do

            c_val = get_c_value()
            pad_val = get_pad_value()

            if (sysbench.opt.auto_inc) then
               query = string.format("(%d, '%s', '%s')",
                                     sysbench.rand.default(1, sysbench.opt.table_size),
                                     c_val, pad_val)
            else
               query = string.format("(%d, %d, '%s', '%s')",
                                     i,
                                     sysbench.rand.default(1, sysbench.opt.table_size),
                                     c_val, pad_val)
            end

            con:bulk_insert_next(query)
        end

        con:bulk_insert_done()
    end
end

function cmd_create()
    local drv = sysbench.sql.driver()
    local con = drv:connect()

    for i = sysbench.tid % sysbench.opt.threads + 1, sysbench.opt.tables, sysbench.opt.threads do
        create_table(drv, con, i)
    end
end

function cmd_insert()
    local drv = sysbench.sql.driver()
    local con = drv:connect()
    insert_data(con, drv)
end

sysbench.cmdline.options.auto_inc[2] = false
sysbench.cmdline.commands.create = {cmd_create, sysbench.cmdline.PARALLEL_COMMAND}
sysbench.cmdline.commands.insert = {cmd_insert, sysbench.cmdline.PARALLEL_COMMAND}
