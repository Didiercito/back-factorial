package service

import (
	"fmt"
	"examen-back/models"
)

type SyntaxAnalyzer struct {
	tokens   []models.Token
	position int
	errors   []string
}

func NewSyntaxAnalyzer(tokens []models.Token) *SyntaxAnalyzer {
	// Filtrar tokens de nueva línea y comentarios para el análisis sintáctico
	filteredTokens := make([]models.Token, 0)
	for _, token := range tokens {
		if token.Type != "NEWLINE" && token.Type != "COMMENT" {
			filteredTokens = append(filteredTokens, token)
		}
	}
	
	return &SyntaxAnalyzer{
		tokens: filteredTokens,
	}
}

func (s *SyntaxAnalyzer) peek() *models.Token {
	if s.position >= len(s.tokens) {
		return nil
	}
	return &s.tokens[s.position]
}

func (s *SyntaxAnalyzer) advance() *models.Token {
	if s.position >= len(s.tokens) {
		return nil
	}
	token := &s.tokens[s.position]
	s.position++
	return token
}

func (s *SyntaxAnalyzer) expect(tokenType, value string) bool {
	token := s.peek()
	if token == nil {
		s.errors = append(s.errors, fmt.Sprintf("Se esperaba '%s' pero se encontró el final del archivo", value))
		return false
	}
	
	if token.Type != tokenType || (value != "" && token.Value != value) {
		s.errors = append(s.errors, fmt.Sprintf("Se esperaba '%s' en línea %d, pero se encontró '%s'", value, token.Line, token.Value))
		return false
	}
	
	s.advance()
	return true
}

func (s *SyntaxAnalyzer) parseExpression() *models.ASTNode {
	return s.parseComparison()
}

func (s *SyntaxAnalyzer) parseComparison() *models.ASTNode {
	left := s.parseArithmetic()
	
	token := s.peek()
	if token != nil && (token.Value == "<" || token.Value == "<=" || token.Value == ">" || token.Value == ">=" || token.Value == "==" || token.Value == "!=") {
		op := s.advance()
		right := s.parseArithmetic()
		
		return &models.ASTNode{
			Type:     "BinaryOp",
			Value:    op.Value,
			Children: []*models.ASTNode{left, right},
			Line:     op.Line,
		}
	}
	
	return left
}

func (s *SyntaxAnalyzer) parseArithmetic() *models.ASTNode {
	left := s.parseTerm()
	
	for {
		token := s.peek()
		if token == nil || (token.Value != "+" && token.Value != "-") {
			break
		}
		
		op := s.advance()
		right := s.parseTerm()
		
		left = &models.ASTNode{
			Type:     "BinaryOp",
			Value:    op.Value,
			Children: []*models.ASTNode{left, right},
			Line:     op.Line,
		}
	}
	
	return left
}

func (s *SyntaxAnalyzer) parseTerm() *models.ASTNode {
	left := s.parseFactor()
	
	for {
		token := s.peek()
		if token == nil || (token.Value != "*" && token.Value != "/" && token.Value != "%") {
			break
		}
		
		op := s.advance()
		right := s.parseFactor()
		
		left = &models.ASTNode{
			Type:     "BinaryOp",
			Value:    op.Value,
			Children: []*models.ASTNode{left, right},
			Line:     op.Line,
		}
	}
	
	return left
}

func (s *SyntaxAnalyzer) parseFactor() *models.ASTNode {
	token := s.peek()
	if token == nil {
		s.errors = append(s.errors, "Se esperaba una expresión")
		return nil
	}
	
	switch token.Type {
	case "NUMBER":
		s.advance()
		return &models.ASTNode{
			Type:  "Number",
			Value: token.Value,
			Line:  token.Line,
		}
		
	case "STRING":
		s.advance()
		return &models.ASTNode{
			Type:  "String",
			Value: token.Value,
			Line:  token.Line,
		}
		
	case "IDENTIFIER":
		identifier := s.advance()
		
		// Verificar si es una llamada a función
		if s.peek() != nil && s.peek().Value == "(" {
			s.advance() // consume '('
			
			args := make([]*models.ASTNode, 0)
			
			// Parsear argumentos
			if s.peek() != nil && s.peek().Value != ")" {
				args = append(args, s.parseExpression())
				
				for s.peek() != nil && s.peek().Value == "," {
					s.advance() // consume ','
					args = append(args, s.parseExpression())
				}
			}
			
			if !s.expect("SYMBOL", ")") {
				return nil
			}
			
			node := &models.ASTNode{
				Type:     "FunctionCall",
				Value:    identifier.Value,
				Children: args,
				Line:     identifier.Line,
			}
			
			return node
		}
		
		return &models.ASTNode{
			Type:  "Identifier",
			Value: identifier.Value,
			Line:  identifier.Line,
		}
		
	case "SYMBOL":
		if token.Value == "(" {
			s.advance() // consume '('
			expr := s.parseExpression()
			if !s.expect("SYMBOL", ")") {
				return nil
			}
			return expr
		}
	}
	
	s.errors = append(s.errors, fmt.Sprintf("Token inesperado '%s' en línea %d", token.Value, token.Line))
	return nil
}

