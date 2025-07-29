package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"
)

// SupabaseStorageService maneja almacenamiento de archivos en Supabase
type SupabaseStorageService struct {
	client     *storage_go.Client
	bucketName string
	projectURL string
}

// SupabaseUploadResponse respuesta de subida a Supabase Storage
type SupabaseUploadResponse struct {
	Success    bool   `json:"success"`
	Key        string `json:"key"`  // Nombre del archivo en Supabase
	Link       string `json:"link"` // URL pública del archivo
	Name       string `json:"name"` // Nombre original del archivo
	Path       string `json:"path"` // Ruta completa en el bucket
	Size       int64  `json:"size"` // Tamaño del archivo
	UploadedBy string `json:"uploaded_by"`
	UploadedAt string `json:"uploaded_at"`
	BucketName string `json:"bucket_name"`
}

// SupabaseFileInfo información de un archivo en Supabase
type SupabaseFileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	MimeType    string `json:"mime_type"`
	UploadedAt  string `json:"uploaded_at"`
	DownloadURL string `json:"download_url"`
}

var supabaseStorageService *SupabaseStorageService

// InitializeSupabaseStorageService inicializa el servicio de Supabase Storage
func InitializeSupabaseStorageService() error {
	// Obtener configuración desde variables de entorno
	projectURL := os.Getenv("SUPABASE_URL")
	if projectURL == "" {
		return fmt.Errorf("SUPABASE_URL no está configurada")
	}

	// Intentar usar service role key primero (si está disponible), luego anon key como fallback
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SUPABASE_KEY")
		if apiKey == "" {
			return fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY o SUPABASE_KEY debe estar configurada")
		}
		log.Printf("⚠️ Usando SUPABASE_KEY (anon key) - Asegúrate de que las políticas de storage estén configuradas")
	} else {
		log.Printf("✅ Usando SUPABASE_SERVICE_ROLE_KEY (service role key)")
	}

	bucketName := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if bucketName == "" {
		bucketName = "uploads" // Bucket por defecto
	}

	// Inicializar cliente de Supabase Storage
	storageURL := fmt.Sprintf("%s/storage/v1", projectURL)
	client := storage_go.NewClient(storageURL, apiKey, nil)

	// Verificar si el bucket existe, si no, crearlo
	if err := ensureBucketExists(client, bucketName); err != nil {
		return fmt.Errorf("error configurando bucket: %v", err)
	}

	supabaseStorageService = &SupabaseStorageService{
		client:     client,
		bucketName: bucketName,
		projectURL: projectURL,
	}

	log.Printf("✅ Servicio de Supabase Storage inicializado")
	log.Printf("📦 Bucket: %s", bucketName)
	log.Printf("🌐 Storage URL: %s", storageURL)

	return nil
}

// GetSupabaseStorageService obtiene la instancia del servicio
func GetSupabaseStorageService() *SupabaseStorageService {
	return supabaseStorageService
}

// ensureBucketExists verifica si el bucket existe, si no lo crea
func ensureBucketExists(client *storage_go.Client, bucketName string) error {
	// Intentar obtener el bucket
	_, err := client.GetBucket(bucketName)
	if err != nil {
		// Si el bucket no existe, crearlo
		log.Printf("📦 Creando bucket: %s", bucketName)
		_, err = client.CreateBucket(bucketName, storage_go.BucketOptions{
			Public: false, // Bucket privado por seguridad
		})
		if err != nil {
			return fmt.Errorf("error creando bucket: %v", err)
		}
		log.Printf("✅ Bucket creado exitosamente: %s", bucketName)
	}
	return nil
}

