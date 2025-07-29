package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nescool101/rentManager/storage"
)

// TelegramService maneja la integración con Telegram para backup de archivos
type TelegramService struct {
	botToken string
	chatID   string
	baseURL  string
}

// TelegramResponse respuesta de la API de Telegram
type TelegramResponse struct {
	Ok     bool   `json:"ok"`
	Result Result `json:"result"`
}

// Result resultado del envío de documento
type Result struct {
	MessageID int      `json:"message_id"`
	Date      int64    `json:"date"`
	Document  Document `json:"document"`
}

// Document información del documento enviado
type Document struct {
	FileName string `json:"file_name"`
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
}

// TelegramFileBackup información del archivo respaldado en Telegram
type TelegramFileBackup struct {
	Success     bool   `json:"success"`
	FileID      string `json:"file_id"`      // ID del archivo en Telegram
	MessageID   int    `json:"message_id"`   // ID del mensaje en Telegram
	FileName    string `json:"file_name"`    // Nombre original del archivo
	FileSize    int64  `json:"file_size"`    // Tamaño del archivo
	BackupDate  string `json:"backup_date"`  // Fecha de respaldo
	OriginalURL string `json:"original_url"` // URL original en Supabase
}

var telegramService *TelegramService

// InitializeTelegramService inicializa el servicio de Telegram
func InitializeTelegramService() error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN no está configurada")
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if chatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID no está configurada")
	}

	telegramService = &TelegramService{
		botToken: botToken,
		chatID:   chatID,
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
	}

	log.Printf("✅ Servicio de Telegram inicializado")
	log.Printf("🤖 Bot: @bescao_bot")
	log.Printf("💬 Chat ID: %s", chatID)

	// Probar conexión
	if err := telegramService.testConnection(); err != nil {
		return fmt.Errorf("error probando conexión con Telegram: %v", err)
	}

	return nil
}

// GetTelegramService obtiene la instancia del servicio
func GetTelegramService() *TelegramService {
	return telegramService
}

// testConnection prueba la conexión con Telegram
func (t *TelegramService) testConnection() error {
	url := fmt.Sprintf("%s/getMe", t.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error de conexión: %d", resp.StatusCode)
	}

	log.Printf("🔗 Conexión con Telegram establecida exitosamente")
	return nil
}

// getUserName obtiene el nombre completo del usuario desde su ID
func (t *TelegramService) getUserName(userID string) string {
	// Convertir userID string a UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Error parseando userID %s: %v", userID, err)
		return "Usuario desconocido"
	}

	// Crear cliente Supabase directamente
	supabaseClient, err := storage.InitializeSupabaseClient()
	if err != nil {
		log.Printf("Error inicializando cliente Supabase: %v", err)
		return "Usuario desconocido"
	}

	// Crear repositorios
	repoFactory := storage.NewRepositoryFactory(supabaseClient)
	userRepo := repoFactory.GetUserRepository()
	personRepo := repoFactory.GetPersonRepository()

	// Buscar usuario por ID
	ctx := context.Background()
	user, err := userRepo.GetByID(ctx, userUUID)
	if err != nil || user == nil {
		log.Printf("Error obteniendo usuario %s: %v", userID, err)
		return "Usuario desconocido"
	}

	// Buscar persona asociada
	person, err := personRepo.GetByID(ctx, user.PersonID)
	if err != nil || person == nil {
		log.Printf("Error obteniendo persona para usuario %s: %v", userID, err)
		return "Usuario desconocido"
	}

	return person.FullName
}

