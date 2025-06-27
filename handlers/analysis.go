package handlers

import (
	"net/http"
	"examen-back/models"
	"examen-back/service"
	
	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct{}

func NewAnalysisHandler() *AnalysisHandler {
	return &AnalysisHandler{}
}

func (h *AnalysisHandler) AnalyzeCode(c *gin.Context) {
	var request models.AnalysisRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "JSON inválido: " + err.Error(),
		})
		return
	}
	
	if request.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "El código no puede estar vacío",
		})
		return
	}
	
	lexicalAnalyzer := service.NewLexicalAnalyzer(request.Code)
	lexicalResult := lexicalAnalyzer.Tokenize()
	
	syntaxAnalyzer := service.NewSyntaxAnalyzer(lexicalResult.Tokens)
	syntaxResult := syntaxAnalyzer.Analyze()
	
	semanticAnalyzer := service.NewSemanticAnalyzer(lexicalResult.Tokens, syntaxResult.AST)
	semanticResult := semanticAnalyzer.Analyze()
	
	response := models.AnalysisResponse{
		Lexical:  lexicalResult,
		Syntax:   syntaxResult,
		Semantic: semanticResult,
		Success:  true,
		Message:  "Análisis completado exitosamente",
	}
	
	if !syntaxResult.Valid || !semanticResult.Valid {
		response.Success = false
		response.Message = "Análisis completado con errores"
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *AnalysisHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "factorial-analyzer",
		"version": "1.0.0",
	})
}

func (h *AnalysisHandler) GetAnalysisInfo(c *gin.Context) {
	info := gin.H{
		"name":        "Analizador Léxico, Sintáctico y Semántico",
		"description": "Sistema de análisis completo para código Python",
		"capabilities": []string{
			"Análisis léxico con reconocimiento de tokens",
			"Análisis sintáctico con construcción de AST",
			"Análisis semántico con tabla de símbolos",
			"Verificaciones específicas para funciones factorial",
		},
		"supported_constructs": []string{
			"Definición de funciones",
			"Estructuras condicionales (if-else)",
			"Llamadas recursivas",
			"Expresiones aritméticas",
			"Asignación de variables",
			"Llamadas a funciones",
		},
		"token_types": []string{
			"KEYWORD", "IDENTIFIER", "NUMBER", "STRING",
			"OPERATOR", "SYMBOL", "COMMENT", "NEWLINE", "ERROR",
		},
	}
	
	c.JSON(http.StatusOK, info)
}