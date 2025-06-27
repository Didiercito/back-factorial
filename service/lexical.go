package service

import (
	"unicode"
	"examen-back/models"
)

type LexicalAnalyzer struct {
	input    string
	position int
	line     int
	col      int
	tokens   []models.Token
}

func NewLexicalAnalyzer(input string) *LexicalAnalyzer {
	return &LexicalAnalyzer{
		input: input,
		line:  1,
		col:   1,
	}
}

var keywords = map[string]bool{
	"def":    true,
	"if":     true,
	"else":   true,
	"elif":   true,
	"return": true,
	"print":  true,
	"for":    true,
	"while":  true,
	"in":     true,
	"and":    true,
	"or":     true,
	"not":    true,
	"True":   true,
	"False":  true,
	"None":   true,
	"import": true,
	"from":   true,
	"as":     true,
	"try":    true,
	"except": true,
	"finally": true,
	"with":   true,
	"class":  true,
	"pass":   true,
	"break":  true,
	"continue": true,
}

func (l *LexicalAnalyzer) peek() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position])
}

func (l *LexicalAnalyzer) advance() rune {
	if l.position >= len(l.input) {
		return 0
	}
	ch := rune(l.input[l.position])
	l.position++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *LexicalAnalyzer) peekNext() rune {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position+1])
}

func (l *LexicalAnalyzer) addToken(tokenType, value string) {
	l.tokens = append(l.tokens, models.Token{
		Type:  tokenType,
		Value: value,
		Line:  l.line,
		Col:   l.col - len(value),
	})
}

func (l *LexicalAnalyzer) skipWhitespace() {
	for l.position < len(l.input) && unicode.IsSpace(l.peek()) && l.peek() != '\n' {
		l.advance()
	}
}

func (l *LexicalAnalyzer) readIdentifier() string {
	start := l.position
	for l.position < len(l.input) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	return l.input[start:l.position]
}

func (l *LexicalAnalyzer) readNumber() string {
	start := l.position
	hasDecimal := false
	
	for l.position < len(l.input) {
		ch := l.peek()
		if unicode.IsDigit(ch) {
			l.advance()
		} else if ch == '.' && !hasDecimal {
			hasDecimal = true
			l.advance()
		} else {
			break
		}
	}
	return l.input[start:l.position]
}

func (l *LexicalAnalyzer) readString(quote rune) string {
	l.advance() 
	start := l.position
	
	for l.position < len(l.input) && l.peek() != quote {
		if l.peek() == '\\' {
			l.advance() 
			if l.position < len(l.input) {
				l.advance() 
			}
		} else {
			l.advance()
		}
	}
	
	value := l.input[start:l.position]
	if l.position < len(l.input) {
		l.advance() // Skip closing quote
	}
	
	return value
}

func (l *LexicalAnalyzer) Tokenize() models.LexicalAnalysis {
	for l.position < len(l.input) {
		l.skipWhitespace()
		
		if l.position >= len(l.input) {
			break
		}
		
		ch := l.peek()
		
		switch {
		case ch == '\n':
			l.advance()
			l.addToken("NEWLINE", "\\n")
			
		case ch == '#':
			// Comentario - leer hasta el final de la línea
			start := l.position
			for l.position < len(l.input) && l.peek() != '\n' {
				l.advance()
			}
			comment := l.input[start:l.position]
			l.addToken("COMMENT", comment)
			
		case unicode.IsLetter(ch) || ch == '_':
			identifier := l.readIdentifier()
			if keywords[identifier] {
				l.addToken("KEYWORD", identifier)
			} else {
				l.addToken("IDENTIFIER", identifier)
			}
			
		case unicode.IsDigit(ch):
			number := l.readNumber()
			l.addToken("NUMBER", number)
			
		case ch == '"' || ch == '\'':
			str := l.readString(ch)
			l.addToken("STRING", str)
			
		case ch == '(':
			l.advance()
			l.addToken("SYMBOL", "(")
			
		case ch == ')':
			l.advance()
			l.addToken("SYMBOL", ")")
			
		case ch == '[':
			l.advance()
			l.addToken("SYMBOL", "[")
			
		case ch == ']':
			l.advance()
			l.addToken("SYMBOL", "]")
			
		case ch == '{':
			l.advance()
			l.addToken("SYMBOL", "{")
			
		case ch == '}':
			l.advance()
			l.addToken("SYMBOL", "}")
			
		case ch == ':':
			l.advance()
			l.addToken("SYMBOL", ":")
			
		case ch == ';':
			l.advance()
			l.addToken("SYMBOL", ";")
			
		case ch == ',':
			l.advance()
			l.addToken("SYMBOL", ",")
			
		case ch == '.':
			l.advance()
			l.addToken("SYMBOL", ".")
			
		case ch == '+':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "+=")
			} else {
				l.addToken("OPERATOR", "+")
			}
			
		case ch == '-':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "-=")
			} else {
				l.addToken("OPERATOR", "-")
			}
			
		case ch == '*':
			l.advance()
			if l.peek() == '*' {
				l.advance()
				if l.peek() == '=' {
					l.advance()
					l.addToken("OPERATOR", "**=")
				} else {
					l.addToken("OPERATOR", "**")
				}
			} else if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "*=")
			} else {
				l.addToken("OPERATOR", "*")
			}
			
		case ch == '/':
			l.advance()
			if l.peek() == '/' {
				l.advance()
				if l.peek() == '=' {
					l.advance()
					l.addToken("OPERATOR", "//=")
				} else {
					l.addToken("OPERATOR", "//")
				}
			} else if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "/=")
			} else {
				l.addToken("OPERATOR", "/")
			}
			
		case ch == '%':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "%=")
			} else {
				l.addToken("OPERATOR", "%")
			}
			
		case ch == '=':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "==")
			} else {
				l.addToken("OPERATOR", "=")
			}
			
		case ch == '!':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "!=")
			} else {
				l.addToken("ERROR", "!")
			}
			
		case ch == '<':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", "<=")
			} else if l.peek() == '<' {
				l.advance()
				l.addToken("OPERATOR", "<<")
			} else {
				l.addToken("OPERATOR", "<")
			}
			
		case ch == '>':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				l.addToken("OPERATOR", ">=")
			} else if l.peek() == '>' {
				l.advance()
				l.addToken("OPERATOR", ">>")
			} else {
				l.addToken("OPERATOR", ">")
			}
			
		case ch == '&':
			l.advance()
			l.addToken("OPERATOR", "&")
			
		case ch == '|':
			l.advance()
			l.addToken("OPERATOR", "|")
			
		case ch == '^':
			l.advance()
			l.addToken("OPERATOR", "^")
			
		case ch == '~':
			l.advance()
			l.addToken("OPERATOR", "~")
			
		default:
			l.advance()
			l.addToken("ERROR", string(ch))
		}
	}
	
	// Calcular estadísticas
	stats := make(map[string]int)
	for _, token := range l.tokens {
		stats[token.Type]++
	}
	
	return models.LexicalAnalysis{
		Tokens: l.tokens,
		Stats:  stats,
	}
}