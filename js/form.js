function form_field_init(elm) {
    if (elm.hasClass('tcal')) {
        f_tcalInit();
    } else if (elm.hasClass('tmcl')) {
        elm.timepicker({ timeFormat: 'H:mm' });
    }
}

