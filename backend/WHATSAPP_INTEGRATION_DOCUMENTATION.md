# WhatsApp Integration Documentation

Esta documentación contiene toda la implementación de WhatsApp que fue removida del sistema principal para ser implementada como microservicio independiente.

## Descripción General

La integración de WhatsApp permitía:
- Conexión automática con WhatsApp Business API usando whatsmeow
- Gestión de sesiones de chat con usuarios
- Menús interactivos configurables
- Envío y recepción de mensajes
- Vinculación de usuarios del sistema con números de WhatsApp
- Interfaz de administración web completa

## Dependencias Go Requeridas

```go
// En go.mod
go.mau.fi/whatsmeow v0.0.0-20250617170509-947866bb9f75
go.mau.fi/libsignal v0.2.0
go.mau.fi/util v0.8.8
```

## Estructura de Archivos

### Backend
- `service/whatsapp_service.go` - Servicio principal de WhatsApp
- `controller/whatsapp_controller.go` - Controlador HTTP para API REST
- `model/whatsapp_messages.go` - Modelos y estructuras de datos
- `whatsapp_data/` - Directorio de datos de sesión
- `whatsapp_data/config.json` - Configuración de mensajes

### Frontend
- `pages/Admin/WhatsAppManagement.tsx` - Interfaz de administración

## Implementación del Servicio Principal

```go
// Archivo: service/whatsapp_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/util/keys"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// FileDeviceStore implementa un almacén de dispositivos basado en archivos
type FileDeviceStore struct {
	dataDir string
	*store.NoopStore
}

func NewFileDeviceStore(dataDir string) *FileDeviceStore {
	return &FileDeviceStore{
		dataDir:   dataDir,
		NoopStore: &store.NoopStore{},
	}
}

// Funciones principales del FileDeviceStore
func (f *FileDeviceStore) PutDevice(device *store.Device) error {
	devicePath := filepath.Join(f.dataDir, "device.json")
	
	deviceData := map[string]interface{}{
		"noise_key":       device.NoiseKey.Pub[:],
		"noise_key_priv":  device.NoiseKey.Priv[:],
		"identity_key":    device.IdentityKey.Pub[:],
		"identity_key_priv": device.IdentityKey.Priv[:],
		"signed_pre_key":  device.SignedPreKey.Pub[:],
		"signed_pre_key_priv": device.SignedPreKey.Priv[:],
		"signed_pre_key_id": device.SignedPreKey.KeyID,
		"signed_pre_key_sig": device.SignedPreKey.Signature,
		"registration_id": device.RegistrationID,
		"adv_secret_key":  device.AdvSecretKey,
	}

	data, err := json.MarshalIndent(deviceData, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando dispositivo: %v", err)
	}

	if err := os.WriteFile(devicePath, data, 0600); err != nil {
		return fmt.Errorf("error guardando dispositivo: %v", err)
	}

	return nil
}

func (f *FileDeviceStore) LoadDevice() (*store.Device, error) {
	devicePath := filepath.Join(f.dataDir, "device.json")
	
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		log.Printf("📱 No existe dispositivo guardado, creando nuevo...")
		return f.createNewDevice()
	}

	data, err := os.ReadFile(devicePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo de dispositivo: %v", err)
	}

	var deviceData map[string]interface{}
	if err := json.Unmarshal(data, &deviceData); err != nil {
		log.Printf("⚠️ Error deserializando dispositivo, creando nuevo: %v", err)
		return f.createNewDevice()
	}

	// Restaurar claves desde los datos guardados
	device := &store.Device{
		Identities:     f.NoopStore,
		Sessions:       f.NoopStore,
		PreKeys:        f.NoopStore,
		SenderKeys:     f.NoopStore,
		AppStateKeys:   f.NoopStore,
		AppState:       f.NoopStore,
		Contacts:       f.NoopStore,
		ChatSettings:   f.NoopStore,
		MsgSecrets:     f.NoopStore,
		PrivacyTokens:  f.NoopStore,
		EventBuffer:    f.NoopStore,
		Container:      f,
	}

	return device, nil
}

// WhatsAppService - Servicio principal
type WhatsAppService struct {
	client        *whatsmeow.Client
	deviceStore   *FileDeviceStore
	sessions      map[string]*model.WhatsAppSession
	sessionsMutex sync.RWMutex
	config        *model.WhatsAppConfiguration
	configMutex   sync.RWMutex
	userRepo      *storage.UserRepository
	isConnected   bool
	qrCode        string
	qrMutex       sync.RWMutex
}

// Inicialización del servicio
func InitializeWhatsAppService(userRepo *storage.UserRepository) (*WhatsAppService, error) {
	whatsappService := &WhatsAppService{
		sessions: make(map[string]*model.WhatsAppSession),
		userRepo: userRepo,
	}

	dataDir := "./whatsapp_data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de datos: %v", err)
	}

	fileStore := NewFileDeviceStore(dataDir)
	whatsappService.deviceStore = fileStore

	device, err := fileStore.LoadDevice()
	if err != nil {
		return nil, fmt.Errorf("error cargando dispositivo: %v", err)
	}

	client := whatsmeow.NewClient(device, nil)
	whatsappService.client = client
	client.AddEventHandler(whatsappService.handleEvents)
	whatsappService.loadConfiguration()

	return whatsappService, nil
}

// Métodos principales del servicio
func (s *WhatsAppService) Connect() error {
	if s.client.IsConnected() {
		s.isConnected = true
		return nil
	}

	if s.client.Store.ID == nil {
		qrChan, err := s.client.GetQRChannel(context.Background())
		if err != nil {
			return fmt.Errorf("error obteniendo canal QR: %v", err)
		}

		go func() {
			err := s.client.Connect()
			if err != nil {
				log.Printf("❌ Error conectando: %v", err)
			}
		}()

		for evt := range qrChan {
			if evt.Event == "code" {
				s.saveQRCode(evt.Code)
			} else if evt.Event == "success" {
				s.isConnected = true
				break
			}
		}
	} else {
		err := s.client.Connect()
		if err != nil {
			return fmt.Errorf("error conectando: %v", err)
		}
		s.isConnected = true
	}

	return nil
}

func (s *WhatsAppService) SendMessage(phone, message string) error {
	if !s.IsConnected() {
		return fmt.Errorf("WhatsApp no está conectado")
	}

	cleanPhone := strings.ReplaceAll(phone, "+", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")

	jid := types.NewJID(cleanPhone, types.DefaultUserServer)

	_, err := s.client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: &message,
	})

	return err
}
```

