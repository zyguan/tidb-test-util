package core

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zyguan/tidb-test-util/cmd/stmtflow/stmt"

	. "github.com/zyguan/tidb-test-util/pkg/stmtflow"
)

func ParseSQL(r io.Reader) []Stmt {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil
	}
	return split(string(raw))
}

func split(text string) []Stmt {
	var (
		stmts  []Stmt
		tokens []antlr.Token
	)
	lexer := stmt.NewStmt(antlr.NewInputStream(text))
	token := lexer.NextToken()
	for {
		if typ := token.GetTokenType(); typ == antlr.TokenEOF {
			if s, ok := toStmt(tokens); ok {
				stmts = append(stmts, s)
			}
			return stmts
		} else if typ == stmt.StmtSEMI {
			// append ;
			tokens = append(tokens, token)
			token = lexer.NextToken()
			// append tailing whitespaces & first line comment
			for hasTypeOf(token, stmt.StmtSPACE, stmt.StmtNEWLINE, stmt.StmtLINE_COMMENT) {
				tokens = append(tokens, token)
				token = lexer.NextToken()
				if token.GetTokenType() != stmt.StmtSPACE {
					break
				}
			}
			if s, ok := toStmt(tokens); ok {
				stmts = append(stmts, s)
			}
			tokens = tokens[:0]
		} else {
			tokens = append(tokens, token)
			token = lexer.NextToken()
		}
	}
}

func toStmt(tokens []antlr.Token) (Stmt, bool) {
	if len(tokens) == 0 {
		return Stmt{}, false
	}
	i := 0
	// skip heading whitespaces & line comments
	for i < len(tokens) && hasTypeOf(tokens[i], stmt.StmtSPACE, stmt.StmtNEWLINE, stmt.StmtLINE_COMMENT) {
		i++
	}
	if i == len(tokens) || tokens[i].GetTokenType() != stmt.StmtBLOCK_COMMENT {
		return Stmt{}, false
	}
	cmd := strings.TrimSpace(strings.Trim(tokens[i].GetText(), "/*"))
	buf := new(strings.Builder)
	prefixSize, prefixMode := 0, true
	for i < len(tokens) {
		if prefixMode {
			if hasTypeOf(tokens[i], stmt.StmtSPACE, stmt.StmtNEWLINE, stmt.StmtBLOCK_COMMENT, stmt.StmtLINE_COMMENT) {
				prefixSize += len(tokens[i].GetText())
			} else {
				prefixMode = false
			}
		}
		buf.WriteString(tokens[i].GetText())
		i++
	}
	s := Stmt{SQL: strings.TrimRight(buf.String(), "\r\n")}
	if isQuery(s.SQL[prefixSize:]) {
		s.Flags |= S_QUERY
	}
	if k := strings.Index(cmd, ":"); k >= 0 {
		s.Sess = cmd[:k]
		for _, m := range strings.Split(cmd[k+1:], ",") {
			m = strings.TrimSpace(m)
			if len(m) == 0 {
				continue
			}
			switch strings.ToLower(m) {
			case "wait":
				s.Flags |= S_WAIT
			case "query":
				s.Flags |= S_QUERY
			case "unordered":
				s.Flags |= S_UNORDERED
			}
		}
	} else {
		s.Sess = cmd
	}
	return s, true
}

func hasTypeOf(token antlr.Token, types ...int) bool {
	typ := token.GetTokenType()
	for _, t := range types {
		if typ == t {
			return true
		}
	}
	return false
}

func isQuery(sql string) bool {
	// a naive impl
	sql = strings.ToLower(strings.TrimLeft(strings.TrimSpace(sql), "("))
	for _, w := range []string{"select ", "show ", "admin show ", "explain ", "desc ", "describe "} {
		if strings.HasPrefix(sql, w) {
			return true
		}
	}
	return false
}
