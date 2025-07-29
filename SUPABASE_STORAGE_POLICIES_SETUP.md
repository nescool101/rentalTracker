# 🔐 Configuración de Políticas de Supabase Storage

## ⚠️ Problema Identificado

Tu aplicación está utilizando la `SUPABASE_KEY` (clave anónima) para operaciones de storage, pero las políticas de seguridad (Row Level Security - RLS) están bloqueando la creación de buckets y operaciones de archivos.

## ✅ Solución: Configurar Políticas de Storage

### 1. Acceder al Dashboard de Supabase

1. Ve a [supabase.com](https://supabase.com)
2. Inicia sesión y selecciona tu proyecto: `wbnoxgtrahnlskrlhkmy`
3. Ve a **Storage** → **Policies**

### 2. Crear Bucket Manualmente

1. Ve a **Storage** → **Buckets**
2. Clic en **"New Bucket"**
3. Nombre: `uploads`
4. **Importante**: Marcar como **privado** (no público)
5. Clic en **Create bucket**

### 3. Configurar Políticas de Acceso

Ejecuta estas consultas SQL en **SQL Editor**:

```sql
-- Política para permitir insertar archivos (upload)
CREATE POLICY "Allow authenticated uploads" ON storage.objects 
FOR INSERT 
TO authenticated 
WITH CHECK (bucket_id = 'uploads');

-- Política para permitir leer archivos propios
CREATE POLICY "Allow users to view own files" ON storage.objects 
FOR SELECT 
TO authenticated 
USING (bucket_id = 'uploads' AND auth.uid()::text = (storage.foldername(name))[1]);

-- Política para permitir eliminar archivos propios
CREATE POLICY "Allow users to delete own files" ON storage.objects 
FOR DELETE 
TO authenticated 
USING (bucket_id = 'uploads' AND auth.uid()::text = (storage.foldername(name))[1]);

-- Política para permitir actualizar archivos propios
CREATE POLICY "Allow users to update own files" ON storage.objects 
FOR UPDATE 
TO authenticated 
USING (bucket_id = 'uploads' AND auth.uid()::text = (storage.foldername(name))[1]);
```

### 4. Política para Administradores

Si necesitas que los administradores puedan ver todos los archivos:

```sql
-- Política para administradores (acceso completo)
CREATE POLICY "Allow admin full access" ON storage.objects 
FOR ALL 
TO authenticated 
USING (
  bucket_id = 'uploads' AND 
  EXISTS (
    SELECT 1 FROM auth.users 
    WHERE auth.users.id = auth.uid() 
    AND auth.users.raw_user_meta_data->>'role' = 'admin'
  )
);
```

## 🔄 Solución Alternativa: Usar Service Role Key

Si prefieres no configurar políticas, puedes usar el service role key:

### Opción A: Service Role Key (Mayor Seguridad)

1. En tu dashboard de Supabase, ve a **Settings** → **API**
2. Copia el **Service Role Key** (secret)
3. Agrega a tu archivo `.env`:

```env
SUPABASE_SERVICE_ROLE_KEY=tu-service-role-key-aqui
```

4. Modifica `backend/service/supabase_storage_service.go`:

```go
// Intentar usar service role key primero, luego anon key como fallback
apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
if apiKey == "" {
    apiKey = os.Getenv("SUPABASE_KEY")
    if apiKey == "" {
        return fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY o SUPABASE_KEY debe estar configurada")
    }
}
```

## 🧪 Probar la Configuración

Después de configurar las políticas:

```bash
cd backend
go run .
```

Deberías ver:
```
✅ Servicio de Supabase Storage inicializado
📦 Bucket: uploads
🌐 Storage URL: https://wbnoxgtrahnlskrlhkmy.supabase.co/storage/v1
```

## 🎯 Estructura de Archivos

Con las políticas configuradas, los archivos se organizarán así:

```
uploads/
├── user_123e4567-e89b-12d3-a456-426614174000/
│   ├── 123e4567_1672531200_document.pdf
│   └── 123e4567_1672531300_image.jpg
└── user_456f7890-f12c-34e5-b678-789123456789/
    └── 456f7890_1672531400_file.zip
```

## 🔒 Consideraciones de Seguridad

### Con SUPABASE_KEY (anon):
- ✅ Más seguro (RLS aplicado)
- ✅ Archivos organizados por usuario
- ✅ Solo usuarios autenticados pueden subir
- ✅ Cada usuario solo ve sus archivos

### Con SUPABASE_SERVICE_ROLE_KEY:
- ⚠️ Acceso completo (bypassa RLS)
- ⚠️ Requiere manejo cuidadoso
- ✅ No requiere configuración de políticas
- ⚠️ Nunca exponer en frontend

## 🤔 Recomendación

**Opción 1 (Políticas + SUPABASE_KEY)** es más segura y apropiada para producción.

**Opción 2 (SERVICE_ROLE_KEY)** es más rápida de configurar pero menos segura.

Para tu caso específico, recomiendo usar **Opción 1** ya que tu aplicación maneja archivos de usuarios y necesita control de acceso adecuado. 