// BackupFileToTelegram respalda un archivo en Telegram antes de eliminarlo
func (t *TelegramService) BackupFileToTelegram(fileData []byte, fileName, originalPath, userID string) (*TelegramFileBackup, error) {
	log.Printf("📤 Respaldando archivo en Telegram: %s (%.2f KB)", fileName, float64(len(fileData))/1024)

	// Obtener nombre del usuario
	userName := t.getUserName(userID)

	// Crear el cuerpo de la petición multipart
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Agregar chat_id
	writer.WriteField("chat_id", t.chatID)

	// Agregar caption con información del archivo incluyendo el nombre del usuario
	caption := fmt.Sprintf("📁 Backup de archivo\n"+
		"📄 Archivo: %s\n"+
		"👤 Usuario: %s (%s)\n"+
		"📂 Ruta: %s\n"+
		"📅 Fecha: %s\n"+
		"💾 Tamaño: %.2f KB",
		fileName,
		userName,
		userID,
		originalPath,
		time.Now().Format("2006-01-02 15:04:05"),
		float64(len(fileData))/1024)

	writer.WriteField("caption", caption)

	// Agregar el archivo
	part, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return nil, fmt.Errorf("error creando form file: %v", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("error escribiendo archivo: %v", err)
	}

	writer.Close()

	// Enviar archivo a Telegram
	url := fmt.Sprintf("%s/sendDocument", t.baseURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error enviando archivo: %v", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error de Telegram: %d - %s", resp.StatusCode, string(responseBody))
	}

	// Parsear respuesta
	var telegramResp TelegramResponse
	if err := json.Unmarshal(responseBody, &telegramResp); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	if !telegramResp.Ok {
		return nil, fmt.Errorf("error de Telegram API: respuesta no OK")
	}

	// Crear respuesta de backup
	backup := &TelegramFileBackup{
		Success:     true,
		FileID:      telegramResp.Result.Document.FileID,
		MessageID:   telegramResp.Result.MessageID,
		FileName:    fileName,
		FileSize:    int64(len(fileData)),
		BackupDate:  time.Now().Format(time.RFC3339),
		OriginalURL: originalPath,
	}

	log.Printf("✅ Archivo respaldado exitosamente en Telegram")
	log.Printf("🆔 File ID: %s", backup.FileID)
	log.Printf("💌 Message ID: %d", backup.MessageID)

	return backup, nil
}

// SendBackupNotification envía una notificación de backup completado
func (t *TelegramService) SendBackupNotification(fileName, userID string, fileSize int64) error {
	// Obtener nombre del usuario
	userName := t.getUserName(userID)

	message := fmt.Sprintf("✅ Backup completado\n\n"+
		"📄 Archivo: %s\n"+
		"👤 Usuario: %s (%s)\n"+
		"💾 Tamaño: %.2f KB\n"+
		"🕐 Fecha: %s\n\n"+
		"El archivo ha sido respaldado exitosamente antes de ser eliminado de Supabase.",
		fileName,
		userName,
		userID,
		float64(fileSize)/1024,
		time.Now().Format("2006-01-02 15:04:05"))

	return t.sendMessage(message)
}

// SendBackupError envía una notificación de error en backup
func (t *TelegramService) SendBackupError(fileName, userID, errorMsg string) error {
	// Obtener nombre del usuario
	userName := t.getUserName(userID)

	message := fmt.Sprintf("❌ Error en backup\n\n"+
		"📄 Archivo: %s\n"+
		"👤 Usuario: %s (%s)\n"+
		"🚨 Error: %s\n"+
		"🕐 Fecha: %s\n\n"+
		"⚠️ El archivo NO fue respaldado. Revisar logs.",
		fileName,
		userName,
		userID,
		errorMsg,
		time.Now().Format("2006-01-02 15:04:05"))

	return t.sendMessage(message)
}

// sendMessage envía un mensaje de texto a Telegram
func (t *TelegramService) sendMessage(text string) error {
	url := fmt.Sprintf("%s/sendMessage", t.baseURL)

	payload := map[string]interface{}{
		"chat_id": t.chatID,
		"text":    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error enviando mensaje: %d", resp.StatusCode)
	}

	return nil
}

// GetFileFromTelegram descarga un archivo desde Telegram usando su file_id
func (t *TelegramService) GetFileFromTelegram(fileID string) ([]byte, error) {
	// Obtener información del archivo
	getFileURL := fmt.Sprintf("%s/getFile?file_id=%s", t.baseURL, fileID)
	resp, err := http.Get(getFileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fileInfo struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, err
	}

	if !fileInfo.Ok {
		return nil, fmt.Errorf("error obteniendo información del archivo")
	}

	// Descargar archivo
	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", t.botToken, fileInfo.Result.FilePath)
	downloadResp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer downloadResp.Body.Close()

	return io.ReadAll(downloadResp.Body)
}
