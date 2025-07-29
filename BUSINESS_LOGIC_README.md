# 🏢 Sistema de Gestión de Arrendamientos - Lógica de Negocio

## 📋 Resumen Ejecutivo

Este es un **Sistema Integral de Gestión de Arrendamientos** diseñado específicamente para el mercado colombiano. El sistema automatiza todo el ciclo de vida del arrendamiento, desde la generación de contratos legales hasta el seguimiento de pagos y mantenimiento de propiedades.

## 🎯 Enfoque Principal del Sistema

### Propósito Central
Gestionar de manera integral propiedades en arrendamiento con énfasis en:
- **Automatización legal**: Generación de contratos colombianos conforme a la Ley 820 de 2003
- **Gestión financiera**: Seguimiento de pagos, recordatorios automáticos, control de mora
- **Firma digital**: Proceso completo de firma electrónica de contratos
- **Multi-tenant**: Soporte para múltiples propietarios y administradores

### Casos de Uso Principal
1. **Propietarios** gestionan múltiples propiedades
2. **Administradores** manejan carteras de propiedades para terceros
3. **Inquilinos** acceden a información de pagos y solicitudes de mantenimiento
4. **Proceso automatizado** de contratos y recordatorios de pago

## 🏗️ Arquitectura del Sistema

### Stack Tecnológico
- **Backend**: Go con Gin Framework
- **Base de Datos**: Supabase (PostgreSQL)
- **Almacenamiento**: Supabase Storage
- **Autenticación**: JWT tokens
- **Email**: ProtonMail/Gmail SMTP
- **PDF**: Generación y firma digital con certificados
- **Backup**: Telegram Bot para respaldo de archivos

### Estructura de Módulos
```
backend/
├── model/          # Entidades de dominio
├── controller/     # Endpoints REST API
├── service/        # Lógica de negocio
├── storage/        # Capa de datos (repositorios)
├── middleware/     # Autenticación y autorización
├── auth/           # JWT management
└── config/         # Configuración del sistema
```

## 🧩 Entidades del Sistema

### 👥 Person (Personas)
```go
type Person struct {
    ID       uuid.UUID // Identificador único
    FullName string    // Nombre completo
    Phone    string    // Teléfono de contacto
    NIT      string    // Número de identificación
}
```
**Roles**: Propietarios, Inquilinos, Administradores, Testigos, Codeudores

### 🏢 Property (Propiedades)
```go
type Property struct {
    ID         uuid.UUID   // Identificador único
    Address    string      // Dirección completa
    AptNumber  string      // Número de apartamento
    City       string      // Ciudad
    State      string      // Departamento
    ZipCode    string      // Código postal
    Type       string      // Tipo de propiedad
    ResidentID uuid.UUID   // Inquilino actual
    ManagerIDs []uuid.UUID // Administradores asignados
}
```
**Características**: Soporte multi-administrador, seguimiento de inquilino actual

### 📄 Rental (Arrendamientos)
```go
type Rental struct {
    ID            uuid.UUID    // Identificador único
    PropertyID    uuid.UUID    // Propiedad arrendada
    RenterID      uuid.UUID    // Inquilino
    BankAccountID uuid.UUID    // Cuenta bancaria para pagos
    StartDate     FlexibleTime // Fecha de inicio
    EndDate       FlexibleTime // Fecha de finalización
    PaymentTerms  string       // Términos de pago
    UnpaidMonths  int          // Meses impagos
}
```
**Características**: Fechas flexibles, integración bancaria, control de mora

### 💰 Pricing (Precios)
```go
type Pricing struct {
    ID                   uuid.UUID // Identificador único
    RentalID             uuid.UUID // Arrendamiento asociado
    MonthlyRent          float64   // Canon mensual
    SecurityDeposit      float64   // Depósito de garantía
    UtilitiesIncluded    []string  // Servicios incluidos
    TenantResponsibleFor []string  // Servicios a cargo del inquilino
    LateFee              float64   // Multa por mora
    DueDay               int       // Día de vencimiento
}
```

### 🏦 BankAccount (Cuentas Bancarias)
```go
type BankAccount struct {
    ID            uuid.UUID // Identificador único
    PersonID      uuid.UUID // Propietario de la cuenta
    BankName      string    // Nombre del banco
    AccountType   string    // Tipo de cuenta
    AccountNumber string    // Número de cuenta
    AccountHolder string    // Titular de la cuenta
}
```

