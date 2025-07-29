# 📱 Guía de Pruebas Móviles para el Dashboard

## Cómo Probar el Dashboard en Móvil y Tablet

### 1. Opciones de Prueba

#### Opción A: Simulador en Chrome DevTools
1. Abrir Chrome y ir a `http://localhost:5173/dashboard`
2. Presionar `F12` o `Ctrl+Shift+I` (Windows) / `Cmd+Option+I` (Mac)
3. Hacer clic en el ícono del móvil en la parte superior izquierda de DevTools
4. Seleccionar diferentes dispositivos:
   - **iPhone SE** (375x667) - Móvil pequeño
   - **iPhone 12 Pro** (390x844) - Móvil medio
   - **iPad** (768x1024) - Tablet
   - **iPad Pro** (1024x1366) - Tablet grande
   - **Galaxy S20 Ultra** (412x915) - Android grande

#### Opción B: Prueba Real en Dispositivo
1. Asegurarse de que tu computadora y móvil estén en la misma red WiFi
2. Obtener la IP local de tu computadora:
   ```bash
   # En terminal/cmd:
   ipconfig getifaddr en0    # Mac
   ipconfig                  # Windows
   hostname -I               # Linux
   ```
3. En el móvil, abrir el navegador y ir a: `http://[TU_IP]:5173/dashboard`
   - Ejemplo: `http://192.168.1.100:5173/dashboard`

### 2. Aspectos a Verificar

#### ✅ Navegación
- [ ] El menú hamburguesa funciona correctamente
- [ ] Los enlaces de navegación son fáciles de tocar (mínimo 44px)
- [ ] El menú se cierra automáticamente al seleccionar una opción
- [ ] La navegación lateral es accesible con el dedo

#### ✅ Dashboard de Estadísticas (Admin/Manager)
- [ ] Las tarjetas de estadísticas se apilan correctamente en móvil
- [ ] Los números y títulos son legibles
- [ ] Los gráficos circulares se redimensionan apropiadamente
- [ ] No hay desbordamiento horizontal (scroll horizontal no deseado)

#### ✅ Lista de Accesos Rápidos (Usuario Regular)
- [ ] Los elementos de la lista son fáciles de tocar
- [ ] Los iconos son del tamaño correcto
- [ ] El texto es legible sin hacer zoom
- [ ] Los enlaces funcionan correctamente

#### ✅ Alertas y Notificaciones
- [ ] Las alertas se muestran completamente
- [ ] Los botones dentro de las alertas son accesibles
- [ ] El texto de las alertas no se corta

#### ✅ Formularios y Botones
- [ ] Los botones tienen el tamaño mínimo recomendado (44x44px)
- [ ] Los formularios no causan zoom automático en iOS
- [ ] Los inputs son fáciles de seleccionar y escribir

### 3. Breakpoints Configurados

Hemos configurado los siguientes breakpoints responsivos:

- **Móvil pequeño**: `< 480px` - Una columna, padding reducido
- **Móvil estándar**: `481px - 767px` - Una columna, elementos centrados
- **Tablet**: `768px - 1199px` - Dos columnas, espaciado medio
- **Desktop**: `≥ 1200px` - Cuatro columnas, espaciado completo

### 4. Funcionalidades Específicas Móvil

#### Touch Targets
- Botones mínimo 44x44px
- Enlaces con área de toque extendida
- Espaciado adecuado entre elementos interactivos

#### Optimizaciones iOS
- Font-size de 16px en inputs (previene zoom automático)
- Viewport configurado correctamente
- Touch callouts deshabilitados donde corresponde

#### Optimizaciones Android
- Áreas de toque optimizadas
- Colores de fondo apropiados para modo oscuro
- Navegación compatible con gestos del sistema

### 5. Problemas Comunes y Soluciones

#### Problema: "Se ve muy pequeño en móvil"
**Solución**: Verificar que esté presente en `index.html`:
```html
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
```

#### Problema: "El menú no se abre en móvil"
**Solución**: Verificar que el estado `opened` se esté manejando correctamente en `MainLayout.tsx`

#### Problema: "Los botones son difíciles de tocar"
**Solución**: Verificar que se apliquen los estilos CSS móviles de `index.css`

#### Problema: "Scroll horizontal no deseado"
**Solución**: Verificar que no haya elementos con width fijo que excedan el viewport

### 6. Modo Oscuro en Móvil

El dashboard también soporta modo oscuro. Para probarlo:
1. Hacer clic en el ícono de luna/sol en el header
2. Verificar que todos los elementos se vean correctamente
3. Comprobar contraste suficiente en todos los elementos

### 7. Pruebas de Rendimiento Móvil

#### Network Throttling
1. En Chrome DevTools, ir a la pestaña "Network"
2. Seleccionar "Slow 3G" o "Fast 3G"
3. Verificar que la aplicación cargue en tiempo razonable

#### Performance Testing
1. En DevTools, ir a la pestaña "Lighthouse"
2. Ejecutar audit para "Mobile"
3. Buscar score de 90+ en Performance y Accessibility

### 8. Comandos Útiles

```bash
# Iniciar servidor de desarrollo
npm run dev

# Ver en red local (para pruebas en dispositivos reales)
npm run dev -- --host

# Verificar build de producción
npm run build
npm run preview
```

### 9. Reporte de Bugs Móviles

Al encontrar problemas, incluir:
- Dispositivo y navegador usado
- Tamaño de pantalla
- Screenshot del problema
- Pasos para reproducir
- Comportamiento esperado vs actual

---

## 🎯 Checklist Final de Pruebas Móvil

- [ ] Dashboard carga correctamente en móvil
- [ ] Navegación funciona sin problemas
- [ ] Todos los botones son fácilmente tocables
- [ ] No hay scroll horizontal
- [ ] Texto legible sin zoom
- [ ] Modo oscuro funciona correctamente
- [ ] Formularios funcionan sin zoom automático
- [ ] Performance acceptable en 3G
- [ ] Todas las funcionalidades accesibles

¡Con estas mejoras, tu dashboard debería funcionar perfectamente en móviles y tablets! 🚀 