#macro    DEBUG  true
#macro MULTILINE_TEST   10 \
    + 5

var still = "here";

#macro COMMENT_TEST \ // this comment should be skipped
    var good = true

function make_game() {
    // Hello world, this is a comment!

    var a = 10;
    COMMENT_TEST;

    /*
        And this is a multiline comment.
    */

    if (DEBUG) {
        print("Hello, world!");
    }

    var fifteen = MULTILINE_TEST;

    while (true) {
        draw_text(1, 1, "Hello, world!");
    }

    obj[0].draw_sprite(10, 10, true, false, "hello, world!");

    if (is_array(_generator_state.__bezier_lengths_array))
    {
        //Prep for Bezier curve shenanigans if necessary
        var _bezier_do              = true;
        var _bezier_lengths         = _generator_state.__bezier_lengths_array;
        var _bezier_search_index    = 0;
        var _bezier_search_d0       = 0;
        var _bezier_search_d1       = _bezier_lengths[1];
        var _bezier_prev_cy         = -infinity;
        var _bezier_param_increment = 1 / (SCRIBBLE_BEZIER_ACCURACY-1);
    }
    else
    {
        _bezier_do = false;
    }

    if (_is_krutidev) _font_data.__is_krutidev = true;
    var _unicode = is_real(_character)? _character : ord(_character);

    if (_colourDict[$ _name] != _colour)
    {
        _colourDict[$ _name] = _colour;

        //Ensure that any custom colours that are in text elements are updated
        scribble_refresh_everything();
    }


    function __scribble_class_event(_name, _data) constructor
    {
        //These are publicly exposed via .get_events()
        name            = _name;
        data            = _data;
        position        = undefined; //Legacy
        character_index = undefined;
        line_index      = undefined;
    }

    static __error = function()
    {
        __scribble_error("Cannot call text element methods using the result from .draw()\nThis can occur in two situations:\n  1. scribble().draw().method();\n  2. t = scribble().draw(); t.method()\n\nInstead use:\n  1. scribble().method().draw();\n  2. t = scribble(); t.method(); t.draw();");
    }

    var _glyph_index = _map[? 0x20];
    vertex_color(_vbuff, c_white, 1.0);

    switch (abcde) {
        case 1: return 1
        default: return 10
    }
}
