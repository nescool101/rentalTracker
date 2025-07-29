package controller

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

// FileUploadController maneja las operaciones de subida de archivos
type FileUploadController struct {
	userRepo   *storage.UserRepository
	personRepo *storage.PersonRepository
}

// NewFileUploadController crea un nuevo controlador de subida de archivos
func NewFileUploadController(userRepo *storage.UserRepository, personRepo *storage.PersonRepository) *FileUploadController {
	return &FileUploadController{
		userRepo:   userRepo,
		personRepo: personRepo,
	}
}

// UploadToken representa un token de subida
type UploadToken struct {
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UserID    string    `json:"user_id"`   // ID del usuario que subirá archivos
	PersonID  string    `json:"person_id"` // ID de la persona asociada
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedBy string    `json:"created_by"` // ID del admin/manager que creó el token
}

// Almacenamiento temporal de tokens (en producción usar base de datos)
var uploadTokens = make(map[string]*UploadToken)

// validateFileType valida el tipo de archivo permitido
func validateFileType(filename, contentType string) error {
	// Obtener extensión del archivo
	ext := strings.ToLower(filepath.Ext(filename))

	// Tipos de archivo permitidos
	allowedExtensions := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".txt":  true,
		".zip":  true,
		".rar":  true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	return nil
}

// GenerateUploadLinkRequest estructura para generar enlace de subida
type GenerateUploadLinkRequest struct {
	RecipientEmail string `json:"recipient_email" binding:"required,email"`
	RecipientName  string `json:"recipient_name" binding:"required"`
	UserID         string `json:"user_id" binding:"required"` // ID del usuario que subirá archivos
	ExpirationDays int    `json:"expiration_days"`
}

// UploadFileRequest estructura para subir archivo
type UploadFileRequest struct {
	Token      string `form:"token" binding:"required"`
	FolderName string `form:"folder_name"`
}

// RegisterRoutes registra las rutas de subida de archivos
func (ctrl *FileUploadController) RegisterRoutes(adminRouter *gin.RouterGroup) {
	uploadRoutes := adminRouter.Group("/file-upload")
	{
		// Solo admins y managers pueden generar enlaces
		uploadRoutes.POST("/generate-link", ctrl.HandleGenerateUploadLink)
		uploadRoutes.GET("/tokens", ctrl.HandleListUploadTokens)

		// Gestión de archivos subidos
		uploadRoutes.GET("/files", ctrl.HandleListUploadedFiles)
		uploadRoutes.GET("/files/:userID", ctrl.HandleListUserFiles)
		uploadRoutes.DELETE("/files/*filePath", ctrl.HandleDeleteFile)
		uploadRoutes.GET("/files/download/*filePath", ctrl.HandleDownloadFile)
		uploadRoutes.GET("/files/download-only/*filePath", ctrl.HandleDownloadFileOnly)
	}
}

// RegisterPublicRoutes registra rutas públicas (con token)
func (ctrl *FileUploadController) RegisterPublicRoutes(router *gin.RouterGroup) {
	publicRoutes := router.Group("/upload")
	{
		publicRoutes.POST("/file", ctrl.HandleUploadFileWithAuth)
		publicRoutes.GET("/validate-token/:token", ctrl.HandleValidateToken)
	}
}

// RegisterAuthenticatedUploadRoutes registra rutas autenticadas para subir archivos
func (ctrl *FileUploadController) RegisterAuthenticatedUploadRoutes(router *gin.RouterGroup) {
	authRoutes := router.Group("/upload")
	{
		authRoutes.POST("/file-authenticated", ctrl.HandleAuthenticatedUpload)
	}
}

