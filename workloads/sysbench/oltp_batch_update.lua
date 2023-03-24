require("oltp_common")

sysbench.cmdline.options.batch_updates = {
    'Number of batch updates per transaction', 1}
sysbench.cmdline.options.batch_size = {
    'Batch size (number of records per update)', 10}
sysbench.cmdline.options.update_index = {
    'Update index column', false}

function prepare_statements()
    if not sysbench.opt.skip_trx then
        prepare_begin()
        prepare_commit()
    end

    local xs = {}
    local stmt_pattern = ''
    for i = 1, sysbench.opt.batch_size do table.insert(xs, "?") end
    if sysbench.opt.update_index then
        stmt_pattern = "UPDATE sbtest%u SET k=k+1 WHERE id in ("..table.concat(xs, ",")..")"
    else
        stmt_pattern = "UPDATE sbtest%u SET c=REVERSE(c) WHERE id in ("..table.concat(xs, ",")..")"
    end
    for t = 1, sysbench.opt.tables do
        stmt[t].batch_update = con:prepare(string.format(stmt_pattern, t))
        param[t].batch_update = {}
        for p = 1, sysbench.opt.batch_size do
            param[t].batch_update[p] = stmt[t].batch_update:bind_create(sysbench.sql.type.INT)
        end
        stmt[t].batch_update:bind_param(unpack(param[t].batch_update))
     end
end

function event()
    if not sysbench.opt.skip_trx then
        begin()
    end

    local tnum = sysbench.rand.uniform(1, sysbench.opt.tables)
    for i = 1, sysbench.opt.batch_updates do
        for j = 1, sysbench.opt.batch_size do
            param[tnum].batch_update[j]:set(sysbench.rand.default(1, sysbench.opt.table_size))
        end
        stmt[tnum].batch_update:execute()
    end

    if not sysbench.opt.skip_trx then
        commit()
    end

    check_reconnect()
end
