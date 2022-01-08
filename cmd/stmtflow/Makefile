.PHONY: default

default: stmt/stmt_lexer.go

stmt/stmt_lexer.go: stmt/Stmt.g4 antlr.jar
	java -jar antlr.jar -Dlanguage=Go -package stmt stmt/Stmt.g4

antlr.jar:
	wget -O $@ https://www.antlr.org/download/antlr-4.9.3-complete.jar