### 👤 User (Usuarios del Sistema)
```go
type User struct {
    ID             uuid.UUID // Identificador único
    Email          string    // Email de acceso
    PasswordBase64 string    // Contraseña encriptada
    Role           string    // Rol: admin, manager, resident
    PersonID       uuid.UUID // Vinculación con Person
    Status         string    // Estado: pending, active, disabled
}
```

## 🔄 Flujos de Negocio Principales

### 1. 📋 Gestión de Contratos

#### Generación Automática de Contratos
- **Plantilla Legal Colombiana**: Cumple con Ley 820 de 2003
- **Datos Dinámicos**: Integra información de propiedades, inquilinos, precios
- **Formato Profesional**: PDF con formato legal estándar
- **Información Incluida**:
  - Datos completos de arrendador y arrendatario
  - Descripción detallada del inmueble
  - Términos financieros y condiciones
  - Cláusulas legales requeridas
  - Información de testigos y codeudores

#### Proceso de Firma Digital
```
1. Generación del contrato PDF
2. Creación de solicitud de firma
3. Envío de email al firmante
4. Firma digital con certificado
5. Almacenamiento del documento firmado
6. Notificación de finalización
```

**Características Técnicas**:
- Certificados auto-generados para desarrollo
- Validación de firma con timestamps
- Almacenamiento seguro en Supabase
- Backup automático vía Telegram

### 2. 💳 Gestión de Pagos

#### Sistema de Recordatorios Automáticos
- **Recordatorios Mensuales**: Enviados el día de vencimiento
- **Recordatorios de Aniversario**: Notificación anual para renovación
- **Gestión de Mora**: Seguimiento de pagos atrasados
- **Plantillas Personalizadas**: Emails en español con branding

#### Funcionalidades de Pago
```go
type RentPayment struct {
    ID          uuid.UUID    // Identificador único
    RentalID    uuid.UUID    // Arrendamiento asociado
    PaymentDate FlexibleTime // Fecha de pago
    AmountPaid  float64      // Monto pagado
    PaidOnTime  bool         // Indicador de pago oportuno
}
```

### 3. 🔧 Gestión de Mantenimiento

#### Solicitudes de Mantenimiento
```go
type MaintenanceRequest struct {
    ID          uuid.UUID    // Identificador único
    PropertyID  uuid.UUID    // Propiedad afectada
    RenterID    uuid.UUID    // Inquilino solicitante
    Description string       // Descripción del problema
    RequestDate FlexibleTime // Fecha de solicitud
    Status      string       // Estado: pending, in_progress, completed
    CreatedAt   FlexibleTime // Timestamp de creación
    UpdatedAt   FlexibleTime // Última actualización
}
```

**Estados**: Pendiente, En Progreso, Completado, Cancelado

### 4. 📧 Sistema de Notificaciones

#### Email Automático
- **SMTP Integrado**: ProtonMail y Gmail
- **Plantillas HTML**: Diseño profesional
- **Recordatorios Programados**: Basados en fechas de vencimiento
- **Adjuntos**: Soporte para envío de contratos y documentos

#### Tipos de Notificaciones
1. **Recordatorios de Pago**: Mensuales en fecha de vencimiento
2. **Aniversarios**: Renovación anual de contratos
3. **Firma de Contratos**: Enlaces para firma digital
4. **Mantenimiento**: Actualizaciones de estado
5. **Invitaciones**: Registro de nuevos administradores

### 5. 👥 Gestión de Usuarios y Roles

#### Sistema de Roles
- **Admin**: Acceso completo al sistema
- **Manager**: Gestión de propiedades asignadas
- **Resident**: Acceso a información personal y solicitudes

#### Flujo de Registro
```
1. Invitación vía email
2. Registro con validación
3. Activación manual por admin
4. Asignación de propiedades (managers)
5. Acceso basado en rol
```

## 🛡️ Seguridad y Autenticación

### Autenticación JWT
- **Tokens firmados**: Validación criptográfica
- **Expiración configurable**: Control de sesiones
- **Renovación automática**: UX sin interrupciones
- **Información de contexto**: PersonID, Role, Email

### Autorización por Roles
```go
// Middleware de autenticación
middleware.AuthMiddleware()

// Middleware específico para admin
middleware.AdminMiddleware()
```

