require("oltp_common")

function prepare_statements()
    if not sysbench.opt.skip_trx then
        prepare_begin()
        prepare_commit()
    end

    prepare_point_selects()

    local stmt_pattern = "SELECT c FROM sbtest%u WHERE id=? FOR UPDATE"
    for t = 1, sysbench.opt.tables do
        stmt[t].point_select_for_update = con:prepare(string.format(stmt_pattern, t))
        param[t].point_select_for_update = {stmt[t].point_select_for_update:bind_create(sysbench.sql.type.INT)}
        stmt[t].point_select_for_update:bind_param(unpack(param[t].point_select_for_update))
     end
end

function event()
    if not sysbench.opt.skip_trx then
        begin()
    end

    local tnum = sysbench.rand.uniform(1, sysbench.opt.tables)
    for i = 1, sysbench.opt.point_selects do
        local id = sysbench.rand.default(1, sysbench.opt.table_size)
        if sysbench.rand.uniform(0, 10000) / 10000 <= sysbench.opt.for_update_proportion then
            param[tnum].point_select_for_update[1]:set(id)
            stmt[tnum].point_select_for_update:execute()
        else
            param[tnum].point_selects[1]:set(id)
            stmt[tnum].point_selects:execute()
        end
    end

    if not sysbench.opt.skip_trx then
        commit()
    end

    check_reconnect()
end

sysbench.cmdline.options.point_selects[2] = 1
sysbench.cmdline.options.for_update_proportion = {'Proportion of select for update', 0.5}
