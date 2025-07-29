# 🚀 GitHub Pages Deployment Setup

Este documento explica cómo configurar el despliegue automático del frontend en GitHub Pages.

## 📋 Configuración Inicial

### 1. **Habilitar GitHub Pages en el Repositorio**

1. Ve a tu repositorio en GitHub
2. Navega a **Settings** → **Pages**
3. En la sección **Source**, selecciona:
   - **Source**: "GitHub Actions"
   - **Branch**: No need to select (GitHub Actions will handle it)

### 2. **Variables de Entorno para Producción**

El frontend está configurado para trabajar con diferentes URLs según el entorno:

- **Desarrollo**: `http://localhost:5173` (con proxy a backend local)
- **Producción**: `https://nescool101.github.io/rentalTracker/` (frontend estático)

### 3. **Backend Configuration**

Para producción, el backend debe:
- Estar desplegado en **Fly.io** (como está configurado actualmente)
- Tener configurado CORS para permitir requests desde GitHub Pages
- URL del backend: configurar en variables de entorno del frontend

## 🔧 Archivos de Configuración

### **`frontend/vite.config.ts`**
```typescript
export default defineConfig({
  // Configure base path for GitHub Pages
  base: process.env.NODE_ENV === 'production' ? '/rentalTracker/' : '/',
  // ... resto de la configuración
})
```

### **`frontend/package.json`**
```json
{
  "homepage": "https://nescao.github.io/rentalTracker",
  "scripts": {
    "deploy": "npm run build && gh-pages -d dist"
  }
}
```

### **`.github/workflows/deploy.yml`**
GitHub Actions workflow que se ejecuta automáticamente cuando:
- Se hace push a `main` branch
- Se modifican archivos en la carpeta `frontend/`

## 🚀 Proceso de Deployment

### **Automático** (Recomendado)
1. Hacer cambios en el frontend
2. Commit y push a la rama `main`
3. GitHub Actions construirá y desplegará automáticamente
4. El sitio estará disponible en: `https://nescool101.github.io/rentalTracker/`

### **Manual** (Alternativo)
```bash
cd frontend
npm run deploy
```

## 🌐 URLs y Configuración

### **Frontend URLs**
- **Desarrollo**: http://localhost:5173
- **Producción**: https://nescool101.github.io/rentalTracker/

### **Backend URLs** 
- **Desarrollo**: http://localhost:8080
- **Producción**: https://rentalfullnescao.fly.dev

### **API Configuration**
En producción, las llamadas API deben apuntar al backend en Fly.io:

```typescript
// src/api/apiService.ts
const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? 'https://rentalfullnescao.fly.dev/api' 
  : '/api';
```

## 🔍 Verificación de Deployment

### **Verificar que funciona:**
1. Ve a: https://nescool101.github.io/rentalTracker/
2. Verifica que la página carga correctamente
3. Verifica que las rutas funcionan (navegación)
4. Verificar que puede conectar con el backend

### **Logs de Deployment:**
- Ve a **Actions** tab en GitHub
- Revisa el workflow "Deploy to GitHub Pages"
- Verifica que no hay errores

## ⚠️ Consideraciones Importantes

### **1. Backend CORS**
El backend debe permitir requests desde GitHub Pages:
```go
// Agregar en backend
c.Header("Access-Control-Allow-Origin", "https://nescool101.github.io")
```

### **2. Rutas del Router**
React Router debe estar configurado con `basename`:
```typescript
// Si es necesario
<BrowserRouter basename="/rentalTracker">
  // ... rutas
</BrowserRouter>
```

### **3. Assets y Recursos**
Todos los assets (imágenes, archivos) deben usar rutas relativas.

## 🎯 Estructura Final

```
Frontend (GitHub Pages): https://nescool101.github.io/rentalTracker/
└── Interfaz de usuario estática
    └── Conecta vía API a ↓

Backend (Fly.io): https://rentalfullnescao.fly.dev
└── API REST + Base de datos
    └── Maneja toda la lógica de negocio
```

## 🚨 Troubleshooting

### **Error: Página en blanco**
- Verificar que `base` en `vite.config.ts` es correcto
- Verificar que `homepage` en `package.json` es correcto

### **Error: Assets no cargan**
- Verificar rutas de assets son relativas
- Verificar build en `dist/` tiene estructura correcta

### **Error: API calls fallan**
- Verificar CORS en backend
- Verificar URL del backend en variables de entorno
- Verificar que backend está desplegado y funcionando

## 💡 Próximos Pasos

1. **Configurar variables de entorno** para API URL en producción
2. **Actualizar CORS** en backend para permitir GitHub Pages
3. **Configurar dominio personalizado** (opcional)
4. **Configurar SSL** (GitHub Pages lo incluye por defecto)

---

**¡Tu frontend ahora se desplegará automáticamente en GitHub Pages!** 🎉 