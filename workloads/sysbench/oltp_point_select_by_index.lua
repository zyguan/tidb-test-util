require("oltp_common")

function prepare_statements()
    sysbench.opt.point_selects = 1

    local stmt_pattern = "SELECT c FROM sbtest%u WHERE k=?"
    for t = 1, sysbench.opt.tables do
        stmt[t].point_select_by_index = con:prepare(string.format(stmt_pattern, t))
        param[t].point_select_by_index = {stmt[t].point_select_by_index:bind_create(sysbench.sql.type.INT)}
        stmt[t].point_select_by_index:bind_param(unpack(param[t].point_select_by_index))
     end
end

function event()
    local tnum = sysbench.rand.uniform(1, sysbench.opt.tables)
    for i = 1, sysbench.opt.point_selects do
        param[tnum].point_select_by_index[1]:set(sysbench.rand.default(1, sysbench.opt.table_size))
        stmt[tnum].point_select_by_index:execute()
    end

    check_reconnect()
end
