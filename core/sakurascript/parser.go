package sakurascript

import (
	"strings"
)

const PString = 0
const PTag = 1
const PEnv = 2

type Ast struct {
	Kind int
	Str string
	Args []string
}

func Parse(tokens []Token) []Ast {
	in_q := false
	in_a := false
	in_raw := false
	ast := make([]Ast, 0)
	index := 0
	for {
		if index >= len(tokens) {
			return ast
		}
		t := tokens[index]
		index++
		switch t.kind {
		case lTag:
			if in_raw {
				switch t.str {
				case `\_!`, `\_?`:
					in_raw = false
				default:
					ast = append(ast, Ast{
						Kind: PString,
						Str: t.str,
						Args: make([]string, 0),
					})
				}
				continue
			}
			switch t.str {
			case `\%`:
				ast = append(ast, Ast{
					Kind: PString,
					Str: "%",
					Args: make([]string, 0),
				})
			case `\0`, `\1`, `\4`, `\5`, `\6`, `\7`, `\__c`, `\__t`, `\__v`, `\_V`, `\_n`, `\_q`, `\_+`, `\C`, `\a`, `\e`, `\h`, `\t`, `\u`, `\v`, `\z`, `\*`, `\-`, `\+`:
				ast = append(ast, Ast{
					Kind: PTag,
					Str: t.str,
					Args: make([]string, 0),
				})
			case `\b`, `\p`, `\s`, `\w`:
				if index >= len(tokens) {
					ast = append(ast, Ast{
						Kind: PString,
						Str: t.str,
						Args: make([]string, 0),
					})
					return ast
				}
				t2 := tokens[index]
				index++
				if t2.kind == lNumber {
					s := strings.SplitN(t2.str, "", 2)
					ast = append(ast, Ast{
						Kind: PTag,
						Str: t.str,
						Args: []string{ s[0] },
					})
					if len(s) == 2 {
						ast = append(ast, Ast{
							Kind: PString,
							Str: s[1],
							Args: make([]string, 0),
						})
					}
				} else if t2.kind == lBraceOpen {
					i, s, ok := parseBrace(tokens[index:])
					if !ok {
						ast = append(ast, Ast{
							Kind: PString,
							Str: t.str,
							Args: make([]string, 0),
						})
						index--
					} else {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: s,
						})
						index += i
					}
				} else {
					ast = append(ast, Ast{
						Kind: PString,
						Str: t.str,
						Args: make([]string, 0),
					})
					index--
				}
			case `\__q`, `\_a`, `\_s`, `\_!`, `\_?`, `\c`, `\n`, `\x`:
				switch t.str {
				case `\__q`:
					if in_q {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: make([]string, 0),
						})
						in_q = false
						continue
					} else {
						if index >= len(tokens) {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							continue
						}
						t2 := tokens[index]
						if t2.kind != lBraceOpen {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							continue
						}
						index++
						i, s, ok := parseBrace(tokens[index:])
						if !ok {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							index--
						} else {
							in_q = true
							ast = append(ast, Ast{
								Kind: PTag,
								Str: t.str,
								Args: s,
							})
							index += i
						}
					}
				case `\_a`:
					if in_a {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: make([]string, 0),
						})
						in_a = false
						continue
					} else {
						if index >= len(tokens) {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							continue
						}
						t2 := tokens[index]
						if t2.kind != lBraceOpen {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							continue
						}
						index++
						i, s, ok := parseBrace(tokens[index:])
						if !ok {
							ast = append(ast, Ast{
								Kind: PString,
								Str: t.str,
								Args: make([]string, 0),
							})
							index--
						} else {
							in_a = true
							ast = append(ast, Ast{
								Kind: PTag,
								Str: t.str,
								Args: s,
							})
							index += i
						}
					}
				case `\_!`, `\_?`:
					in_raw = true
				default:
					if index >= len(tokens) {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: make([]string, 0),
						})
						continue
					}
					t2 := tokens[index]
					if t2.kind != lBraceOpen {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: make([]string, 0),
						})
						continue
					}
					index++
					i, s, ok := parseBrace(tokens[index:])
					if !ok {
						ast = append(ast, Ast{
							Kind: PString,
							Str: t.str,
							Args: make([]string, 0),
						})
						index--
					} else {
						ast = append(ast, Ast{
							Kind: PTag,
							Str: t.str,
							Args: s,
						})
						index += i
					}
				}
			default:
				t2 := tokens[index]
				if t2.kind != lBraceOpen {
					ast = append(ast, Ast{
						Kind: PString,
						Str: t.str,
						Args: make([]string, 0),
					})
					continue
				}
				index++
				i, s, ok := parseBrace(tokens[index:])
				if !ok {
					ast = append(ast, Ast{
						Kind: PString,
						Str: t.str,
						Args: make([]string, 0),
					})
					index--
				} else {
					ast = append(ast, Ast{
						Kind: PTag,
						Str: t.str,
						Args: s,
					})
					index += i
				}
			}
		case lEnv:
			ast = append(ast, Ast{
				Kind: PEnv,
				Str: t.str,
				Args: make([]string, 0),
			})
		default:
			ast = append(ast, Ast{
				Kind: PString,
				Str: t.str,
				Args: make([]string, 0),
			})
		}
	}
	return ast
}

func parseBrace(tokens []Token) (int, []string, bool) {
	args := make([]string, 0)
	arg := ""
	in_quote := false
	index := 0
	for {
		if (index >= len(tokens)) {
			return 0, args, false
		}
		token := tokens[index]
		index++
		switch token.kind {
		case lBraceClose:
			if in_quote {
				arg += token.str
			} else {
				args = append(args, arg)
				return index, args, true
			}
		case lComma:
			if in_quote {
				arg += token.str
			} else {
				args = append(args, arg)
				arg = ""
			}
		case lQuote:
			if in_quote {
				if index >= len(tokens) {
					return 0, args, false
				}
				token2 := tokens[index]
				if token2.kind == lQuote {
					arg += token2.str
					index++
				} else {
					in_quote = false
				}
			} else {
				in_quote = true
			}
		default:
			arg += token.str
		}
	}
	return 0, args, false
}
