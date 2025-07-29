# 🤖 Integración con Telegram Bot para Backup de Archivos

## 📋 Descripción

Tu aplicación ahora está integrada con **Telegram Bot** para crear backups automáticos de todos los archivos antes de eliminarlos de Supabase Storage. Esta funcionalidad garantiza que nunca pierdas archivos importantes, utilizando Telegram como almacenamiento secundario gratuito.

## 🔧 Configuración del Bot

### Bot Información:
- **Nombre**: @bescao_bot  
- **Token**: `7918141497:AAF225FnXmvATYI1gZHsSx3lUJkrXCxNlh8`
- **Chat ID**: `1540590265`
- **URL**: https://t.me/bescao_bot

### Variables de Entorno Requeridas:

```env
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=7918141497:AAF225FnXmvATYI1gZHsSx3lUJkrXCxNlh8
TELEGRAM_CHAT_ID=1540590265
```

## 🚀 Funcionalidades Implementadas

### 1. **Backup Automático Antes de Eliminación**
- Cuando un admin descarga y elimina un archivo, se respalda automáticamente en Telegram
- El archivo se envía con metadatos completos (usuario, fecha, tamaño, ruta original)
- Solo después del backup exitoso se elimina de Supabase

### 2. **Notificaciones Inteligentes**
- ✅ **Backup Exitoso**: Notificación con detalles del archivo respaldado
- ❌ **Error en Backup**: Alerta cuando falla el respaldo
- 📊 **Información Detallada**: Usuario, tamaño, fecha y ubicación original

### 3. **Gestión de Errores**
- Si Telegram no está disponible, el sistema continúa funcionando
- Los archivos se eliminan normalmente, pero sin backup
- Logs detallados para debugging

## 📁 Estructura de Mensajes en Telegram

### Backup de Archivo:
```
📁 Backup de archivo
📄 Archivo: documento_importante.pdf
👤 Usuario: 123e4567-e89b-12d3-a456-426614174000
📂 Ruta: user_123e4567-e89b-12d3-a456-426614174000/documento_importante.pdf
📅 Fecha: 2024-01-15 14:30:00
💾 Tamaño: 2.5 KB
```

### Notificación de Éxito:
```
✅ Backup completado

📄 Archivo: documento_importante.pdf
👤 Usuario: 123e4567-e89b-12d3-a456-426614174000
💾 Tamaño: 2.5 KB
🕐 Fecha: 2024-01-15 14:30:00

El archivo ha sido respaldado exitosamente antes de ser eliminado de Supabase.
```

### Notificación de Error:
```
❌ Error en backup

📄 Archivo: documento_importante.pdf
👤 Usuario: 123e4567-e89b-12d3-a456-426614174000
🚨 Error: timeout sending file to Telegram
🕐 Fecha: 2024-01-15 14:30:00

⚠️ El archivo NO fue respaldado. Revisar logs.
```

## 🔄 Flujo de Proceso

### Descarga y Eliminación de Archivo:

1. **Admin descarga archivo** desde `/admin/file-management`
2. **Sistema descarga** archivo de Supabase Storage
3. **Intenta backup** en Telegram con metadatos completos
4. **Si backup exitoso**:
   - ✅ Envía archivo a Telegram
   - ✅ Envía notificación de éxito
   - ✅ Elimina archivo de Supabase
   - ✅ Devuelve archivo al admin
5. **Si backup falla**:
   - ❌ Envía notificación de error
   - ⚠️ **AÚN elimina** archivo de Supabase (comportamiento configurable)
   - ✅ Devuelve archivo al admin

## 🛠️ Servicios Implementados

### TelegramService (`backend/service/telegram_service.go`)

**Métodos principales:**
- `InitializeTelegramService()` - Inicializa y verifica conexión
- `BackupFileToTelegram()` - Respalda archivo con metadatos
- `SendBackupNotification()` - Notifica backup exitoso
- `SendBackupError()` - Notifica errores
- `GetFileFromTelegram()` - Descarga archivo desde Telegram

### Integración con SupabaseStorageService

