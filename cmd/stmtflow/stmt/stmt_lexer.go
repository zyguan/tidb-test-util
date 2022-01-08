// Code generated from stmt/Stmt.g4 by ANTLR 4.9.3. DO NOT EDIT.

package stmt

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 11, 107,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 3, 2, 6, 2, 23, 10, 2, 13,
	2, 14, 2, 24, 3, 3, 3, 3, 3, 3, 5, 3, 30, 10, 3, 3, 4, 3, 4, 3, 4, 3, 4,
	7, 4, 36, 10, 4, 12, 4, 14, 4, 39, 11, 4, 3, 4, 3, 4, 3, 4, 3, 5, 3, 5,
	3, 5, 3, 5, 5, 5, 48, 10, 5, 3, 5, 7, 5, 51, 10, 5, 12, 5, 14, 5, 54, 11,
	5, 3, 5, 3, 5, 5, 5, 58, 10, 5, 3, 6, 3, 6, 7, 6, 62, 10, 6, 12, 6, 14,
	6, 65, 11, 6, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 7, 7, 73, 10, 7, 12,
	7, 14, 7, 76, 11, 7, 3, 7, 3, 7, 3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 7,
	8, 86, 10, 8, 12, 8, 14, 8, 89, 11, 8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3,
	9, 3, 9, 3, 9, 7, 9, 99, 10, 9, 12, 9, 14, 9, 102, 11, 9, 3, 9, 3, 9, 3,
	10, 3, 10, 3, 37, 2, 11, 3, 3, 5, 4, 7, 5, 9, 6, 11, 7, 13, 8, 15, 9, 17,
	10, 19, 11, 3, 2, 7, 4, 2, 11, 11, 34, 34, 4, 2, 12, 12, 15, 15, 4, 2,
	41, 41, 94, 94, 4, 2, 36, 36, 94, 94, 4, 2, 94, 94, 98, 98, 2, 122, 2,
	3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2,
	11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2, 2,
	2, 19, 3, 2, 2, 2, 3, 22, 3, 2, 2, 2, 5, 29, 3, 2, 2, 2, 7, 31, 3, 2, 2,
	2, 9, 47, 3, 2, 2, 2, 11, 59, 3, 2, 2, 2, 13, 66, 3, 2, 2, 2, 15, 79, 3,
	2, 2, 2, 17, 92, 3, 2, 2, 2, 19, 105, 3, 2, 2, 2, 21, 23, 9, 2, 2, 2, 22,
	21, 3, 2, 2, 2, 23, 24, 3, 2, 2, 2, 24, 22, 3, 2, 2, 2, 24, 25, 3, 2, 2,
	2, 25, 4, 3, 2, 2, 2, 26, 30, 9, 3, 2, 2, 27, 28, 7, 15, 2, 2, 28, 30,
	7, 12, 2, 2, 29, 26, 3, 2, 2, 2, 29, 27, 3, 2, 2, 2, 30, 6, 3, 2, 2, 2,
	31, 32, 7, 49, 2, 2, 32, 33, 7, 44, 2, 2, 33, 37, 3, 2, 2, 2, 34, 36, 11,
	2, 2, 2, 35, 34, 3, 2, 2, 2, 36, 39, 3, 2, 2, 2, 37, 38, 3, 2, 2, 2, 37,
	35, 3, 2, 2, 2, 38, 40, 3, 2, 2, 2, 39, 37, 3, 2, 2, 2, 40, 41, 7, 44,
	2, 2, 41, 42, 7, 49, 2, 2, 42, 8, 3, 2, 2, 2, 43, 44, 7, 47, 2, 2, 44,
	45, 7, 47, 2, 2, 45, 48, 7, 34, 2, 2, 46, 48, 7, 37, 2, 2, 47, 43, 3, 2,
	2, 2, 47, 46, 3, 2, 2, 2, 48, 52, 3, 2, 2, 2, 49, 51, 10, 3, 2, 2, 50,
	49, 3, 2, 2, 2, 51, 54, 3, 2, 2, 2, 52, 50, 3, 2, 2, 2, 52, 53, 3, 2, 2,
	2, 53, 57, 3, 2, 2, 2, 54, 52, 3, 2, 2, 2, 55, 58, 5, 5, 3, 2, 56, 58,
	7, 2, 2, 3, 57, 55, 3, 2, 2, 2, 57, 56, 3, 2, 2, 2, 58, 10, 3, 2, 2, 2,
	59, 63, 7, 61, 2, 2, 60, 62, 5, 3, 2, 2, 61, 60, 3, 2, 2, 2, 62, 65, 3,
	2, 2, 2, 63, 61, 3, 2, 2, 2, 63, 64, 3, 2, 2, 2, 64, 12, 3, 2, 2, 2, 65,
	63, 3, 2, 2, 2, 66, 74, 7, 41, 2, 2, 67, 68, 7, 94, 2, 2, 68, 73, 11, 2,
	2, 2, 69, 70, 7, 41, 2, 2, 70, 73, 7, 41, 2, 2, 71, 73, 10, 4, 2, 2, 72,
	67, 3, 2, 2, 2, 72, 69, 3, 2, 2, 2, 72, 71, 3, 2, 2, 2, 73, 76, 3, 2, 2,
	2, 74, 72, 3, 2, 2, 2, 74, 75, 3, 2, 2, 2, 75, 77, 3, 2, 2, 2, 76, 74,
	3, 2, 2, 2, 77, 78, 7, 41, 2, 2, 78, 14, 3, 2, 2, 2, 79, 87, 7, 36, 2,
	2, 80, 81, 7, 94, 2, 2, 81, 86, 11, 2, 2, 2, 82, 83, 7, 36, 2, 2, 83, 86,
	7, 36, 2, 2, 84, 86, 10, 5, 2, 2, 85, 80, 3, 2, 2, 2, 85, 82, 3, 2, 2,
	2, 85, 84, 3, 2, 2, 2, 86, 89, 3, 2, 2, 2, 87, 85, 3, 2, 2, 2, 87, 88,
	3, 2, 2, 2, 88, 90, 3, 2, 2, 2, 89, 87, 3, 2, 2, 2, 90, 91, 7, 36, 2, 2,
	91, 16, 3, 2, 2, 2, 92, 100, 7, 98, 2, 2, 93, 94, 7, 94, 2, 2, 94, 99,
	11, 2, 2, 2, 95, 96, 7, 98, 2, 2, 96, 99, 7, 98, 2, 2, 97, 99, 10, 6, 2,
	2, 98, 93, 3, 2, 2, 2, 98, 95, 3, 2, 2, 2, 98, 97, 3, 2, 2, 2, 99, 102,
	3, 2, 2, 2, 100, 98, 3, 2, 2, 2, 100, 101, 3, 2, 2, 2, 101, 103, 3, 2,
	2, 2, 102, 100, 3, 2, 2, 2, 103, 104, 7, 98, 2, 2, 104, 18, 3, 2, 2, 2,
	105, 106, 11, 2, 2, 2, 106, 20, 3, 2, 2, 2, 16, 2, 24, 29, 37, 47, 52,
	57, 63, 72, 74, 85, 87, 98, 100, 2,
}

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames []string

