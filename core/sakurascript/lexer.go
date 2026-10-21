package sakurascript

import (
	"regexp"
)

const lString = 0
const lNumber = 1
const lBraceOpen = 2
const lBraceClose = 3
const lTag = 4
const lEnv = 5
const lQuote = 6
const lComma = 7

var rMeta = regexp.MustCompile(`[-+\[\]",0-9\\%]`)
var rNumber = regexp.MustCompile(`^[+-]?\d+(\.\d+)?`)
var rTag = regexp.MustCompile(`^\\([-!&*+0145678Cabcefhijmnpqstuvwxz%]|_([!+?Vablmnqsuvw]|_[cqtvw]))`)
var rEnv = regexp.MustCompile(`^%(month|day|hour|minute|second|username|selfname2?|keroname|screenwidth|screenheight|exh|et|wronghour|ms|mz|ml|mc|mh|mt|me|mp|m\?|dms|lastghostname|lastobjectname|\*|property)`)

type Token struct {
	kind int
	str string
}

func Lex(str string) []Token {
	data := make([]Token, 0)
	for {
		if str == "" {
			break
		}
		switch c := string(str[0]); c {
		case "[":
			data = append(data, Token{ kind: lBraceOpen, str: c })
			str = str[1:]
		case "]":
			data = append(data, Token{ kind: lBraceClose, str: c })
			str = str[1:]
		case `"`:
			data = append(data, Token{ kind: lQuote, str: c })
			str = str[1:]
		case ",":
			data = append(data, Token{ kind: lComma, str: c })
			str = str[1:]
		case "+", "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			match := rNumber.FindString(str)
			if match != "" {
				data = append(data, Token{ kind: lNumber, str: match })
				str = str[len(match):]
				continue
			}
			next_meta := rMeta.FindStringIndex(str[1:])
			if next_meta == nil {
				data = append(data, Token{ kind: lString, str: str })
				return data
			}
			n := next_meta[0] + 1
			data = append(data, Token{ kind: lString, str: str[0:n + 1] })
			str = str[n + 1:]
		case `\`:
			match := rTag.FindString(str)
			if match != "" {
				data = append(data, Token{ kind: lTag, str: match })
				str = str[len(match):]
				continue
			}
			next_meta := rMeta.FindStringIndex(str[1:])
			if next_meta == nil {
				data = append(data, Token{ kind: lString, str: str })
				return data
			}
			n := next_meta[0] + 1
			data = append(data, Token{ kind: lString, str: str[0:n + 1] })
			str = str[n + 1:]
		case "%":
			match := rEnv.FindString(str)
			if match != "" {
				data = append(data, Token{ kind: lEnv, str: match })
				str = str[len(match):]
				continue
			}
			next_meta := rMeta.FindStringIndex(str[1:])
			if next_meta == nil {
				data = append(data, Token{ kind: lString, str: str })
				return data
			}
			n := next_meta[0] + 1
			data = append(data, Token{ kind: lString, str: str[0:n + 1] })
			str = str[n + 1:]
		default:
			next_meta := rMeta.FindStringIndex(str)
			if next_meta == nil {
				data = append(data, Token{ kind: lString, str: str })
				return data
			}
			n := next_meta[0]
			data = append(data, Token{ kind: lString, str: str[0:n] })
			str = str[n:]
		}
	}
	return data
}