### Protección de Datos
- **Contraseñas encriptadas**: Base64 + hashing
- **Certificados SSL**: Comunicación segura
- **Validación de entrada**: Sanitización de datos
- **Auditoría**: Logs de acciones críticas

## 📊 Estado Actual del MVP

### ✅ Funcionalidades Implementadas

#### Gestión de Entidades
- [x] CRUD completo de Personas
- [x] CRUD completo de Propiedades
- [x] CRUD completo de Arrendamientos
- [x] CRUD completo de Usuarios
- [x] Gestión de Precios
- [x] Gestión de Cuentas Bancarias
- [x] Gestión de Pagos
- [x] Solicitudes de Mantenimiento

#### Automatización
- [x] Generación automática de contratos PDF
- [x] Sistema de firma digital
- [x] Recordatorios de pago por email
- [x] Notificaciones de aniversario
- [x] Backup automático de archivos

#### Integración
- [x] Supabase como base de datos
- [x] Supabase Storage para archivos
- [x] Email SMTP (ProtonMail/Gmail)
- [x] Telegram para backup
- [x] JWT para autenticación

### 🔧 Áreas que Requieren Pulimiento

#### 1. Validación de Datos
- [ ] Validaciones robustas en modelos
- [ ] Sanitización de entrada de usuario
- [ ] Manejo de errores más específico
- [ ] Validación de formatos colombianos (NIT, teléfonos)

#### 2. Manejo de Errores
- [ ] Códigos de error estandarizados
- [ ] Mensajes de error informativos
- [ ] Logging estructurado
- [ ] Recuperación de errores transitorios

#### 3. Optimización de Rendimiento
- [ ] Indexación de base de datos
- [ ] Paginación en listados
- [ ] Cache para consultas frecuentes
- [ ] Optimización de queries SQL

#### 4. Testing
- [ ] Tests unitarios para servicios
- [ ] Tests de integración para APIs
- [ ] Tests de carga para performance
- [ ] Mocks para servicios externos

#### 5. Documentación
- [ ] Documentación de API (Swagger)
- [ ] Comentarios en código
- [ ] Manual de usuario
- [ ] Guía de deployment

#### 6. Configuración
- [ ] Variables de entorno estandarizadas
- [ ] Configuración por ambiente
- [ ] Secrets management
- [ ] Health checks

## 🚀 Próximos Pasos para MVP

### Fase 1: Estabilización (2-3 semanas)
1. **Validación de Datos**
   - Implementar validaciones robustas en todos los modelos
   - Agregar validaciones específicas colombianas
   - Mejorar manejo de errores

2. **Testing**
   - Crear tests unitarios para servicios críticos
   - Implementar tests de integración para APIs principales
   - Setup de CI/CD básico

3. **Documentación**
   - Generar documentación de API con Swagger
   - Crear guía de instalación y configuración
   - Documentar flujos de negocio

### Fase 2: Optimización (1-2 semanas)
1. **Performance**
   - Implementar paginación en endpoints de listado
   - Optimizar queries de base de datos
   - Agregar cache para consultas frecuentes

2. **UX Improvements**
   - Mejorar mensajes de error para frontend
   - Estandarizar formatos de respuesta API
   - Optimizar tiempo de respuesta

3. **Seguridad**
   - Audit de seguridad básico
   - Implementar rate limiting
   - Mejorar validación de permisos

## 🤖 Roadmap de Inteligencia Artificial

### Fase 1: Validación de Documentos con IA (3-4 semanas)
#### Objetivo
Implementar IA para validar automáticamente documentos de inquilinos y propiedades, reduciendo la intervención humana y agilizando el proceso de aprobación.

#### Características Propuestas
1. **Extracción de Datos de Documentos**
   ```
   - OCR para extraer texto de cédulas, recibos de sueldo
   - Validación automática de datos personales
   - Verificación de consistencia entre documentos
   - Detección de documentos fraudulentos
   ```

2. **Sistema de Scoring de Riesgo**
   ```
   - Análisis de historial crediticio (si disponible)
   - Evaluación de capacidad de pago
   - Scoring automático de inquilinos
   - Recomendaciones de aprobación/rechazo
   ```

3. **Workflow Humano-IA**
   ```
   - IA realiza primera validación
   - Casos complejos escalados a humanos
   - Dashboard de revisión para administradores
   - Aprobación final manual para casos límite
   ```

