# 🔐 Guía del Sistema de Subida Segura de Archivos

## Descripción General

El sistema de subida de archivos ha sido mejorado para requerir autenticación de usuario y vincular todos los archivos a carpetas específicas por usuario. Esto garantiza la seguridad y organización de los documentos.

## 🔄 Flujo Completo del Sistema

### 1. **Admin/Manager genera enlace de subida**
- Accede a "Gestión de Archivos" en el panel admin
- Selecciona un usuario específico del sistema
- El email y nombre se autocompletan basados en el usuario seleccionado
- Define días de validez del enlace (1-365 días)
- Sistema genera token único vinculado al usuario

### 2. **Email automático al usuario**
- Se envía email con enlace seguro al usuario seleccionado
- El enlace incluye el token único: `http://dominio.com/file-upload?token=abc123`
- Email incluye instrucciones y tipos de archivo permitidos

### 3. **Usuario accede al enlace**
- Usuario debe estar **autenticado** en el sistema
- Sistema valida que el usuario logueado coincida con el destinatario del token
- Si no está logueado, se redirige al login con returnURL

### 4. **Subida de archivos**
- Usuario sube archivos que se organizan automáticamente en:
  - Carpeta: `User_{UserID}_{UserName}`
  - Ejemplo: `User_123e4567-e89b-12d3-a456-426614174000_juan.perez@email.com`
- Los archivos quedan vinculados permanentemente al usuario

## ✅ Funcionalidades de Seguridad

### **Autenticación Obligatoria**
- ✅ Usuario debe iniciar sesión antes de subir archivos
- ✅ Token vinculado a usuario específico
- ✅ Validación cruzada entre usuario logueado y token

### **Organización por Usuario**
- ✅ Cada usuario tiene su carpeta única en Google Drive
- ✅ Formato: `User_{UserID}_{UserEmail}`
- ✅ Imposible mezclar archivos entre usuarios

### **Tokens Seguros**
- ✅ Tokens únicos de 64 caracteres hexadecimales
- ✅ Fecha de expiración configurable
- ✅ Un solo uso por token
- ✅ Tracking de quién creó cada token

## 📧 Sistema de Emails

### **Configuración SMTP**
El sistema utiliza las variables de entorno:
```bash
EMAIL_USER=tu-email@gmail.com
EMAIL_PASS=tu-app-password
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_FROM_NAME="Sistema de Gestión de Propiedades"
```

### **Email de Invitación**
- Enviado automáticamente al generar token
- Incluye instrucciones claras
- Lista tipos de archivo permitidos (PDF, JPG, PNG, GIF, WEBP)
- Enlace directo con token embebido

## 🎯 Flujo de Uso Práctico

### **Para Administradores:**

1. **Enviar invitación a usuario:**
   ```
   Dashboard → Gestión de Archivos → Generar Enlace
   - Seleccionar usuario del dropdown
   - Configurar días de validez
   - Enviar (email automático)
   ```

2. **Monitorear tokens:**
   ```
   Dashboard → Gestión de Archivos → Tokens Activos
   - Ver todos los tokens generados
   - Estado: Activo/Usado/Expirado
   - Información del destinatario
   ```

### **Para Usuarios:**

1. **Recibir email de invitación**
2. **Hacer clic en el enlace del email**
3. **Iniciar sesión si no está logueado**
4. **Subir archivos usando drag & drop**
5. **Confirmación de subida exitosa**

## 📁 Estructura de Carpetas en Google Drive

```
📁 Google Drive Root
  └── 📁 User_12345678-1234-1234-1234-123456789abc_juan.perez@email.com
      ├── 📄 contrato_firmado.pdf
      ├── 📄 cedula_identidad.pdf
      └── 🖼️ foto_propiedad.jpg
  └── 📁 User_87654321-4321-4321-4321-987654321def_maria.lopez@email.com
      ├── 📄 comprobante_ingresos.pdf
      └── 🖼️ foto_documento.png
```

## 🔧 Configuración Técnica

### **Backend (Go)**
- Nuevos campos en `UploadToken`: `UserID`, `PersonID`, `CreatedBy`
- Validación de usuario en `HandleUploadFileWithAuth`
- Carpetas específicas por usuario
- Tokens seguros con validación de expiración

### **Frontend (React/TypeScript)**
- Componente de selección de usuario en admin
- Validación de autenticación en FileUploadPage
- Autocompletado de email basado en usuario seleccionado
- Verificación de coincidencia usuario-token

### **Google Drive Integration**
- Service Account para manejo de archivos
- Creación automática de carpetas por usuario
- Permisos y organización centralizada

## 📊 Monitoring y Auditoría

### **Logs del Sistema**
```
✅ [UPLOAD] Archivo: contrato.pdf por usuario juan@email.com en carpeta User_123_juan
✅ [TOKEN CREATED] Enlace generado para juan@email.com por admin@sistema.com
✅ [EMAIL SENT] Invitación enviada a juan@email.com
```

### **Dashboard de Tokens**
- Lista todos los tokens generados
- Estados en tiempo real
- Información del creador (admin/manager)
- Fechas de creación y expiración

## 🚨 Casos de Error Manejados

### **Usuario no autenticado**
- Redirección automática al login
- Preservación del token en URL de retorno
- Mensaje claro de requerimiento de autenticación

### **Token inválido/expirado**
- Validación en backend y frontend
- Mensajes de error específicos
- Logging de intentos de acceso inválidos

### **Usuario incorrecto**
- Validación cruzada usuario-token
- Mensaje indicando usuario autorizado vs actual
- Prevención de acceso no autorizado

### **Archivos no permitidos**
- Validación de tipos MIME
- Lista clara de formatos aceptados
- Mensajes de error específicos por tipo

## 🔄 APIs Disponibles

### **Generar Token de Subida**
```
POST /api/admin/file-upload/generate-link
{
  "recipient_email": "usuario@email.com",
  "recipient_name": "Juan Pérez", 
  "user_id": "uuid-del-usuario",
  "expiration_days": 7
}
```

### **Subir Archivo Autenticado**
```
POST /api/upload/file
Headers: Authorization: Bearer {jwt-token}
FormData: 
  - token: {upload-token}
  - file: {archivo}
```

### **Validar Token**
```
GET /api/upload/validate-token/{token}
Response: {
  "valid": true,
  "name": "Usuario",
  "email": "usuario@email.com",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

## 🎉 Beneficios del Nuevo Sistema

1. **Seguridad Mejorada**: Autenticación obligatoria + validación de tokens
2. **Organización Clara**: Archivos separados por usuario automáticamente  
3. **Auditoría Completa**: Tracking de quién sube qué y cuándo
4. **Experiencia de Usuario**: Proceso simple pero seguro
5. **Administración Eficiente**: Panel centralizado para gestión de tokens
6. **Escalabilidad**: Preparado para múltiples usuarios y archivos

¡El sistema está listo para uso en producción con máxima seguridad y organización! 🚀 