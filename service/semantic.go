package service

import (
	"fmt"
	"strings"
	"examen-back/models"
)

type SemanticAnalyzer struct {
	tokens      []models.Token
	ast         *models.ASTNode
	symbolTable map[string]models.Symbol
	scopes      []string
	errors      []string
	warnings    []string
	checks      []models.SemanticCheck
}

func NewSemanticAnalyzer(tokens []models.Token, ast *models.ASTNode) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		tokens:      tokens,
		ast:         ast,
		symbolTable: make(map[string]models.Symbol),
		scopes:      []string{"global"},
	}
}

func (s *SemanticAnalyzer) Analyze() models.SemanticAnalysis {
	s.buildSymbolTable()
	
	s.checkFactorialFunction()
	s.checkVariableUsage()
	s.checkFunctionCalls()
	s.checkVariableScopes()
	
	symbolTableSlice := make([]models.Symbol, 0, len(s.symbolTable))
	for _, symbol := range s.symbolTable {
		symbolTableSlice = append(symbolTableSlice, symbol)
	}
	
	return models.SemanticAnalysis{
		Checks:      s.checks,
		SymbolTable: symbolTableSlice,
		Warnings:    s.warnings,
		Valid:       len(s.errors) == 0,
	}
}

func (s *SemanticAnalyzer) buildSymbolTable() {
	if s.ast == nil {
		return
	}
	
	s.analyzeNode(s.ast, "global")
}

func (s *SemanticAnalyzer) analyzeNode(node *models.ASTNode, scope string) {
	if node == nil {
		return
	}
	
	switch node.Type {
	case "FunctionDef":
		s.symbolTable[node.Value] = models.Symbol{
			Name:  node.Value,
			Type:  "function",
			Scope: scope,
			Line:  node.Line,
		}
		
		functionScope := scope + "." + node.Value
		
		for _, child := range node.Children {
			if child.Type == "Parameter" {
				s.symbolTable[child.Value] = models.Symbol{
					Name:  child.Value,
					Type:  "parameter",
					Scope: functionScope,
					Line:  child.Line,
				}
			} else {
				s.analyzeNode(child, functionScope)
			}
		}
		
	case "Assignment":
		s.symbolTable[node.Value] = models.Symbol{
			Name:  node.Value,
			Type:  "variable",
			Scope: scope,
			Line:  node.Line,
		}
		
		for _, child := range node.Children {
			s.analyzeNode(child, scope)
		}
		
	case "Identifier":
		if _, exists := s.symbolTable[node.Value]; !exists {
			builtins := map[string]bool{
				"print": true,
				"len":   true,
				"str":   true,
				"int":   true,
				"float": true,
			}
			
			if !builtins[node.Value] {
				s.warnings = append(s.warnings, fmt.Sprintf("Identificador '%s' usado sin definir en línea %d", node.Value, node.Line))
			}
		}
		
	default:
		for _, child := range node.Children {
			s.analyzeNode(child, scope)
		}
	}
}

func (s *SemanticAnalyzer) checkFactorialFunction() {
	factorialDefined := false
	factorialUsed := false
	factorialDefLine := 0
	var factorialUseLines []int
	
	if symbol, exists := s.symbolTable["factorial"]; exists && symbol.Type == "function" {
		factorialDefined = true
		factorialDefLine = symbol.Line
	}
	
	for _, token := range s.tokens {
		if token.Type == "IDENTIFIER" && token.Value == "factorial" {
			prevTokenIndex := -1
			for i, t := range s.tokens {
				if &t == &token {
					if i > 0 {
						prevTokenIndex = i - 1
					}
					break
				}
			}
			
			if prevTokenIndex >= 0 && s.tokens[prevTokenIndex].Value != "def" {
				factorialUsed = true
				factorialUseLines = append(factorialUseLines, token.Line)
			}
		}
	}
	
	definedBeforeUse := true
	if factorialUsed && factorialDefined {
		for _, useLine := range factorialUseLines {
			if useLine <= factorialDefLine {
				definedBeforeUse = false
				break
			}
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "factorial está definido antes de usarse",
		Passed:      !factorialUsed || (factorialDefined && definedBeforeUse),
		Line:        factorialDefLine,
	})
	
	if factorialUsed && !factorialDefined {
		s.errors = append(s.errors, "Función 'factorial' usada sin definir")
	}
}

