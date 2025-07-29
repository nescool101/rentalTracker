# 🔧 Configuración de Service Account para Google Drive

## ✅ **Opción Recomendada: Service Account (Sin callback URL)**

Ya tienes las credenciales OAuth2, pero para subida de archivos automática es mejor usar **Service Account**.

### 📋 **Pasos para crear Service Account:**

#### 1. Ve a Google Cloud Console
- Abre [Google Cloud Console](https://console.cloud.google.com/)
- Usa la misma cuenta: `nescool10001@gmail.com`
- Selecciona tu proyecto existente (donde creaste las credenciales OAuth2)

#### 2. Habilitar Google Drive API
- Ve a **"APIs y servicios"** > **"Biblioteca"**
- Busca **"Google Drive API"**
- Haz clic en **"Habilitar"**

#### 3. Crear Service Account
- Ve a **"APIs y servicios"** > **"Credenciales"**
- Haz clic en **"Crear credenciales"** > **"Cuenta de servicio"**
- Nombre: `RentalSystem Drive Service`
- ID: `rental-drive-service`
- Descripción: `Service account para subida de archivos al sistema`
- Haz clic en **"Crear y continuar"**

#### 4. Asignar permisos (opcional)
- Puedes omitir los roles por ahora
- Haz clic en **"Continuar"** y luego **"Listo"**

#### 5. Descargar credenciales JSON
- En la lista de cuentas de servicio, haz clic en la que acabas de crear
- Ve a la pestaña **"Claves"**
- Haz clic en **"Agregar clave"** > **"Crear nueva clave"**
- Selecciona **"JSON"**
- Se descargará un archivo JSON automáticamente

#### 6. Guardar el archivo
- Renombra el archivo descargado a: `google_service_account.json`
- Muévelo a la carpeta `backend/` de tu proyecto
- **¡IMPORTANTE!** No subas este archivo a Git (ya está en .gitignore)

### 🎯 **Configuración en tu proyecto:**

#### Opción 1: Archivo en la raíz (recomendado)
```bash
# Archivo: backend/google_service_account.json
# No necesitas variables de entorno adicionales
```

#### Opción 2: Ruta personalizada
```bash
# En tu archivo .env
GOOGLE_SERVICE_ACCOUNT_PATH=/ruta/personalizada/credenciales.json
```

### ✅ **Verificar que funciona:**

1. **Ejecuta el backend:**
```bash
cd backend
go run .
```

2. **Busca este mensaje en los logs:**
```
✅ Servicio de Google Drive inicializado con Service Account
```

3. **Si no aparece, verás instrucciones en los logs:**
```
⚠️  No se encontró archivo de Service Account
ℹ️  Para crear uno:
   1. Ve a Google Cloud Console
   2. APIs y servicios > Credenciales
   3. Crear credenciales > Cuenta de servicio
   4. Descarga el archivo JSON y guárdalo como 'google_service_account.json'
```

### 🔒 **Compartir carpetas (opcional):**

Si quieres que los archivos aparezcan en una carpeta específica de tu Google Drive:

1. Ve a [Google Drive](https://drive.google.com)
2. Crea una carpeta llamada `Documentos Sistema Rental`
3. Haz clic derecho > **"Compartir"**
4. Añade el email del Service Account (aparece en el archivo JSON como `client_email`)
5. Dale permisos de **"Editor"**

### 🎉 **¡Listo!**

Ahora tu sistema puede:
- ✅ Subir archivos directamente a Google Drive
- ✅ Crear carpetas automáticamente
- ✅ Hacer archivos públicos para compartir
- ✅ Funcionar sin interacción del usuario
- ✅ No necesita callback URL ni OAuth2 complejo

### 🔧 **Variables de entorno necesarias:**

```bash
# Email (ya configurado)
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=nescool10001@gmail.com
EMAIL_PASS=bndp fcme oyhh udyz

# Solo si quieres ruta personalizada del Service Account
# GOOGLE_SERVICE_ACCOUNT_PATH=./google_service_account.json
```

### 🆚 **Comparación con OAuth2:**

| Aspecto | Service Account ✅ | OAuth2 ❌ |
|---------|-------------------|------------|
| Callback URL | No necesario | Obligatorio |
| Interacción usuario | No | Sí |
| Configuración | Simple | Compleja |
| Para subidas automáticas | Perfecto | Innecesario |
| Tokens | No expiran | Expiran y hay que renovar |

**Recomendación:** Usa Service Account para este caso de uso. 