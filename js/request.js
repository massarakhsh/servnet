
function request_select_cmd(path) {
    setTimeout(function() {
        let dt = jQuery("#apicmd").val();
        front_post(path, "data=" + string_to_XS(dt));
        request_build();
    }, 100);
}

function request_select_table(path) {
    setTimeout(function() {
        let dt = jQuery("#apitable").val();
        front_post(path, "data=" + string_to_XS(dt));
        request_build();
    }, 100);
}

function request_select_id(path) {
    setTimeout(function() {
        let dt = jQuery("#apiid").val();
        front_post(path, "data=" + string_to_XS(dt));
        request_build();
    }, 100);
}

function request_select_phone(path) {
    setTimeout(function() {
        let dt = jQuery("#apiphone").val();
        front_post(path, "data=" + string_to_XS(dt));
        request_build();
    }, 100);
}

function request_select_string(path) {
    setTimeout(function() {
        let dt = jQuery("#apistring").val();
        front_post(path, "data=" + string_to_XS(dt));
        request_build();
    }, 100);
}

function request_select_prm(path) {
    setTimeout(function() {
        let dt = jQuery("#apiprm").val();
        front_post(path, "data=" + string_to_XS(dt));
    }, 100);
}

function request_call(path) {
    let call = request_build();
    if (call) {
        let data = null;
        if (jQuery("#apistring").size() > 0) {
            let parm = jQuery("#apistring").val();
            if (parm) data = parm;
        } else if (jQuery("#apiprm").size() > 0) {
            let parm = jQuery("#apiprm").val();
            if (parm) data = parm;
        }
        request_ajax("POST", call, data, path);
    }
}

function request_build() {
    let call = "";
    let cmd = jQuery("#apicmd").val();
    if (cmd) {
        call += "/api/" + cmd;
    }
    if (jQuery("#apitable").size() > 0) {
        let val = jQuery("#apitable").val();
        if (!val) val = "ERR";
        call += "/" + val;
    }
    if (jQuery("#apiid").size() > 0) {
        let val = jQuery("#apiid").val();
        if (!(val>0)) val = "ERR";
        call += "/" + val;
    }
    if (jQuery("#apiphone").size() > 0) {
        let val = jQuery("#apiphone").val();
        val = val.replace(/\D/g, "");
        if (!val) val = "ERR";
        call += "/" + val;
    }
    let text = call;
    if (jQuery("#apistring").size() > 0 || jQuery("#apiprm").size()) {
        text += " (+data)";
    }
    jQuery("#apicall").text(text);
    return call;
}

function request_ajax(type, url, data, answer) {
    if (!data) {
    } else if (/^{/.exec(data)) {
    } else if (/^data=/.exec(data)) {
    } else {
        data = "data=" + data;
    }
    jQuery.ajax({
        type: type,
        url: url,
        data: data,
        dataType: "json",
        success: function(code) {
            request_answer(answer, code);
        },
        error: function(xhr, status) {
            request_answer(answer, null);
        }
    });
}

function request_answer(answer, lika) {
    var data = "";
    if (lika) {
        let json = JSON.stringify(lika);
        data ="answer=" + string_to_XS(json);
    }
    front_post(answer, data);
}

