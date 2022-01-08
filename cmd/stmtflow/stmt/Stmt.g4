lexer grammar Stmt;

SPACE:                               [ \t]+;
NEWLINE:                             '\n' | '\r' | '\r\n';
BLOCK_COMMENT:                       '/*' .*? '*/';
LINE_COMMENT:                        ('-- ' | '#') ~[\r\n]* (NEWLINE | EOF);

SEMI:                                ';' SPACE*;

SINGLE_QUOTE_STRING:                       '\'' ('\\'. | '\'\'' | ~('\'' | '\\'))* '\'';
DOUBLE_QUOTE_STRING:                       '"' ( '\\'. | '""' | ~('"'| '\\') )* '"';
BACK_QUOTE_STRING:                       '`' ( '\\'. | '``' | ~('`'|'\\'))* '`';

ANY:                                 .;
