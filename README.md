# gmtc

Currently in its infancy, this will (maybe) one day be a type checker for GameMaker games. Right now, it has a custom-built parser, which can be used to parse single files into an AST. It can also be pointed at a project file, which it will parse, and then collect all the related script files, parsing each of them individually.

Very early stages, probably buggy, although I do use some fairly large Game Maker libraries to test it (primarily [Scribble by Juju Adams](https://github.com/JujuAdams/scribble)).