func (s *SyntaxAnalyzer) parseStatement() *models.ASTNode {
	token := s.peek()
	if token == nil {
		return nil
	}
	
	switch {
	case token.Type == "KEYWORD" && token.Value == "def":
		return s.parseFunctionDef()
		
	case token.Type == "KEYWORD" && token.Value == "if":
		return s.parseIfStatement()
		
	case token.Type == "KEYWORD" && token.Value == "return":
		return s.parseReturnStatement()
		
	case token.Type == "IDENTIFIER":
		// Puede ser asignación o llamada a función
		identifier := s.advance()
		
		if s.peek() != nil && s.peek().Value == "=" {
			s.advance() // consume '='
			value := s.parseExpression()
			
			return &models.ASTNode{
				Type:  "Assignment",
				Value: identifier.Value,
				Children: []*models.ASTNode{value},
				Line:  identifier.Line,
			}
		} else {
			// Es una expresión (probablemente llamada a función)
			s.position-- // retroceder
			return s.parseExpression()
		}
		
	default:
		return s.parseExpression()
	}
}

func (s *SyntaxAnalyzer) parseFunctionDef() *models.ASTNode {
	if !s.expect("KEYWORD", "def") {
		return nil
	}
	
	nameToken := s.peek()
	if !s.expect("IDENTIFIER", "") {
		return nil
	}
	
	if !s.expect("SYMBOL", "(") {
		return nil
	}
	
	params := make([]*models.ASTNode, 0)
	
	// Parsear parámetros
	if s.peek() != nil && s.peek().Value != ")" {
		if s.expect("IDENTIFIER", "") {
			params = append(params, &models.ASTNode{
				Type:  "Parameter",
				Value: s.tokens[s.position-1].Value,
				Line:  s.tokens[s.position-1].Line,
			})
		}
		
		for s.peek() != nil && s.peek().Value == "," {
			s.advance() // consume ','
			if s.expect("IDENTIFIER", "") {
				params = append(params, &models.ASTNode{
					Type:  "Parameter",
					Value: s.tokens[s.position-1].Value,
					Line:  s.tokens[s.position-1].Line,
				})
			}
		}
	}
	
	if !s.expect("SYMBOL", ")") {
		return nil
	}
	
	if !s.expect("SYMBOL", ":") {
		return nil
	}
	
	// Parsear cuerpo de la función (simplificado)
	body := make([]*models.ASTNode, 0)
	
	// En un parser real aquí manejaríamos la indentación
	// Por simplicidad, parseamos hasta encontrar otra función o el final
	for s.peek() != nil && !(s.peek().Type == "KEYWORD" && s.peek().Value == "def") {
		stmt := s.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			break
		}
	}
	
	children := append(params, body...)
	
	return &models.ASTNode{
		Type:     "FunctionDef",
		Value:    nameToken.Value,
		Children: children,
		Line:     nameToken.Line,
	}
}

func (s *SyntaxAnalyzer) parseIfStatement() *models.ASTNode {
	ifToken := s.advance() // consume 'if'
	
	condition := s.parseExpression()
	if condition == nil {
		return nil
	}
	
	if !s.expect("SYMBOL", ":") {
		return nil
	}
	
	// Parsear cuerpo del if
	thenBody := make([]*models.ASTNode, 0)
	
	// Simplificado: parseamos hasta encontrar 'else' o el final
	for s.peek() != nil && !(s.peek().Type == "KEYWORD" && s.peek().Value == "else") {
		stmt := s.parseStatement()
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		} else {
			break
		}
	}
	
	children := []*models.ASTNode{condition}
	children = append(children, thenBody...)
	
	// Verificar si hay else
	if s.peek() != nil && s.peek().Type == "KEYWORD" && s.peek().Value == "else" {
		s.advance() // consume 'else'
		if !s.expect("SYMBOL", ":") {
			return nil
		}
		
		// Parsear cuerpo del else
		elseBody := make([]*models.ASTNode, 0)
		for s.peek() != nil {
			stmt := s.parseStatement()
			if stmt != nil {
				elseBody = append(elseBody, stmt)
			} else {
				break
			}
		}
		
		children = append(children, elseBody...)
	}
	
	return &models.ASTNode{
		Type:     "IfStatement",
		Children: children,
		Line:     ifToken.Line,
	}
}