#### Implementación Técnica
```go
// Servicio de validación con IA
type AIValidationService struct {
    documentOCR     OCRProvider    // Google Vision, AWS Textract
    fraudDetection  FraudProvider  // Servicio antifraude
    riskScoring     RiskProvider   // Motor de scoring
}

// Resultado de validación
type ValidationResult struct {
    DocumentID    uuid.UUID
    IsValid       bool
    ConfidenceScore float64
    ExtractedData map[string]interface{}
    RiskScore     float64
    RequiresHumanReview bool
    Issues        []ValidationIssue
}
```

#### Flujo de Validación con IA
```
1. Inquilino sube documentos (cédula, ingresos, referencias)
2. IA extrae y valida datos automáticamente
3. Sistema calcula score de riesgo
4. Si score > umbral: aprobación automática
5. Si score medio: revisión humana requerida
6. Si score bajo: rechazo automático con explicación
7. Administrador revisa casos dudosos
8. Aprobación final y generación de contrato
```

### Fase 2: IA para Gestión de Pagos (2-3 semanas)
#### Predicción de Morosidad
- **Análisis predictivo** de patrones de pago
- **Alertas tempranas** de posible mora
- **Estrategias personalizadas** de cobranza
- **Optimización de recordatorios** basada en comportamiento

#### Optimización de Precios
- **Análisis de mercado** automático
- **Sugerencias de precios** basadas en ubicación y características
- **Alertas de ajuste** de canon según inflación
- **Comparación competitiva** automática

### Fase 3: IA Conversacional (4-5 semanas)
#### Chatbot para Inquilinos
- **Consultas automatizadas** sobre pagos y contratos
- **Solicitudes de mantenimiento** vía chat
- **Información de cuenta** instantánea
- **Escalamiento automático** a humanos cuando necesario

#### Asistente Virtual para Administradores
- **Análisis de portfolio** automático
- **Reportes inteligentes** con insights
- **Recomendaciones de gestión** basadas en datos
- **Alertas predictivas** de problemas potenciales

## 📈 Métricas de Éxito para MVP

### Métricas Técnicas
- **Uptime**: > 99.5%
- **Tiempo de respuesta**: < 500ms para APIs
- **Cobertura de tests**: > 80%
- **Errores de producción**: < 1% de requests

### Métricas de Negocio
- **Tiempo de generación de contratos**: < 5 minutos
- **Tiempo de firma digital**: < 10 minutos
- **Reducción de tareas manuales**: > 70%
- **Satisfacción de usuario**: > 4/5

### Métricas de IA (Futuro)
- **Precisión de validación de documentos**: > 95%
- **Reducción de revisión manual**: > 80%
- **Precisión de predicción de mora**: > 85%
- **Tiempo de procesamiento de documentos**: < 2 minutos

## 🔧 Configuración y Deployment

### Variables de Entorno Requeridas
```bash
# Base de datos
SUPABASE_URL=your-supabase-url
SUPABASE_ANON_KEY=your-supabase-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Email
SMTP_HOST=smtp.protonmail.ch
SMTP_PORT=587
SMTP_USERNAME=your-email@protonmail.com
SMTP_PASSWORD=your-password

# JWT
JWT_SECRET=your-super-secret-key

# Telegram (opcional)
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-chat-id

# Certificados (desarrollo)
CERT_PATH=./certs/certificate.crt
KEY_PATH=./certs/private.key
```

### Comandos de Deployment
```bash
# Construcción
go build -o rental-manager

# Ejecución con variables de entorno
./rental-manager

# Docker (opcional)
docker build -t rental-manager .
docker run -p 8080:8080 rental-manager
```

## 📝 Conclusiones

Este sistema representa una solución integral para la gestión de arrendamientos en Colombia, con características únicas como:

1. **Contratos legalmente válidos** generados automáticamente
2. **Firma digital integrada** para proceso sin papel
3. **Automatización completa** de recordatorios y notificaciones
4. **Arquitectura escalable** preparada para múltiples propietarios
5. **Cumplimiento legal** con legislación colombiana

El roadmap de IA posiciona el sistema como una solución de próxima generación que combinará **automatización inteligente** con **supervisión humana**, optimizando tanto la eficiencia operativa como la experiencia del usuario.

### Estado Actual: MVP Funcional ✅
### Próximo Paso: Pulimiento y Estabilización 🔧
### Visión Futura: Sistema Inteligente con IA 🤖 