## Modelos de Datos

```go
// Archivo: model/whatsapp_messages.go
package model

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
)

type WhatsAppMessage struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FromNumber  string    `json:"from_number" db:"from_number"`
	ToNumber    string    `json:"to_number" db:"to_number"`
	MessageText string    `json:"message_text" db:"message_text"`
	MessageType string    `json:"message_type" db:"message_type"`
	Direction   string    `json:"direction" db:"direction"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type WhatsAppSession struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	PhoneNumber  string                 `json:"phone_number" db:"phone_number"`
	UserID       *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	CurrentState string                 `json:"current_state" db:"current_state"`
	SessionData  map[string]interface{} `json:"session_data" db:"session_data"`
	LastActivity time.Time              `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

type WhatsAppConfiguration struct {
	WelcomeMessage string                      `json:"welcome_message"`
	Responses      map[string]WhatsAppResponse `json:"responses"`
	DefaultMessage string                      `json:"default_message"`
	ErrorMessage   string                      `json:"error_message"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}
```

## Controlador HTTP

```go
// Archivo: controller/whatsapp_controller.go
package controller

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/service"
)

type WhatsAppController struct {
	whatsappService *service.WhatsAppService
}

func NewWhatsAppController() *WhatsAppController {
	return &WhatsAppController{
		whatsappService: service.GetWhatsAppService(),
	}
}