func (s *SemanticAnalyzer) checkVariableUsage() {
	definedVars := make(map[string]int) 
	
	for _, token := range s.tokens {
		if token.Type == "IDENTIFIER" {
			isAssignment := false
			for i, t := range s.tokens {
				if &t == &token && i+1 < len(s.tokens) && s.tokens[i+1].Value == "=" {
					isAssignment = true
					definedVars[token.Value] = token.Line
					break
				}
			}
			
			if !isAssignment {
				if defLine, exists := definedVars[token.Value]; !exists {
					isParameter := false
					if symbol, symbolExists := s.symbolTable[token.Value]; symbolExists && symbol.Type == "parameter" {
						isParameter = true
					}
					
					builtins := map[string]bool{"print": true, "len": true, "str": true, "int": true, "float": true}
					isBuiltin := builtins[token.Value]
					
					if !isParameter && !isBuiltin {
						s.warnings = append(s.warnings, fmt.Sprintf("Variable '%s' usada sin definir en línea %d", token.Value, token.Line))
					}
				} else if defLine > token.Line {
					s.warnings = append(s.warnings, fmt.Sprintf("Variable '%s' usada antes de definir en línea %d", token.Value, token.Line))
				}
			}
		}
	}
	
	nUsed := false
	nDefined := false
	
	if symbol, exists := s.symbolTable["n"]; exists && symbol.Type == "parameter" {
		nDefined = true
	}
	
	for _, token := range s.tokens {
		if token.Type == "IDENTIFIER" && token.Value == "n" {
			isParamDef := false
			for i, t := range s.tokens {
				if &t == &token && i > 0 && s.tokens[i-1].Value == "(" {
					isParamDef = true
					break
				}
			}
			
			if !isParamDef {
				nUsed = true
				break
			}
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "El parámetro 'n' es utilizado correctamente",
		Passed:      !nDefined || nUsed,
	})
}

func (s *SemanticAnalyzer) checkFunctionCalls() {
	printHasArgs := false
	
	for i, token := range s.tokens {
		if token.Type == "IDENTIFIER" && token.Value == "print" {
			if i+1 < len(s.tokens) && s.tokens[i+1].Value == "(" {
				if i+2 < len(s.tokens) && s.tokens[i+2].Value != ")" {
					printHasArgs = true
				}
			}
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "print recibe argumentos válidos",
		Passed:      printHasArgs,
	})
	
	recursiveCallFound := false
	
	for i, token := range s.tokens {
		if token.Type == "IDENTIFIER" && token.Value == "factorial" {
			if i > 0 && s.tokens[i-1].Value != "def" {
				if i+1 < len(s.tokens) && s.tokens[i+1].Value == "(" {
					recursiveCallFound = true
				}
			}
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "Se detecta llamada recursiva correcta",
		Passed:      recursiveCallFound,
	})
}

func (s *SemanticAnalyzer) checkVariableScopes() {
	scopeConflicts := false
	conflictDetails := []string{}
	
	varsByName := make(map[string][]models.Symbol)
	for _, symbol := range s.symbolTable {
		if symbol.Type == "variable" || symbol.Type == "parameter" {
			varsByName[symbol.Name] = append(varsByName[symbol.Name], symbol)
		}
	}
	
	for name, symbols := range varsByName {
		if len(symbols) > 1 {
			// Verificar si están en scopes diferentes (esto es válido)
			scopes := make(map[string]bool)
			for _, symbol := range symbols {
				scopes[symbol.Scope] = true
			}
			
			// Si hay múltiples definiciones en el mismo scope, es un conflicto
			if len(scopes) == 1 && len(symbols) > 1 {
				scopeConflicts = true
				conflictDetails = append(conflictDetails, fmt.Sprintf("Variable '%s' definida múltiples veces en el mismo scope", name))
			}
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "Las variables tienen alcance correcto (sin colisiones)",
		Passed:      !scopeConflicts,
	})
	
	if scopeConflicts {
		s.warnings = append(s.warnings, conflictDetails...)
	}
	
	// Verificar específicamente que x y n no colisionen
	xSymbol, xExists := s.symbolTable["x"]
	nSymbol, nExists := s.symbolTable["n"]
	
	xnConflict := false
	if xExists && nExists {
		// x debería estar en scope global, n en scope de función
		if xSymbol.Scope == nSymbol.Scope {
			xnConflict = true
		}
	}
	
	s.checks = append(s.checks, models.SemanticCheck{
		Description: "Variables 'x' y 'n' no colisionan (scopes separados)",
		Passed:      !xnConflict,
	})
}

// Función auxiliar para verificar si una cadena contiene otra
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}