// HandleGenerateUploadLink genera un enlace de subida para un usuario
// @Summary Generar enlace de subida
// @Description Genera un enlace temporal para que un usuario pueda subir archivos
// @Tags file-upload
// @Accept json
// @Produce json
// @Param request body GenerateUploadLinkRequest true "Datos del destinatario"
// @Success 200 {object} map[string]interface{}
// @Router /admin/file-upload/generate-link [post]
func (ctrl *FileUploadController) HandleGenerateUploadLink(ctx *gin.Context) {
	var req GenerateUploadLinkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin y manager pueden generar enlaces
	if authUser.Role != "admin" && authUser.Role != "manager" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores y managers pueden generar enlaces de subida"})
		return
	}

	// Verificar que el usuario destinatario existe
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	targetUser, err := ctrl.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		log.Printf("Error buscando usuario destinatario: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error validando usuario destinatario"})
		return
	}
	if targetUser == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Usuario destinatario no encontrado"})
		return
	}

	// Verificar que el email coincide con el usuario
	if targetUser.Email != req.RecipientEmail {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El email no coincide con el usuario especificado"})
		return
	}

	// Generar token único
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Error generando token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando token"})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Configurar expiración (por defecto 7 días)
	expirationDays := req.ExpirationDays
	if expirationDays <= 0 {
		expirationDays = 7
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, expirationDays)

	// Crear token de subida
	uploadToken := &UploadToken{
		Token:     token,
		Email:     req.RecipientEmail,
		Name:      req.RecipientName,
		UserID:    req.UserID,
		PersonID:  targetUser.PersonID.String(), // Convert UUID to string
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedBy: authUser.ID.String(), // Admin/Manager que creó el token
	}

	// Almacenar token (en producción usar base de datos)
	uploadTokens[token] = uploadToken

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Enlace de subida generado exitosamente",
		"token":       token,
		"expires_at":  expiresAt,
		"upload_link": fmt.Sprintf("/upload/file?token=%s", token),
	})
}

// HandleListUploadTokens lista todos los tokens de subida
// @Summary Listar tokens de subida
// @Description Lista todos los tokens de subida generados
// @Tags file-upload
// @Produce json
// @Success 200 {array} UploadToken
// @Router /admin/file-upload/tokens [get]
func (ctrl *FileUploadController) HandleListUploadTokens(ctx *gin.Context) {
	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin y manager pueden ver tokens
	if authUser.Role != "admin" && authUser.Role != "manager" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores y managers pueden ver tokens"})
		return
	}

	// Convertir mapa a slice
	tokens := make([]*UploadToken, 0, len(uploadTokens))
	for _, token := range uploadTokens {
		tokens = append(tokens, token)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"tokens":  tokens,
	})
}

// HandleValidateToken valida un token de subida
// @Summary Validar token de subida
// @Description Valida si un token de subida es válido y no ha expirado
// @Tags file-upload
// @Param token path string true "Token de subida"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /upload/validate-token/{token} [get]
func (ctrl *FileUploadController) HandleValidateToken(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token requerido"})
		return
	}

	uploadToken, exists := uploadTokens[token]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Token no válido"})
		return
	}

	// Verificar si el token ha expirado
	if time.Now().After(uploadToken.ExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}

	// Verificar si el token ya fue usado
	if uploadToken.Used {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token ya utilizado"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Token válido",
		"recipient":  uploadToken.Name,
		"expires_at": uploadToken.ExpiresAt,
	})
}

// HandleUploadFileWithAuth maneja la subida de archivos con token de autorización
// @Summary Subir archivo con token
// @Description Sube un archivo usando un token de autorización temporal
// @Tags file-upload
// @Accept multipart/form-data
// @Produce json
// @Param token formData string true "Token de autorización"
// @Param file formData file true "Archivo a subir"
// @Success 200 {object} service.SupabaseUploadResponse
// @Router /upload/file [post]
func (ctrl *FileUploadController) HandleUploadFileWithAuth(ctx *gin.Context) {
	// Obtener token del formulario
	token := ctx.PostForm("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token requerido"})
		return
	}

	// Validar token
	uploadToken, exists := uploadTokens[token]
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token no válido"})
		return
	}

	// Verificar si el token ha expirado
	if time.Now().After(uploadToken.ExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}

	// Verificar si el token ya fue usado
	if uploadToken.Used {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token ya utilizado"})
		return
	}

	// Obtener archivo del formulario
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Archivo requerido: " + err.Error()})
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	if err := validateFileType(header.Filename, header.Header.Get("Content-Type")); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Subir archivo usando Supabase Storage
	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	uploadResponse, err := supabaseStorage.UploadFile(file, header, uploadToken.UserID, uploadToken.Email)
	if err != nil {
		log.Printf("Error subiendo archivo: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo archivo"})
		return
	}

	// Marcar token como usado
	uploadToken.Used = true

	log.Printf("✅ Archivo subido con token: %s por %s", header.Filename, uploadToken.Email)

	ctx.JSON(http.StatusOK, uploadResponse)
}