**Método modificado:**
- `DownloadAndDeleteFile()` - Ahora incluye backup automático en Telegram

## 📊 Logs del Sistema

### Inicio Exitoso:
```
✅ Servicio de Telegram inicializado
🤖 Bot: @bescao_bot
💬 Chat ID: 1540590265
🔗 Conexión con Telegram establecida exitosamente
```

### Backup de Archivo:
```
📥🗑️ Descargando, respaldando y eliminando archivo: user_123/documento.pdf
📤 Respaldando archivo en Telegram antes de eliminar...
📤 Respaldando archivo en Telegram: documento.pdf (2.50 KB)
✅ Archivo respaldado exitosamente en Telegram (File ID: BAADBAADrwADBREAAWYWXY...)
🗑️ Archivo eliminado exitosamente de Supabase después de descarga: user_123/documento.pdf
```

### Error de Telegram:
```
⚠️ Advertencia: Servicio de Telegram no disponible: TELEGRAM_BOT_TOKEN no está configurada
ℹ️ Los archivos se eliminarán sin backup en Telegram
```

## 🔒 Consideraciones de Seguridad

### ✅ **Ventajas**:
- Backup automático antes de eliminación permanente
- Metadatos completos para identificación
- Notificaciones en tiempo real
- Almacenamiento gratuito e ilimitado en Telegram
- Acceso desde cualquier dispositivo con Telegram

### ⚠️ **Consideraciones**:
- Los archivos en Telegram son accesibles para quien tenga acceso al chat
- Telegram tiene límite de 2GB por archivo
- La conexión a Telegram debe estar disponible

## 🧪 Pruebas y Validación

### Para probar la integración:

1. **Configurar variables de entorno** en `.env`
2. **Ejecutar backend**: `cd backend && go run .`
3. **Verificar logs** de inicialización de Telegram
4. **Subir archivo** como usuario autenticado
5. **Ir a Admin Panel** → File Management
6. **Descargar archivo** (esto activará el backup)
7. **Verificar en Telegram** que llegó el archivo y notificaciones

### Comandos de prueba:

```bash
# Verificar configuración
cd backend
go run . 2>&1 | grep -E "(Telegram|🤖|💬)"

# Probar upload y descarga
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -F "file=@test.pdf" \
  http://localhost:8080/api/upload/file-authenticated
```

## 🎯 Beneficios de la Integración

### **Para Administradores:**
- ✅ Tranquilidad: archivos siempre respaldados
- ✅ Visibilidad: notificaciones inmediatas
- ✅ Acceso: archivos disponibles en Telegram
- ✅ Historial: búsqueda fácil en chat de Telegram

### **Para el Sistema:**
- ✅ Resiliencia: backup antes de eliminación
- ✅ Trazabilidad: logs detallados
- ✅ Flexibilidad: funciona con/sin Telegram
- ✅ Escalabilidad: Telegram maneja almacenamiento

## 🔧 Configuración de Producción

### Variables de entorno requeridas:
```env
# Telegram Bot (obligatorio para backup)
TELEGRAM_BOT_TOKEN=7918141497:AAF225FnXmvATYI1gZHsSx3lUJkrXCxNlh8
TELEGRAM_CHAT_ID=1540590265

# Supabase (obligatorio)
SUPABASE_URL=https://wbnoxgtrahnlskrlhkmy.supabase.co
SUPABASE_KEY=tu-supabase-key-aqui
SUPABASE_SERVICE_ROLE_KEY=tu-service-role-key-aqui

# Storage
SUPABASE_STORAGE_BUCKET=uploads
```

## 📈 Estadísticas y Monitoreo

El sistema genera logs que puedes monitorear para:
- Número de archivos respaldados exitosamente
- Errores de backup
- Tamaño total de archivos respaldados
- Usuarios más activos
- Tiempos de respuesta de Telegram

---

**¡Tu sistema ahora tiene backup automático integrado con Telegram!** 🎉

Cada archivo eliminado será preservado automáticamente, dándote la tranquilidad de que nunca perderás información importante. 