var lexerSymbolicNames = []string{
	"", "SPACE", "NEWLINE", "BLOCK_COMMENT", "LINE_COMMENT", "SEMI", "SINGLE_QUOTE_STRING",
	"DOUBLE_QUOTE_STRING", "BACK_QUOTE_STRING", "ANY",
}

var lexerRuleNames = []string{
	"SPACE", "NEWLINE", "BLOCK_COMMENT", "LINE_COMMENT", "SEMI", "SINGLE_QUOTE_STRING",
	"DOUBLE_QUOTE_STRING", "BACK_QUOTE_STRING", "ANY",
}

type Stmt struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

// NewStmt produces a new lexer instance for the optional input antlr.CharStream.
//
// The *Stmt instance produced may be reused by calling the SetInputStream method.
// The initial lexer configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewStmt(input antlr.CharStream) *Stmt {
	l := new(Stmt)
	lexerDeserializer := antlr.NewATNDeserializer(nil)
	lexerAtn := lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)
	lexerDecisionToDFA := make([]*antlr.DFA, len(lexerAtn.DecisionToState))
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "Stmt.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// Stmt tokens.
const (
	StmtSPACE               = 1
	StmtNEWLINE             = 2
	StmtBLOCK_COMMENT       = 3
	StmtLINE_COMMENT        = 4
	StmtSEMI                = 5
	StmtSINGLE_QUOTE_STRING = 6
	StmtDOUBLE_QUOTE_STRING = 7
	StmtBACK_QUOTE_STRING   = 8
	StmtANY                 = 9
)
