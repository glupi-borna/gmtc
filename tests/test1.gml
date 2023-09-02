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

    draw_sprite(10, 10, true, false, "hello, world!");
}
