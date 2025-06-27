package models

type Token struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Line  int    `json:"line"`
	Col   int    `json:"col"`
}

type LexicalAnalysis struct {
	Tokens []Token        `json:"tokens"`
	Stats  map[string]int `json:"stats"`
}

type SyntaxAnalysis struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
	AST    *ASTNode `json:"ast,omitempty"`
}

type SemanticAnalysis struct {
	Checks      []SemanticCheck `json:"checks"`
	SymbolTable []Symbol        `json:"symbolTable"`
	Warnings    []string        `json:"warnings"`
	Valid       bool            `json:"valid"`
}

type SemanticCheck struct {
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Line        int    `json:"line,omitempty"`
}

type Symbol struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Scope string `json:"scope"`
	Line  int    `json:"line"`
}

type ASTNode struct {
	Type     string     `json:"type"`
	Value    string     `json:"value,omitempty"`
	Children []*ASTNode `json:"children,omitempty"`
	Line     int        `json:"line"`
}

type AnalysisRequest struct {
	Code string `json:"code" binding:"required"`
}

type AnalysisResponse struct {
	Lexical  LexicalAnalysis  `json:"lexical"`
	Syntax   SyntaxAnalysis   `json:"syntax"`
	Semantic SemanticAnalysis `json:"semantic"`
	Success  bool             `json:"success"`
	Message  string           `json:"message,omitempty"`
}