// UploadFile sube un archivo a Supabase Storage
func (s *SupabaseStorageService) UploadFile(file multipart.File, header *multipart.FileHeader, userID, userName string) (*SupabaseUploadResponse, error) {
	log.Printf("📤 Subiendo archivo a Supabase: %s (%.2f KB)", header.Filename, float64(header.Size)/1024)

	// Crear ruta del archivo en el bucket
	fileName := fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), header.Filename)
	filePath := fmt.Sprintf("user_%s/%s", userID, fileName)

	// Leer el contenido del archivo
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Convertir []byte a io.Reader
	fileReader := bytes.NewReader(fileContent)

	// Subir archivo a Supabase Storage
	uploadResult, err := s.client.UploadFile(s.bucketName, filePath, fileReader)
	if err != nil {
		return nil, fmt.Errorf("error subiendo archivo a Supabase: %v", err)
	}

	// Generar URL pública para descargar el archivo
	publicURL := s.client.GetPublicUrl(s.bucketName, filePath)

	// Crear respuesta
	response := &SupabaseUploadResponse{
		Success:    true,
		Key:        uploadResult.Key,
		Link:       publicURL.SignedURL,
		Name:       header.Filename,
		Path:       filePath,
		Size:       header.Size,
		UploadedBy: userName,
		UploadedAt: time.Now().Format(time.RFC3339),
		BucketName: s.bucketName,
	}

	log.Printf("✅ Archivo subido exitosamente a Supabase: %s", filePath)
	log.Printf("📤 UPLOAD DETAILS: Key='%s', Path='%s', Name='%s'", response.Key, response.Path, response.Name)
	return response, nil
}

// DownloadFile descarga un archivo de Supabase Storage
func (s *SupabaseStorageService) DownloadFile(filePath string) ([]byte, error) {
	log.Printf("📥 Descargando archivo de Supabase: %s", filePath)

	// Si no contiene un slash, intentar resolver la ruta completa
	if !strings.Contains(filePath, "/") {
		log.Printf("🔍 Ruta sin carpeta detectada, buscando archivo: %s", filePath)
		resolvedPath, err := s.resolveFilePath(filePath)
		if err != nil {
			return nil, fmt.Errorf("error resolviendo ruta del archivo: %v", err)
		}
		filePath = resolvedPath
		log.Printf("✅ Ruta resuelta: %s", filePath)
	}

	// Descargar archivo
	fileData, err := s.client.DownloadFile(s.bucketName, filePath)
	if err != nil {
		return nil, fmt.Errorf("error descargando archivo: %v", err)
	}

	log.Printf("✅ Archivo descargado exitosamente: %s", filePath)
	return fileData, nil
}

// DownloadAndDeleteFile descarga un archivo, lo respalda en Telegram y luego lo elimina (para admin)
func (s *SupabaseStorageService) DownloadAndDeleteFile(filePath string) ([]byte, error) {
	log.Printf("📥🗑️ Descargando, respaldando y eliminando archivo: %s", filePath)

	// Si no contiene un slash, intentar resolver la ruta completa
	if !strings.Contains(filePath, "/") {
		log.Printf("🔍 Ruta sin carpeta detectada, buscando archivo: %s", filePath)
		resolvedPath, err := s.resolveFilePath(filePath)
		if err != nil {
			return nil, fmt.Errorf("error resolviendo ruta del archivo: %v", err)
		}
		filePath = resolvedPath
		log.Printf("✅ Ruta resuelta: %s", filePath)
	}

	// Primero descargar el archivo
	fileData, err := s.DownloadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Extraer información del archivo desde la ruta
	fileName := filepath.Base(filePath)
	userID := s.extractUserIDFromPath(filePath)

	// Intentar respaldar en Telegram antes de eliminar (solo si está habilitado)
	if IsTelegramEnabled() {
		telegramService := GetTelegramService()
		if telegramService != nil {
			log.Printf("📤 Respaldando archivo en Telegram antes de eliminar...")
			backup, err := telegramService.BackupFileToTelegram(fileData, fileName, filePath, userID)
			if err != nil {
				log.Printf("⚠️ Error respaldando archivo en Telegram: %v", err)
				// Enviar notificación de error
				telegramService.SendBackupError(fileName, userID, err.Error())
			} else {
				log.Printf("✅ Archivo respaldado exitosamente en Telegram (File ID: %s)", backup.FileID)
				// Enviar notificación de éxito
				telegramService.SendBackupNotification(fileName, userID, backup.FileSize)
			}
		} else {
			log.Printf("⚠️ Servicio de Telegram no disponible, continuando sin backup")
		}
	} else {
		log.Printf("ℹ️ Backup de Telegram deshabilitado por feature flag, continuando sin backup")
	}

	// Luego eliminar el archivo de Supabase
	err = s.DeleteFile(filePath)
	if err != nil {
		log.Printf("⚠️ Error eliminando archivo después de descarga: %v", err)
		// No retornar error aquí ya que la descarga fue exitosa
	} else {
		log.Printf("🗑️ Archivo eliminado exitosamente de Supabase después de descarga: %s", filePath)
	}

	return fileData, nil
}