// Endpoints principales:
// GET /admin/whatsapp/status - Estado de conexión
// POST /admin/whatsapp/connect - Conectar servicio
// POST /admin/whatsapp/disconnect - Desconectar servicio
// GET /admin/whatsapp/qr - Obtener código QR
// GET /admin/whatsapp/config - Obtener configuración
// PUT /admin/whatsapp/config - Actualizar configuración
// POST /admin/whatsapp/send-message - Enviar mensaje manual
// GET /admin/whatsapp/sessions - Ver sesiones activas
// POST /admin/whatsapp/link-user - Vincular usuario con teléfono
```

## Interfaz de Administración Frontend

```tsx
// Archivo: pages/Admin/WhatsAppManagement.tsx
import React, { useState, useEffect } from 'react';
import {
  Container, Title, Paper, Group, Button, Badge, Stack,
  Text, Alert, LoadingOverlay, Tabs, Textarea, TextInput,
  Modal, Code,
} from '@mantine/core';
import {
  IconBrandWhatsapp, IconWifi, IconWifiOff, IconSettings,
  IconMessage, IconSend, IconRefresh, IconQrcode,
  IconCheck, IconAlertCircle,
} from '@tabler/icons-react';

const WhatsAppManagement: React.FC = () => {
  // Estados para manejo de conexión, QR, envío de mensajes
  // Funciones para conectar, desconectar, enviar mensajes
  // Interfaz completa de administración con pestañas
  
  return (
    <Container size="xl" py="md">
      {/* Interfaz completa con estado de conexión, envío de mensajes, configuración */}
    </Container>
  );
};
```

## Configuración de Mensajes por Defecto

```json
{
  "welcome_message": "¡Hola! 👋 Bienvenido al sistema de gestión de alquileres. ¿En qué puedo ayudarte hoy?",
  "responses": {
    "menu": {
      "state": "main_menu",
      "message": "📋 *Menú Principal*\n\nSelecciona una opción:",
      "menu": {
        "title": "Opciones Disponibles",
        "options": [
          {"key": "1", "text": "📊 Ver mis propiedades", "next_state": "properties"},
          {"key": "2", "text": "💰 Estado de pagos", "next_state": "payments"},
          {"key": "3", "text": "🔧 Reportar mantenimiento", "next_state": "maintenance"},
          {"key": "4", "text": "📄 Mis contratos", "next_state": "contracts"},
          {"key": "5", "text": "👤 Información de contacto", "next_state": "contact_info"},
          {"key": "0", "text": "❌ Salir", "next_state": "goodbye"}
        ]
      }
    }
  },
  "default_message": "Lo siento, no entendí tu mensaje. Escribe *menu* para ver las opciones disponibles.",
  "error_message": "Ha ocurrido un error. Por favor, intenta nuevamente o contacta a soporte."
}
```

## Funcionalidades Implementadas

### 1. Conexión y Autenticación
- Generación automática de códigos QR para vinculación
- Manejo de sesiones persistentes con archivos JSON
- Reconexión automática en caso de desconexión

### 2. Gestión de Mensajes
- Envío y recepción de mensajes de texto
- Sistema de menús interactivos configurables
- Sesiones de chat con estado persistente
- Respuestas automáticas basadas en configuración

### 3. Integración con Sistema
- Vinculación de números de WhatsApp con usuarios del sistema
- Autenticación requerida para ciertas funciones
- Acceso a información de propiedades, pagos, contratos

### 4. Administración
- Interfaz web completa para gestión
- Envío manual de mensajes
- Monitoreo de sesiones activas
- Configuración de mensajes automáticos

## Arquitectura del Microservicio

Para implementar como microservicio independiente:

1. **API Gateway**: Exponer endpoints RESTful
2. **Base de Datos**: PostgreSQL para sesiones y mensajes
3. **Redis**: Cache para sesiones activas
4. **WebHooks**: Notificaciones al sistema principal
5. **Docker**: Contenerización completa

## Consideraciones de Seguridad

- Validación de números de teléfono
- Rate limiting para envío de mensajes
- Encriptación de datos de sesión
- Logs de auditoría completos
- Validación de tokens de autenticación

## Estado de la Implementación

✅ **Completado:**
- Conexión básica con WhatsApp
- Envío y recepción de mensajes
- Sistema de menús interactivos
- Interfaz de administración web
- Configuración flexible de mensajes

🔧 **Por Implementar en Microservicio:**
- Base de datos persistente para mensajes
- Sistema de notificaciones push
- Métricas y monitoreo
- Escalamiento horizontal
- Integración con múltiples números

## Notas de Migración

Este código fue removido del sistema principal en la fecha actual para ser implementado como microservicio independiente. Toda la funcionalidad está documentada aquí para referencia futura. 