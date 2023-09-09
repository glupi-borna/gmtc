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
}



#macro __SCRIBBLE_PARSER_WRITE_GLYPH  ;\//Pull info out of the font's data structures
                                      ;\//We floor this value to work around floating point issues on HTML5
                                      var _data_index = _font_glyphs_map[? floor(_glyph_write)];\
                                      ;\//If our glyph is missing, choose the missing character glyph instead!
                                      if (_data_index == undefined)\
                                      {\
                                          __scribble_trace("Couldn't find glyph data for character code " + string(_glyph_write) + " (" + chr(_glyph_write) + ") in font \"" + string(_font_name) + "\"");\
                                          _data_index = _font_glyphs_map[? ord(SCRIBBLE_MISSING_CHARACTER)];\
                                      }\
                                      if (_data_index == undefined)\
                                      {\
                                          ;\//This should only happen if SCRIBBLE_MISSING_CHARACTER is missing for a font
                                          __scribble_trace("Couldn't find \"missing character\" glyph data, character code " + string(ord(SCRIBBLE_MISSING_CHARACTER)) + " (" + string(SCRIBBLE_MISSING_CHARACTER) + ") in font \"" + string(_font_name) + "\"");\
                                      }\
                                      else\
                                      {\
                                          ;\//Add this glyph to our grid by copying from the font's own glyph data grid
                                          ds_grid_set_grid_region(_glyph_grid, _font_glyph_data_grid, _data_index, SCRIBBLE_GLYPH.UNICODE, _data_index, SCRIBBLE_GLYPH.BILINEAR, _glyph_count, __SCRIBBLE_GEN_GLYPH.__UNICODE);\
                                          _glyph_grid[# _glyph_count, __SCRIBBLE_GEN_GLYPH.__CONTROL_COUNT] = _control_count;\
                                          ;\
                                          if (SCRIBBLE_USE_KERNING)\
                                          {\
                                              var _kerning = _font_kerning_map[? ((_glyph_write & 0xFFFF) << 16) | (_glyph_prev & 0xFFFF)];\
                                              if (_kerning != undefined) _glyph_grid[# _glyph_count-1, __SCRIBBLE_GEN_GLYPH.__SEPARATION] += _kerning;\
                                          }\
                                          ;\
                                          __SCRIBBLE_PARSER_NEXT_GLYPH\
                                      }

__SCRIBBLE_PARSER_WRITE_GLYPH