// HandleAuthenticatedUpload maneja la subida de archivos con autenticación de usuario
// @Summary Subir archivo autenticado
// @Description Sube un archivo cuando el usuario está autenticado
// @Tags file-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Archivo a subir"
// @Success 200 {object} service.SupabaseUploadResponse
// @Router /upload/file-authenticated [post]
func (ctrl *FileUploadController) HandleAuthenticatedUpload(ctx *gin.Context) {
	// Verificar autenticación del usuario
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Debe iniciar sesión para subir archivos"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Obtener archivo del formulario
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Printf("Error obteniendo archivo del formulario: %v", err)
		// También intentar con otros nombres de campo comunes
		if file2, header2, err2 := ctx.Request.FormFile("files"); err2 == nil {
			file = file2
			header = header2
		} else if file3, header3, err3 := ctx.Request.FormFile("upload"); err3 == nil {
			file = file3
			header = header3
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Archivo requerido. Campo 'file' esperado: " + err.Error()})
			return
		}
	}
	defer file.Close()

	// Validar tipo de archivo
	if err := validateFileType(header.Filename, header.Header.Get("Content-Type")); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Subir archivo usando Supabase Storage
	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	uploadResponse, err := supabaseStorage.UploadFile(file, header, authUser.ID.String(), authUser.Email)
	if err != nil {
		log.Printf("Error subiendo archivo: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error subiendo archivo"})
		return
	}

	log.Printf("✅ Archivo subido autenticado: %s por usuario %s", header.Filename, authUser.Email)

	ctx.JSON(http.StatusOK, uploadResponse)
}

// HandleListUploadedFiles lista todos los archivos subidos para administradores
// @Summary Listar archivos subidos
// @Description Lista todos los archivos subidos por usuarios
// @Tags file-upload
// @Produce json
// @Success 200 {array} service.SupabaseFileInfo
// @Router /admin/file-upload/files [get]
func (ctrl *FileUploadController) HandleListUploadedFiles(ctx *gin.Context) {
	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin y manager pueden ver archivos
	if authUser.Role != "admin" && authUser.Role != "manager" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores y managers pueden ver archivos"})
		return
	}

	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	files, err := supabaseStorage.ListAllFiles()
	if err != nil {
		log.Printf("Error listando archivos: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo archivos"})
		return
	}

	log.Printf("📋 CONTROLLER: Enviando %d archivos al frontend", len(files))
	for i, file := range files {
		log.Printf("📄 CONTROLLER File %d: Name='%s', Path='%s'", i+1, file.Name, file.Path)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"files":   files,
	})
}

// HandleListUserFiles lista archivos de un usuario específico
// @Summary Listar archivos de un usuario
// @Description Lista todos los archivos subidos por un usuario específico
// @Tags file-upload
// @Param userID path string true "ID del usuario"
// @Produce json
// @Success 200 {array} service.SupabaseFileInfo
// @Router /admin/file-upload/files/{userID} [get]
func (ctrl *FileUploadController) HandleListUserFiles(ctx *gin.Context) {
	userID := ctx.Param("userID")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario requerido"})
		return
	}

	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin y manager pueden ver archivos de otros usuarios
	if authUser.Role != "admin" && authUser.Role != "manager" && authUser.ID.String() != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo puede ver sus propios archivos"})
		return
	}

	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	files, err := supabaseStorage.ListUserFiles(userID)
	if err != nil {
		log.Printf("Error listando archivos del usuario: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo archivos del usuario"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"files":   files,
	})
}