// extractUserIDFromPath extrae el ID del usuario desde la ruta del archivo
func (s *SupabaseStorageService) extractUserIDFromPath(filePath string) string {
	pathParts := filepath.Dir(filePath)
	if pathParts == "." {
		return "unknown"
	}

	// La ruta debería ser algo como "user_123e4567-e89b-12d3-a456-426614174000/archivo.pdf"
	if len(pathParts) > 5 && pathParts[:5] == "user_" {
		return pathParts[5:] // Remover "user_" del inicio
	}

	return "unknown"
}

// DeleteFile elimina un archivo de Supabase Storage
func (s *SupabaseStorageService) DeleteFile(filePath string) error {
	log.Printf("🗑️ Eliminando archivo de Supabase: %s", filePath)

	// Si no contiene un slash, intentar resolver la ruta completa
	if !strings.Contains(filePath, "/") {
		log.Printf("🔍 Ruta sin carpeta detectada, buscando archivo: %s", filePath)
		resolvedPath, err := s.resolveFilePath(filePath)
		if err != nil {
			return fmt.Errorf("error resolviendo ruta del archivo: %v", err)
		}
		filePath = resolvedPath
		log.Printf("✅ Ruta resuelta: %s", filePath)
	}

	// Eliminar archivo
	_, err := s.client.RemoveFile(s.bucketName, []string{filePath})
	if err != nil {
		return fmt.Errorf("error eliminando archivo: %v", err)
	}

	log.Printf("✅ Archivo eliminado exitosamente: %s", filePath)
	return nil
}

