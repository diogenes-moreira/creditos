# Crédito Villanueva

Sistema de gestión de microcréditos para Argentina. Clientes obtienen cuentas corrientes, líneas de crédito y pueden solicitar préstamos con amortización francesa o alemana. Administradores aprueban créditos, ajustan pagos y gestionan el portafolio con dashboard de KPIs.

## Arquitectura

```
creditos/
├── backend/          # Go (Gin + GORM) - API REST + Swagger
├── frontend/         # React (MUI + TypeScript) - SPA
├── static-site/      # Astro - Landing page
└── docker-compose.yml
```

| Componente | Tecnología | Puerto |
|---|---|---|
| Backend API | Go 1.22, Gin, GORM | `:8080` |
| Frontend | React 18, MUI 5, Vite | `:3000` |
| Landing Page | Astro 4.5 | `:4321` |
| Base de datos | PostgreSQL 16 | `:5432` |
| Swagger UI | swaggo/gin-swagger | `:8080/swagger/index.html` |

## Requisitos previos

- [Docker](https://docs.docker.com/get-docker/) y Docker Compose
- [Go 1.22+](https://go.dev/dl/) (para desarrollo local)
- [Node.js 20+](https://nodejs.org/) (para desarrollo local)
- [Make](https://www.gnu.org/software/make/)

## Inicio rápido con Docker

### 1. Clonar el repositorio

```bash
git clone git@github.com:diogenes-moreira/creditos.git
cd creditos
```

### 2. Configurar variables de entorno

```bash
cp .env.example .env
```

Editar `.env` si es necesario. Los valores por defecto funcionan para desarrollo local.

### 3. Levantar todos los servicios

```bash
make docker-up
```

Esto construye y levanta:
- **PostgreSQL** en `localhost:5432`
- **Backend API** en `localhost:8080`
- **Frontend** en `localhost:3000`

### 4. Cargar datos de prueba

```bash
# Opción 1: con Docker
docker exec -it creditos-backend ./seed

# Opción 2: local (requiere Go y PostgreSQL corriendo)
make seed
```

El seed genera:
- 2 administradores (`admin@creditovillanueva.com` / `admin123`, `admin2@creditovillanueva.com` / `admin123`)
- 50 clientes con datos argentinos (DNI, CUIT válidos)
- 30 líneas de crédito en distintos estados
- 40 préstamos con cuotas y pagos
- Registros de auditoría

### 5. Acceder

| Servicio | URL |
|---|---|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080/api/v1 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Health Check | http://localhost:8080/health |

## Desarrollo local

### Backend

```bash
# Instalar dependencias
cd backend && go mod download

# Copiar y configurar variables
cp ../.env.example ../.env

# Ejecutar servidor (requiere PostgreSQL corriendo)
make run

# Ejecutar tests
make test

# Solo tests de dominio
make test-domain

# Lint
make lint

# Generar documentación Swagger (requiere swag CLI)
# go install github.com/swaggo/swag/cmd/swag@latest
cd backend && swag init -g cmd/server/main.go
```

### Frontend

```bash
# Instalar dependencias
make frontend-install

# Servidor de desarrollo (hot reload)
make frontend-dev

# Build de producción
make frontend-build
```

El frontend en desarrollo corre en `http://localhost:5173` y proxea `/api` al backend.

### Landing Page (Astro)

```bash
cd static-site
npm install
npm run dev    # http://localhost:4321
```

## Variables de entorno

| Variable | Descripción | Default |
|---|---|---|
| `DB_HOST` | Host de PostgreSQL | `localhost` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de PostgreSQL | `creditos` |
| `DB_PASSWORD` | Contraseña de PostgreSQL | `creditos_secret` |
| `DB_NAME` | Nombre de la base de datos | `creditos` |
| `SERVER_PORT` | Puerto del backend | `8080` |
| `GIN_MODE` | Modo de Gin (`debug`/`release`) | `debug` |
| `JWT_SECRET` | Secreto para firmar tokens JWT | `change-me-in-production` |
| `JWT_EXPIRATION_HOURS` | Duración del token en horas | `24` |
| `FIREBASE_PROJECT_ID` | ID del proyecto Firebase (opcional en dev) | - |
| `FIREBASE_CREDENTIALS_FILE` | Credenciales de Firebase (opcional en dev) | - |
| `GCS_BUCKET` | Bucket de Google Cloud Storage (opcional en dev) | - |
| `GCS_CREDENTIALS_FILE` | Credenciales de GCS (opcional en dev) | - |
| `LOCAL_STORAGE_PATH` | Path de almacenamiento local (fallback) | `./storage` |

## API Endpoints

### Públicos

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/health/ready` | Readiness probe |
| `POST` | `/api/v1/auth/register` | Registro de usuario |
| `POST` | `/api/v1/auth/login` | Login (retorna JWT) |

### Cliente (requiere JWT)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/me/profile` | Mi perfil |
| `PUT` | `/api/v1/me/profile` | Actualizar perfil |
| `PUT` | `/api/v1/me/mercadopago` | Configurar MercadoPago |
| `GET` | `/api/v1/me/account` | Mi cuenta corriente |
| `GET` | `/api/v1/me/account/movements` | Movimientos de cuenta |
| `POST` | `/api/v1/loans/simulate` | Simular préstamo |
| `POST` | `/api/v1/me/loans` | Solicitar préstamo |
| `GET` | `/api/v1/me/loans` | Mis préstamos |
| `GET` | `/api/v1/me/loans/:id` | Detalle de préstamo |
| `POST` | `/api/v1/me/loans/:loanId/payments` | Registrar pago |
| `GET` | `/api/v1/me/payments` | Mis pagos |

### Administrador (requiere JWT + rol admin)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/admin/clients` | Listar clientes |
| `GET` | `/api/v1/admin/clients/:id` | Detalle de cliente |
| `GET` | `/api/v1/admin/clients/search` | Buscar clientes |
| `POST` | `/api/v1/admin/clients/:id/block` | Bloquear cliente |
| `POST` | `/api/v1/admin/clients/:id/unblock` | Desbloquear cliente |
| `POST` | `/api/v1/admin/credit-lines` | Crear línea de crédito |
| `GET` | `/api/v1/admin/credit-lines/pending` | Líneas pendientes |
| `POST` | `/api/v1/admin/credit-lines/:id/approve` | Aprobar línea |
| `POST` | `/api/v1/admin/credit-lines/:id/reject` | Rechazar línea |
| `GET` | `/api/v1/admin/loans/pending` | Préstamos pendientes |
| `POST` | `/api/v1/admin/loans/:id/approve` | Aprobar préstamo |
| `POST` | `/api/v1/admin/loans/:id/disburse` | Desembolsar préstamo |
| `POST` | `/api/v1/admin/loans/:id/cancel` | Cancelar préstamo |
| `POST` | `/api/v1/admin/loans/:id/prepay` | Cancelación anticipada |
| `PUT` | `/api/v1/admin/payments/:id/adjust` | Ajustar pago |
| `GET` | `/api/v1/admin/dashboard/portfolio` | Resumen portafolio |
| `GET` | `/api/v1/admin/dashboard/delinquency` | Métricas morosidad |
| `GET` | `/api/v1/admin/dashboard/kpis` | KPIs generales |
| `GET` | `/api/v1/admin/dashboard/trends/disbursements` | Tendencia desembolsos |
| `GET` | `/api/v1/admin/dashboard/trends/collections` | Tendencia cobranzas |
| `GET` | `/api/v1/admin/audit` | Registros de auditoría |

Documentación interactiva completa en Swagger UI: `http://localhost:8080/swagger/index.html`

## Estructura del backend

```
backend/
├── cmd/
│   ├── server/main.go          # Entry point
│   └── seed/main.go            # Generador de datos de prueba
├── docs/                        # Swagger generado (swagger.json, swagger.yaml)
├── pkg/
│   ├── money/                   # Aritmética segura con shopspring/decimal
│   └── validator/               # Validación CUIT/DNI argentino
└── internal/
    ├── domain/
    │   ├── model/               # Entidades con lógica de negocio
    │   └── port/                # Interfaces (repositorios, servicios)
    ├── application/
    │   ├── service/             # Casos de uso
    │   └── dto/                 # Request/Response DTOs + mappers
    └── infrastructure/
        ├── config/              # Configuración por env vars
        ├── persistence/postgres/# Repositorios GORM
        ├── http/gin/            # Handlers, middleware, router
        ├── auth/                # Adaptador JWT (Firebase en producción)
        ├── storage/             # Almacenamiento local (GCS en producción)
        └── pdf/                 # Generación de PDFs
```

## Estructura del frontend

```
frontend/src/
├── api/                 # Cliente Axios, endpoints, tipos TypeScript
├── auth/                # Contexto de autenticación, rutas protegidas
├── components/          # AppLayout, DataTable, StatusBadge, KPICard, etc.
├── i18n/                # Traducciones ES/EN/PT con react-i18next
├── pages/
│   ├── auth/            # Login, Register
│   ├── client/          # Dashboard, Account, Loans, Payments, Profile
│   └── admin/           # Dashboard, Clients, Credits, Loans, Payments, Audit
└── theme.ts             # Tema MUI personalizado
```

## Comandos Make

```bash
make build             # Compilar binarios Go
make run               # Ejecutar backend localmente
make test              # Ejecutar todos los tests
make test-domain       # Tests solo del dominio
make seed              # Cargar datos de prueba
make docker-up         # Levantar con Docker Compose
make docker-down       # Detener servicios Docker
make docker-logs       # Ver logs de Docker
make lint              # Ejecutar go vet
make clean             # Limpiar binarios y volúmenes
make frontend-install  # Instalar deps del frontend
make frontend-dev      # Servidor de desarrollo frontend
make frontend-build    # Build de producción frontend
```

## Internacionalización (i18n)

El frontend soporta tres idiomas:
- Español (es) - por defecto
- English (en)
- Português (pt)

El idioma se detecta automáticamente del navegador. Se puede cambiar desde el selector de idioma en la barra superior.

## Decisiones técnicas

- **shopspring/decimal** para toda aritmética monetaria (nunca `float64`)
- **UUID** como primary key en todas las entidades
- **Soft deletes** via `gorm.DeletedAt`
- **Sin SQL directo** - solo API chainable de GORM, incluso para agregaciones del dashboard
- **Arquitectura hexagonal** - el dominio no depende de infraestructura
- **Auditoría por contexto** - middleware inyecta IP/UserAgent
- **JWT local** en desarrollo, Firebase Auth en producción
- **Almacenamiento local** como fallback de Google Cloud Storage

## Deploy a producción

### Requisitos

- Servidor con Docker y Docker Compose
- PostgreSQL (managed o containerizado)
- Dominio configurado
- (Opcional) Firebase Auth, Google Cloud Storage

### Pasos

1. Clonar el repo en el servidor:
   ```bash
   git clone git@github.com:diogenes-moreira/creditos.git
   cd creditos
   ```

2. Configurar variables de producción:
   ```bash
   cp .env.example .env
   # Editar .env con valores de producción:
   # - JWT_SECRET fuerte y único
   # - GIN_MODE=release
   # - Credenciales reales de PostgreSQL
   # - Firebase y GCS si se usan
   ```

3. Construir y levantar:
   ```bash
   docker-compose up -d --build
   ```

4. Cargar datos iniciales (primera vez):
   ```bash
   docker exec -it creditos-backend ./seed
   ```

5. Verificar:
   ```bash
   curl http://localhost:8080/health
   # {"status":"ok"}
   ```

### Nginx reverse proxy (ejemplo)

```nginx
server {
    listen 80;
    server_name creditovillanueva.com;

    location / {
        proxy_pass http://localhost:3000;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
    }

    location /swagger/ {
        proxy_pass http://localhost:8080;
    }
}
```

## Licencia

MIT
