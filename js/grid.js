var grid_Instant;
var row_likSelect;

function grid_redraw(elm) {
    let path = elm.attr('path')+'init';
    front_proc(path, grid_redraw_init, elm);
}

function grid_redraw_init(elm, lika) {
    let instant = ('grid' in lika) ? lika.grid : [];
    instant['ajax'] = lik_build_url('/front'+ elm.attr('path') + 'data');
    grid_prepare(instant);
    let grid = elm.DataTable(instant);
    grid_Instant = instant;
    row_likSelect = ('likSelect' in instant) ? instant.likSelect : 0;
    grid.on( 'select', grid_select);
    grid.on( 'draw', grid_draw_done);
}

function grid_prepare(data) {
    if (data !== null && typeof(data) == 'object') {
        for (var key in data) {
            let value = data[key];
            if (typeof(value) == 'string') {
                var match;
                if (match = /^function_(.+)\((.*)\)/.exec(value)) {
                    let func = match[1];
                    let parm = match[2];
                    if (func in window) {
                        data[key] = function () {
                            window[func](this, parm);
                        };
                    } else {
                        data[key] = grid_nothing;
                    }
                } else if (match = /^function_(.+)/.exec(value)) {
                    let func = match[1];
                    if (func in window) {
                        data[key] = window[func];
                    } else {
                        data[key] = grid_nothing;
                    }
                }
            } else if (value !== null && typeof(value) == 'object') {
                grid_prepare(data[key]);
            }
        }
    }
}

function grid_nothing() {
}

function grid_select( e, dt, type, indexes ) {
    if ( type === 'row' ) {
        if (indexes && indexes.length > 0) {
            var datas = dt.rows(indexes).data();
            if (datas && datas.length>0) {
                row_likSelect = datas[0].DT_RowId;
                let path = grid_Instant.ajax;
                if (match = /^(.+)griddata(.*)$/.exec(path)) {
                    path = match[1] + "select/" + row_likSelect + match[2];
                    get_data_part(path);
                }
            }
        }
    }
}

function grid_draw_done( e, settings ) {
    //alert('grid_draw_done');
    var api = new $.fn.dataTable.Api( settings );
    api.rows().eq(0).each( function ( index ) {
        var row = api.row( index );
        var data = row.data();
        if (data.DT_RowId == row_likSelect) {
            row.select();
        }
        // ... do something with data(), or row.node(), etc
    } );    //api.row(':eq(7)').select();
    //alert(api.rows().length);
    //console.log( api.rows( {page:'current'} ).data() );
}

function grid_refresh(parm) {
    if (match = /(.*?)__(.*?)__(.*)/.exec(parm)) {
        let id = match[1];
        let grid_index = match[2];
        let grid_length = match[3];
        let elm = jQuery("#" + id);
        if (elm) {
            let grid = elm.DataTable();
            if (grid_index != elm.attr('grid_index') || grid_length < elm.attr('grid_length')) {
                elm.attr('grid_index', grid_index);
                grid.clear();
            }
            grid_refresh_next(elm);
        }
    }
}

function grid_refresh_next(elm) {
    let grid = elm.DataTable();
    let path = elm.attr('path') + 'refresh/' + elm.attr('grid_index') + "/" + grid.data().length;
    front_proc(path, grid_refresh_load, elm);
}

function grid_refresh_load(elm, lika) {
    let data = ('data' in lika) ? lika.data : [];
    let index = ('index' in lika) ? lika.index : 0;
    let start = ('start' in lika) ? lika.start : 0;
    let grid = elm.DataTable();
    if (index != elm.attr('grid_index')) {
        elm.attr('grid_index', index);
        grid.clear();
    }
    let dep = 0;
    let have = grid.data().length;
    if (have > start) dep = have - start;
    for (n = dep; n < data.length; n++) {
        let line = data[n];
        grid.row.add(line);
    }
    grid.draw();
    if (data.length > 0) {
        grid_refresh_next(elm);
    }
}

function grid_clear(id) {
    let elm = jQuery("#"+id);
    if (elm) {
        let grid = elm.DataTable();
        grid.clear();
        grid.draw();
    }
}