func (s *SyntaxAnalyzer) parseReturnStatement() *models.ASTNode {
	returnToken := s.advance() // consume 'return'
	
	var value *models.ASTNode
	if s.peek() != nil && s.peek().Type != "KEYWORD" {
		value = s.parseExpression()
	}
	
	children := make([]*models.ASTNode, 0)
	if value != nil {
		children = append(children, value)
	}
	
	return &models.ASTNode{
		Type:     "ReturnStatement",
		Children: children,
		Line:     returnToken.Line,
	}
}

func (s *SyntaxAnalyzer) Analyze() models.SyntaxAnalysis {
	if len(s.tokens) == 0 {
		return models.SyntaxAnalysis{
			Valid:  false,
			Errors: []string{"No hay tokens para analizar"},
		}
	}
	
	// Verificaciones básicas
	s.checkBasicSyntax()
	
	// Construir AST
	statements := make([]*models.ASTNode, 0)
	
	for s.position < len(s.tokens) {
		stmt := s.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
		
		// Evitar bucle infinito
		if s.position >= len(s.tokens) {
			break
		}
	}
	
	ast := &models.ASTNode{
		Type:     "Program",
		Children: statements,
		Line:     1,
	}
	
	return models.SyntaxAnalysis{
		Valid:  len(s.errors) == 0,
		Errors: s.errors,
		AST:    ast,
	}
}

func (s *SyntaxAnalyzer) checkBasicSyntax() {
	// Verificar paréntesis balanceados
	s.checkBalancedParentheses()
	
	// Verificar estructura básica de función
	s.checkFunctionStructure()
}

func (s *SyntaxAnalyzer) checkBalancedParentheses() {
	stack := make([]string, 0)
	pairs := map[string]string{
		")": "(",
		"]": "[",
		"}": "{",
	}
	
	for _, token := range s.tokens {
		if token.Type == "SYMBOL" {
			switch token.Value {
			case "(", "[", "{":
				stack = append(stack, token.Value)
			case ")", "]", "}":
				if len(stack) == 0 {
					s.errors = append(s.errors, fmt.Sprintf("Paréntesis/corchete de cierre sin apertura en línea %d", token.Line))
					return
				}
				expected := pairs[token.Value]
				if stack[len(stack)-1] != expected {
					s.errors = append(s.errors, fmt.Sprintf("Paréntesis/corchetes no coinciden en línea %d", token.Line))
					return
				}
				stack = stack[:len(stack)-1]
			}
		}
	}
	
	if len(stack) > 0 {
		s.errors = append(s.errors, "Paréntesis/corchetes sin cerrar")
	}
}

func (s *SyntaxAnalyzer) checkFunctionStructure() {
	for i, token := range s.tokens {
		if token.Type == "KEYWORD" && token.Value == "def" {
			// Verificar que después de 'def' venga un identificador
			if i+1 >= len(s.tokens) || s.tokens[i+1].Type != "IDENTIFIER" {
				s.errors = append(s.errors, fmt.Sprintf("Se esperaba nombre de función después de 'def' en línea %d", token.Line))
				continue
			}
			
			// Verificar que después del nombre venga '('
			if i+2 >= len(s.tokens) || s.tokens[i+2].Value != "(" {
				s.errors = append(s.errors, fmt.Sprintf("Se esperaba '(' después del nombre de función en línea %d", token.Line))
				continue
			}
			
			// Buscar el ':' que debe estar antes del cuerpo
			foundColon := false
			for j := i + 2; j < len(s.tokens); j++ {
				if s.tokens[j].Value == ":" {
					foundColon = true
					break
				}
				if s.tokens[j].Type == "KEYWORD" && s.tokens[j].Value == "def" {
					break // Nueva función encontrada
				}
			}
			
			if !foundColon {
				s.errors = append(s.errors, fmt.Sprintf("Se esperaba ':' después de la definición de función en línea %d", token.Line))
			}
		}
	}
}