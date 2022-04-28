require("oltp_common")

sysbench.cmdline.options.batch_index_updates = {
    'Number of BATCH UPDATE index queries per transaction', 1}
sysbench.cmdline.options.batch_size = {
    'Size of batch queries', 10}

function prepare_statements()
    if not sysbench.opt.skip_trx then
        prepare_begin()
        prepare_commit()
    end

    local xs = {}
    for i = 1, sysbench.opt.batch_size do table.insert(xs, "?") end
    local stmt_pattern = "UPDATE sbtest%u SET k=k+1 WHERE id in ("..table.concat(xs, ",")..")"
    for t = 1, sysbench.opt.tables do
        stmt[t].update_index_batch = con:prepare(string.format(stmt_pattern, t))
        param[t].update_index_batch = {}
        for p = 1, sysbench.opt.batch_size do
            param[t].update_index_batch[p] = stmt[t].update_index_batch:bind_create(sysbench.sql.type.INT)
        end
        stmt[t].update_index_batch:bind_param(unpack(param[t].update_index_batch))
     end
end

function event()
    if not sysbench.opt.skip_trx then
        begin()
    end

    local tnum = sysbench.rand.uniform(1, sysbench.opt.tables)
    for i = 1, sysbench.opt.batch_index_updates do
        for j = 1, 10 do
            param[tnum].update_index_batch[j]:set(sysbench.rand.default(1, sysbench.opt.table_size))
        end
        stmt[tnum].update_index_batch:execute()
    end

    if not sysbench.opt.skip_trx then
        commit()
    end

    check_reconnect()
end