// ListUserFiles lista archivos de un usuario específico
func (s *SupabaseStorageService) ListUserFiles(userID string) ([]SupabaseFileInfo, error) {
	log.Printf("📋 Listando archivos del usuario: %s", userID)

	// Listar archivos en la carpeta del usuario
	userFolder := fmt.Sprintf("user_%s", userID)
	files, err := s.client.ListFiles(s.bucketName, userFolder, storage_go.FileSearchOptions{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("error listando archivos: %v", err)
	}

	var fileInfos []SupabaseFileInfo
	for _, file := range files {
		// Generar URL de descarga para cada archivo
		publicURL := s.client.GetPublicUrl(s.bucketName, file.Name)

		// Obtener tamaño del archivo desde metadata
		size := int64(0)
		if file.Metadata != nil {
			if metadata, ok := file.Metadata.(map[string]interface{}); ok {
				if sizeValue, exists := metadata["size"]; exists {
					if sizeFloat, ok := sizeValue.(float64); ok {
						size = int64(sizeFloat)
					}
				}
			}
		}

		// Obtener tipo MIME desde metadata
		mimeType := ""
		if file.Metadata != nil {
			if metadata, ok := file.Metadata.(map[string]interface{}); ok {
				if typeValue, exists := metadata["mimetype"]; exists {
					if typeStr, ok := typeValue.(string); ok {
						mimeType = typeStr
					}
				}
			}
		}

		fileInfo := SupabaseFileInfo{
			Name:        filepath.Base(file.Name),
			Size:        size,
			Path:        file.Name,
			MimeType:    mimeType,
			UploadedAt:  file.CreatedAt,
			DownloadURL: publicURL.SignedURL,
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	log.Printf("📋 Encontrados %d archivos para el usuario %s", len(fileInfos), userID)
	return fileInfos, nil
}

// ListAllFiles lista todos los archivos en el bucket (solo para admins)
func (s *SupabaseStorageService) ListAllFiles() ([]SupabaseFileInfo, error) {
	log.Printf("📋 Listando todos los archivos del bucket: %s", s.bucketName)

	// Primero listar carpetas/directorios
	folders, err := s.client.ListFiles(s.bucketName, "", storage_go.FileSearchOptions{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("error listando carpetas: %v", err)
	}

	log.Printf("🔍 DEBUG: Encontrados %d elementos en el bucket", len(folders))

	var allFileInfos []SupabaseFileInfo

	// Procesar cada elemento encontrado
	for _, folder := range folders {
		log.Printf("📁 DEBUG ListAllFiles: Name='%s', ID='%s'", folder.Name, folder.Id)

		// Si es una carpeta (user_xxx), listar archivos dentro de ella
		if strings.HasPrefix(folder.Name, "user_") && !strings.Contains(folder.Name, ".") {
			log.Printf("📂 Explorando carpeta de usuario: %s", folder.Name)

			// Listar archivos dentro de esta carpeta de usuario
			userFiles, err := s.client.ListFiles(s.bucketName, folder.Name, storage_go.FileSearchOptions{
				Limit:  100,
				Offset: 0,
			})
			if err != nil {
				log.Printf("⚠️ Error listando archivos en carpeta %s: %v", folder.Name, err)
				continue
			}

			log.Printf("📄 Encontrados %d archivos en carpeta %s", len(userFiles), folder.Name)

			// Procesar archivos de esta carpeta
			for _, file := range userFiles {
				log.Printf("📄 DEBUG UserFile: Name='%s', ID='%s'", file.Name, file.Id)

				// Solo procesar archivos reales (que tengan extensión)
				if !strings.Contains(file.Name, ".") {
					log.Printf("⏭️ Saltando elemento sin extensión: %s", file.Name)
					continue
				}

				// Crear FileInfo para este archivo
				fileInfo := s.createFileInfo(file)
				allFileInfos = append(allFileInfos, fileInfo)
			}
		} else if strings.Contains(folder.Name, ".") {
			// Es un archivo directo en la raíz
			log.Printf("📄 Archivo en raíz: %s", folder.Name)
			fileInfo := s.createFileInfo(folder)
			allFileInfos = append(allFileInfos, fileInfo)
		} else {
			log.Printf("⏭️ Saltando elemento: %s", folder.Name)
		}
	}

	log.Printf("📋 Encontrados %d archivos totales en el bucket", len(allFileInfos))
	return allFileInfos, nil
}

// createFileInfo crea un SupabaseFileInfo desde un FileObject
func (s *SupabaseStorageService) createFileInfo(file storage_go.FileObject) SupabaseFileInfo {
	// Generar URL de descarga para cada archivo
	publicURL := s.client.GetPublicUrl(s.bucketName, file.Name)

	// Obtener tamaño del archivo desde metadata
	size := int64(0)
	if file.Metadata != nil {
		if metadata, ok := file.Metadata.(map[string]interface{}); ok {
			if sizeValue, exists := metadata["size"]; exists {
				if sizeFloat, ok := sizeValue.(float64); ok {
					size = int64(sizeFloat)
				}
			}
		}
	}

	// Obtener tipo MIME desde metadata
	mimeType := ""
	if file.Metadata != nil {
		if metadata, ok := file.Metadata.(map[string]interface{}); ok {
			if typeValue, exists := metadata["mimetype"]; exists {
				if typeStr, ok := typeValue.(string); ok {
					mimeType = typeStr
				}
			}
		}
	}

	return SupabaseFileInfo{
		Name:        filepath.Base(file.Name),
		Size:        size,
		Path:        file.Name,
		MimeType:    mimeType,
		UploadedAt:  file.CreatedAt,
		DownloadURL: publicURL.SignedURL,
	}
}

// GetFileInfo obtiene información de un archivo específico
func (s *SupabaseStorageService) GetFileInfo(filePath string) (*SupabaseFileInfo, error) {
	// Intentar obtener la información del archivo
	files, err := s.client.ListFiles(s.bucketName, filepath.Dir(filePath), storage_go.FileSearchOptions{
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del archivo: %v", err)
	}

	// Buscar el archivo específico
	var targetFile *storage_go.FileObject
	for _, file := range files {
		if file.Name == filePath {
			targetFile = &file
			break
		}
	}

	if targetFile == nil {
		return nil, fmt.Errorf("archivo no encontrado: %s", filePath)
	}

	// Generar URL de descarga
	publicURL := s.client.GetPublicUrl(s.bucketName, targetFile.Name)

	// Obtener tamaño del archivo desde metadata
	size := int64(0)
	if targetFile.Metadata != nil {
		if metadata, ok := targetFile.Metadata.(map[string]interface{}); ok {
			if sizeValue, exists := metadata["size"]; exists {
				if sizeFloat, ok := sizeValue.(float64); ok {
					size = int64(sizeFloat)
				}
			}
		}
	}

	// Obtener tipo MIME desde metadata
	mimeType := ""
	if targetFile.Metadata != nil {
		if metadata, ok := targetFile.Metadata.(map[string]interface{}); ok {
			if typeValue, exists := metadata["mimetype"]; exists {
				if typeStr, ok := typeValue.(string); ok {
					mimeType = typeStr
				}
			}
		}
	}

	fileInfo := &SupabaseFileInfo{
		Name:        filepath.Base(targetFile.Name),
		Size:        size,
		Path:        targetFile.Name,
		MimeType:    mimeType,
		UploadedAt:  targetFile.CreatedAt,
		DownloadURL: publicURL.SignedURL,
	}

	return fileInfo, nil
}

// resolveFilePath busca la ruta completa de un archivo basado en su nombre
func (s *SupabaseStorageService) resolveFilePath(fileName string) (string, error) {
	log.Printf("🔍 Resolviendo ruta para archivo: %s", fileName)

	// Listar todas las carpetas/directorios
	folders, err := s.client.ListFiles(s.bucketName, "", storage_go.FileSearchOptions{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return "", fmt.Errorf("error listando carpetas: %v", err)
	}

	// Buscar en cada carpeta de usuario
	for _, folder := range folders {
		// Solo buscar en carpetas de usuario
		if strings.HasPrefix(folder.Name, "user_") && !strings.Contains(folder.Name, ".") {
			log.Printf("🔍 Buscando en carpeta: %s", folder.Name)

			// Listar archivos dentro de esta carpeta
			userFiles, err := s.client.ListFiles(s.bucketName, folder.Name, storage_go.FileSearchOptions{
				Limit:  100,
				Offset: 0,
			})
			if err != nil {
				log.Printf("⚠️ Error listando archivos en carpeta %s: %v", folder.Name, err)
				continue
			}

			// Buscar el archivo específico
			for _, file := range userFiles {
				if filepath.Base(file.Name) == fileName || file.Name == fileName {
					log.Printf("✅ Archivo encontrado: %s", file.Name)
					// Retornar la ruta completa con la carpeta del usuario
					fullPath := file.Name
					if !strings.HasPrefix(fullPath, folder.Name+"/") {
						fullPath = folder.Name + "/" + filepath.Base(file.Name)
					}
					log.Printf("✅ Ruta completa resuelta: %s", fullPath)
					return fullPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("archivo no encontrado: %s", fileName)
}