// HandleDownloadFile descarga un archivo y lo elimina (para admins)
// @Summary Descargar archivo
// @Description Descarga un archivo y lo elimina después de la descarga (solo admins)
// @Tags file-upload
// @Param filePath path string true "Ruta del archivo"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Router /admin/file-upload/files/download/{filePath} [get]
func (ctrl *FileUploadController) HandleDownloadFile(ctx *gin.Context) {
	filePath := ctx.Param("filePath")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ruta de archivo requerida"})
		return
	}

	// Remove leading slash from wildcard parameter
	filePath = strings.TrimPrefix(filePath, "/")

	log.Printf("🔍 HandleDownloadFile - Intentando descargar archivo con ruta: '%s'", filePath)

	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin puede descargar y eliminar archivos
	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores pueden descargar archivos"})
		return
	}

	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	// Descargar y eliminar archivo automáticamente
	fileData, err := supabaseStorage.DownloadAndDeleteFile(filePath)
	if err != nil {
		log.Printf("Error descargando archivo: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error descargando archivo"})
		return
	}

	// Servir archivo
	fileName := filepath.Base(filePath)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Data(http.StatusOK, "application/octet-stream", fileData)

	log.Printf("✅ Archivo descargado y eliminado: %s por admin %s", filePath, authUser.Email)
}

// HandleDownloadFileOnly descarga un archivo SIN eliminarlo (para admins)
// @Summary Descargar archivo solamente
// @Description Descarga un archivo sin eliminarlo (solo admins)
// @Tags file-upload
// @Param filePath path string true "Ruta del archivo"
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Router /admin/file-upload/files/download-only/{filePath} [get]
func (ctrl *FileUploadController) HandleDownloadFileOnly(ctx *gin.Context) {
	filePath := ctx.Param("filePath")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ruta de archivo requerida"})
		return
	}

	// Remove leading slash from wildcard parameter
	filePath = strings.TrimPrefix(filePath, "/")

	log.Printf("🔍 HandleDownloadFileOnly - Intentando descargar archivo con ruta: '%s'", filePath)

	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin puede descargar archivos
	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores pueden descargar archivos"})
		return
	}

	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	// Solo descargar archivo (SIN eliminar)
	fileData, err := supabaseStorage.DownloadFile(filePath)
	if err != nil {
		log.Printf("Error descargando archivo: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error descargando archivo"})
		return
	}

	// Servir archivo
	fileName := filepath.Base(filePath)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Data(http.StatusOK, "application/octet-stream", fileData)

	log.Printf("✅ Archivo descargado (sin eliminar): %s por admin %s", filePath, authUser.Email)
}

// HandleDeleteFile elimina un archivo específico
// @Summary Eliminar archivo
// @Description Elimina un archivo específico (solo admins)
// @Tags file-upload
// @Param filePath path string true "Ruta del archivo"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/file-upload/files/{filePath} [delete]
func (ctrl *FileUploadController) HandleDeleteFile(ctx *gin.Context) {
	filePath := ctx.Param("filePath")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ruta de archivo requerida"})
		return
	}

	// Remove leading slash from wildcard parameter
	filePath = strings.TrimPrefix(filePath, "/")

	// Verificar autenticación y permisos
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Datos de usuario inválidos"})
		return
	}

	// Solo admin puede eliminar archivos
	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Solo administradores pueden eliminar archivos"})
		return
	}

	supabaseStorage := service.GetSupabaseStorageService()
	if supabaseStorage == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Servicio de archivos no disponible"})
		return
	}

	err := supabaseStorage.DeleteFile(filePath)
	if err != nil {
		log.Printf("Error eliminando archivo: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando archivo"})
		return
	}

	log.Printf("✅ Archivo eliminado: %s por admin %s", filePath, authUser.Email)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Archivo eliminado exitosamente",
	})
}
