# ImageKit

Servicio de optimización y transformación de imágenes en tiempo real con caché LRU y almacenamiento multi-provider (S3, MinIO/rustfs, GCS).

## Arquitectura

```
┌──────────┐   GET /{slug}/{path}?tr=params   ┌──────────┐
│  Cliente  │ ────────────────────────────────→ │  Imager  │ :9000
└──────────┘ ←──────────────────────────────── └────┬─────┘
      │                                               │
      │                                             1. Busca proyecto por slug en cache
      │                                             2. Obtiene imagen desde storage (S3/GCS)
      │                                             3. Transforma con libvips (CGO)
      │                                             4. Cachea resultado en LRU
      │                                             5. Sirve con Cache-Control immutable
      │
      │                                ┌──────────────────────┐
      │                                │   Project Cache      │
      │                                │  (refresh cada 30s) │
      │                                └──────────┬───────────┘
      │                                           │
      │                                ┌──────────▼───────────┐
      │                                │   imagekit DB        │
      │                                │ (image_projects,     │
      │                                │  image_tenants, etc) │
      │                                └──────────────────────┘
      │
      ├── POST /auth/login ──────────→ API :9090 ─→ imagekit DB
      ├── GET  /admin/projects ──────→ API :9090 ─→ imagekit DB
      └── GET  /panel/projects ──────→ API :9090 ─→ imagekit DB
```

## Servicios

| Servicio | Puerto | Descripción |
|----------|--------|-------------|
| **imager** | `:9000` | Sirve y transforma imágenes |
| **api** | `:9090` | API REST para gestión de proyectos, tenants, métricas |
| **panel** | `:8080` | Panel web SvelteKit (admin + tenant) |

## Inicio rápido

```bash
# Variables de entorno
DB_URL=postgresql://user:pass@host:5432/imagekit?sslmode=disable
JWT_SECRET=tu-secreto

docker compose up -d
```

## Configuración

### Variables de entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8080` | Puerto API |
| `IMAGER_HTTP_ADDR` | `:9000` | Puerto imager |
| `DB_URL` | — | PostgreSQL URL (reemplaza DB_HOST/PORT/USER/PASS/NAME) |
| `DB_HOST` | `127.0.0.1` | Host PostgreSQL (si no se usa DB_URL) |
| `DB_PORT` | `5432` | Puerto PostgreSQL |
| `DB_USER` | `linuxer` | Usuario PostgreSQL |
| `DB_PASSWORD` | — | Password PostgreSQL |
| `DB_NAME` | `vendemas_v2` | Base de datos |
| `DB_SSLMODE` | `disable` | SSL mode |
| `DB_MIN_CONNS` | `2` | Pool mínimo |
| `DB_MAX_CONNS` | `10` | Pool máximo |
| `JWT_SECRET` | — | Secreto JWT (requerido) |
| `CACHE_SIZE_MB` | `256` | Tamaño máximo de caché LRU (~20 keys por MB) |
| `CACHE_TTL_SEC` | `86400` | TTL de caché (24h) |
| `TRANSFORM_CONCURRENCY` | `8` | Transformaciones simultáneas máximas (protege la CPU) |
| `ADMIN_USER` | `admin` | Usuario admin inicial (se crea automáticamente) |
| `ADMIN_PASSWORD` | `admin123` | Password admin inicial |
| `CORS_ORIGINS` | `*` | Orígenes CORS |
| `LOG_LEVEL` | `info` | Nivel de log (debug/info/warn/error) |

---

## API de Imágenes (Imager :9000)

### Servir imagen original

```
GET /{slug}/{filepath}
```

**Ejemplo:**
```
GET /vendemas/images/productos/foto.webp
```

Respuesta: `200` con Content-Type detectado automáticamente, `Cache-Control: public, max-age=4096, immutable`, `ETag`.

### Servir imagen transformada

```
GET /{slug}/{filepath}?tr={param1},{param2},...
```

Los parámetros se especifican como `clave-valor` separados por `,` usando el formato `key-value`.

---

## Parámetros de Transformación (`tr`)

### Dimensiones y Redimensionamiento

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `w` | `w-{n}` | Ancho en px | `w-800` |
| `h` | `h-{n}` | Alto en px | `h-600` |
| `cm` | `cm-{mode}` | Crop mode | `cm-pad_resize` |
| `ar` | `ar-{num}-{den}` | Relación de aspecto (se aplica **antes** del resize) | `ar-4-3`, `ar-16-9` |

#### Crop Modes (`cm`)

| Valor | Descripción |
|-------|-------------|
| `force` | Redimensiona forzando a las dimensiones exactas (distorsiona si es necesario) |
| `at_max` | Reduce hasta que quepa dentro de w×h, sin agrandar si es más pequeña |
| `at_least` | Similar a force pero mantiene proporción |
| `max_size_enum` | Similar a at_max |
| `maintain_ratio` | Mantiene proporción |
| `pad_resize` | Escala para que quepa dentro de w×h y rellena con color de fondo |
| `extract` | Extrae una región |
| `crop` | Recorta el centro después de escalar |
| `trim` | Recorta bordes |

**Comportamiento de `pad_resize`:**
1. Escala la imagen para que quepa dentro de w×h (usa la escala menor)
2. Si una dimensión no se especifica, se calcula proporcionalmente
3. Rellena el espacio sobrante con el color `bg`
4. Nunca agranda la imagen (solo escala `≤ 1`)
5. Si la imagen ya es más pequeña que w×h, solo la centra sobre el fondo

**Comportamiento de `at_max`:**
- Escala solo si la imagen excede w×h
- Usa la escala menor para asegurar que quepa completamente
- No agranda ni recorta

### Formato y Calidad

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `f` | `f-{formato}` | Formato de salida | `f-webp`, `f-png`, `f-jpg`, `f-avif`, `f-gif` |
| `q` | `q-{1-100}` | Calidad | `q-80` |
| `pr` | `pr-true` | Progressive (JPEG/PNG) | `pr-true` |
| `lo` | `lo-true` | Lossless (WebP) | `lo-true` |
| `md` | `md-true` | Conservar metadatos | `md-true` |

Si no se especifica `f`, el formato se auto-detecta:
1. Si el cliente acepta AVIF via `Accept` header → avif
2. Si el cliente acepta WebP → webp
3. Si no → usa el `default_format` del proyecto

### Color de Fondo

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `bg` | `bg-{hex}` | Color de fondo (RRGGBB o RRGGBBAA) | `bg-FF0000`, `bg-00000000` (transparente) |

Se usa en `pad_resize` para rellenar el área sobrante, y en rotaciones como color de relleno.

### Rotación

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `rt` | `rt-{deg}` | Rotación en grados | `rt-90`, `rt-180` |

### Efectos

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `e` | `e-{effect}` | Efecto | `e-grayscale`, `e-sharpen` |
| `bl` | `bl-{float}` | Difuminado (Gaussian Blur) | `bl-5.0` |

Efectos disponibles:
- `grayscale` — Convierte a escala de grises
- `sharpen` — Aplica enfoque (sharpness=1.0, sigma=2.0)
- `contrast` — (pendiente implementación)
- `bright` — (pendiente implementación)

### Recorte Automático

| Parámetro | Formato | Descripción | Ejemplo |
|-----------|---------|-------------|---------|
| `t` | `t-{threshold}` | Recorta bordes transparentes/uniformes | `t-10` |

### Superposiciones

| Parámetro | Formato | Descripción |
|-----------|---------|-------------|
| `oi` | `oi-{path}` | Ruta de imagen a superponer |
| `ot` | `ot-{text}` | Texto a superponer |
| `ox` | `ox-{px}` | Posición X |
| `oy` | `oy-{px}` | Posición Y |
| `ow` | `ow-{px}` | Ancho de superposición |
| `oh` | `oh-{px}` | Alto de superposición |

### Bordes y Radios

| Parámetro | Formato | Descripción |
|-----------|---------|-------------|
| `b` | `b-{n}_{color}` | Borde (ancho_color) |
| `l` | `l-{px}` | Radio de borde redondeado |

### Otros

| Parámetro | Formato | Descripción |
|-----------|---------|-------------|
| `pg` | `pg-{n}` | Página (para GIF/PDF multipágina) |

### Ejemplos de uso

```bash
# Redimensionar a 800px de ancho, recortando a 4:3
GET /vendemas/foto.webp?tr=ar-4-3,w-800

# Pad resize a 800x600 con fondo transparente, formato PNG
GET /vendemas/foto.webp?tr=ar-4-3,cm-pad_resize,f-png,bg-00000000,w-800

# Escala de grises, 300px, calidad 90
GET /vendemas/foto.webp?tr=w-300,e-grayscale,q-90

# Rotar 90 grados, formato WebP lossless
GET /vendemas/foto.webp?tr=rt-90,f-webp,lo-true

# Recorte automático de bordes + blur
GET /vendemas/foto.webp?tr=t-5,bl-3.0
```

---

## API REST (API :9090)

### Autenticación

#### Panel (Tenant)

```bash
# Login
POST /api/panel/auth/login
{"email":"tenant@example.com","password":"..."}
# → {"token":"jwt...","tenant_id":1}

# Registro (requiere código de invitación)
POST /api/panel/auth/register
{"name":"Mi Empresa","email":"tenant@example.com","password":"...","invitation_code":"ABC123"}
# → {"token":"jwt...","tenant_id":1}
```

#### Admin

```bash
POST /api/admin/auth/login
{"username":"admin-ik@iathings.com","password":"..."}
# → {"token":"jwt...","role":"admin"}
```

Todas las rutas protegidas usan `Authorization: Bearer <token>`.

#### Refresh Token

```
POST /api/panel/auth/refresh
Authorization: Bearer <token>
# → {"token":"nuevo_jwt..."}
```

### Panel API (`/api/panel/*`) — Requiere JWT (rol tenant)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/panel/projects` | Listar proyectos del tenant |
| `POST` | `/api/panel/projects` | Crear proyecto |
| `GET` | `/api/panel/projects/{id}` | Obtener proyecto |
| `PUT` | `/api/panel/projects/{id}` | Actualizar proyecto |
| `DELETE` | `/api/panel/projects/{id}` | Eliminar proyecto |
| `POST` | `/api/panel/auth/refresh` | Refrescar token |
| `GET` | `/api/panel/account` | Obtener datos de la cuenta |
| `PUT` | `/api/panel/account` | Actualizar nombre de la cuenta |
| `DELETE` | `/api/panel/images/{slug}/*` | Eliminar imagen del storage e invalidar caché |
| `GET` | `/api/panel/projects/{id}/metrics` | Métricas del proyecto |
| `GET` | `/api/panel/projects/{id}/summary` | Resumen del proyecto |
| `GET` | `/api/panel/metrics` | Métricas del tenant |

### Admin API (`/api/admin/*`) — Requiere JWT (rol admin)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/admin/stats` | Estadísticas globales (cantidad de tenants, proyectos, usuarios) |
| `GET` | `/api/admin/tenants` | Listar todos los tenants |
| `GET` | `/api/admin/tenants/{id}` | Obtener tenant con sus proyectos |
| `GET` | `/api/admin/projects` | Listar todos los proyectos activos |
| `POST` | `/api/admin/projects` | Crear proyecto (requiere `tenant_id`) |
| `GET` | `/api/admin/users` | Listar usuarios admin |
| `GET` | `/api/admin/invitations` | Listar invitaciones |
| `POST` | `/api/admin/invitations` | Crear invitación |
| `GET` | `/api/admin/projects/{id}/metrics` | Métricas de cualquier proyecto |
| `GET` | `/api/admin/projects/{id}/summary` | Resumen de cualquier proyecto |
| `GET` | `/api/admin/metrics` | Métricas globales |

### Open Endpoints (sin auth)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/health` | Health check (con DB ping) |
| `GET` | `/metrics` | Métricas Prometheus |
| `GET` | `/openapi.json` | Documentación OpenAPI |

### Estructura del Proyecto (`image_projects`)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | `SERIAL PK` | ID del proyecto |
| `tenant_id` | `INT FK` | Dueño del proyecto |
| `slug` | `VARCHAR UNIQUE` | Identificador único en URLs (`/slug/path`) |
| `name` | `VARCHAR` | Nombre descriptivo |
| `base_path` | `VARCHAR` | Prefijo para rutas de almacenamiento (default `/`) |
| `provider` | `VARCHAR` | `gcs`, `s3`, o `rustfs` (MinIO) |
| `bucket` | `VARCHAR` | Nombre del bucket |
| `region` | `VARCHAR` | Región (us-east-1 por defecto en custom endpoints) |
| `access_key_id` | `TEXT` | Access Key (S3/rustfs) |
| `secret_access_key` | `TEXT` | Secret Key (S3/rustfs) |
| `credentials_json` | `TEXT` | JSON de credenciales GCS |
| `endpoint` | `TEXT` | Endpoint personalizado (MinIO, etc.) |
| `default_quality` | `INT` | Calidad por defecto (80) |
| `default_format` | `VARCHAR` | Formato por defecto (webp) |
| `max_width` | `INT` | Ancho máximo permitido (4096) |
| `max_height` | `INT` | Alto máximo permitido (4096) |
| `rps` | `INT` | Rate limit: requests/sec (100) |
| `burst` | `INT` | Rate limit: burst (200) |
| `is_active` | `BOOLEAN` | Proyecto activo |

## Storage Providers

| Provider | Tipo | Descripción |
|----------|------|-------------|
| `gcs` | `Google Cloud Storage` | Autenticación via `credentials_json` |
| `s3` | `AWS S3` o compatible | Endpoint estándar `s3.amazonaws.com`, virtual-hosted style |
| `rustfs` | `MinIO` o S3-compatible | Endpoint personalizado, path-style (`/bucket/key`) |

Cada provider implementa: `Get`, `Delete`, `TestConnection`.

## Caché

### Project Cache (tenant.ProjectCache)

- Almacena en memoria los proyectos activos con sus conexiones a storage
- Refresca automáticamente cada **30 segundos** desde la DB
- Los proyectos inactivos (`is_active=false`) se eliminan del cache en el próximo refresh

### LRU Cache (cache.LRUCache)

- Usa `hashicorp/golang-lru/v2`
- Tamaño configurable: ~20 entries por MB (`CACHE_SIZE_MB`)
- TTL configurable (`CACHE_TTL_SEC`, default 24h)
- Entradas expiradas se evitan al ser solicitadas
- Soporta invalidación por prefijo (`slug:path`) — usado en delete de imágenes
- Prometheus metrics: hits, misses, evictions, size

### Flujo de una request

```
Request GET /vendemas/logo.webp?tr=w-200
                │
                ▼
  ┌─ ¿Ruta existe? ─→ No → 404
  │
  ▼ Sí
  ┌─ ¿Path traversal? (solo componentes `..`) ─→ Sí → 403
  │
  ▼ No
  ┌─ ¿Proyecto en cache? ─→ No → 404
  │
  ▼ Sí
  ┌─ Clamp params (w≤maxWidth, etc.)
  │
  ▼
  ┌─ ¿En caché LRU? ─→ Sí → Sirve desde caché
  │
  ▼ No
  ┌─ Obtener desde storage (S3/GCS)
  │
  ▼
  ┌─ ¿Tiene params de transformación? ─→ No → Sirve original
  │
  ▼ Sí
  ┌─ Procesar con libvips:
  │   trim → rotate → resize (ar + cm) → effects → export
  │
  ▼ ¿Transformación exitosa?
  │   Sí → Sirve imagen transformada (Cache-Control immutable)
  │   No  → Fallback: sirve imagen original con content-type detectado
  │
  ▼
  Guardar en LRU cache
  │
  ▼
  Sirve imagen (Cache-Control immutable)
```

## Métricas Prometheus

Todas las métricas expuestas en `/metrics` (puerto del servicio):

| Métrica | Tipo | Labels | Descripción |
|---------|------|--------|-------------|
| `imagekit_requests_total` | Counter | project, format, has_transform | Total de requests |
| `imagekit_request_duration_seconds` | Histogram | project, type | Duración de requests |
| `imagekit_transform_duration_seconds` | Histogram | format | Duración de transformaciones |
| `imagekit_storage_get_duration_seconds` | Histogram | provider | Duración de descargas desde storage |
| `imagekit_errors_total` | Counter | project, type | Total de errores |
| `imagekit_cache_hits_total` | Counter | — | Aciertos de caché |
| `imagekit_cache_misses_total` | Counter | — | Fallos de caché |
| `imagekit_cache_evictions_total` | Counter | — | Entradas eliminadas de caché |
| `imagekit_cache_entries` | Gauge | — | Entradas actuales en caché |
| `imagekit_active_projects` | Gauge | — | Proyectos activos en cache |

---

## Notas de implementación

### Validación de Path Traversal

La validación solo bloquea `..` como **componente de directorio** (ej: `../file`, `subdir/../../file`), no como parte del nombre de archivo. Esto permite archivos como `image_v2.a.b..jpg` o `photo (1).png` sin falsos positivos.

La validación también decodifica URLs URL-encoded (`%2f`, `%2e%2e`) antes de verificar.

### Fallback en Transformación

Si la transformación con libvips falla (formato no soportado, decoder faltante, etc.), el servidor sirve la imagen original en vez de retornar HTTP 500. Esto permite que imágenes en formatos exóticos (AVIF sin soporte, etc.) se sirvan sin transformar.

---

## Tests

```bash
# Sin CGO (transformaciones no, pero infraestructura sí)
CGO_ENABLED=0 go test ./... -count=1

# Con CGO (requiere libvips)
go test ./internal/transform/... -count=1
```

Archivos de test:
- `internal/image/handler_test.go` — Test del handler HTTP (routing, cache, path traversal)
- `internal/params/parser_test.go` — Test del parser de parámetros `tr`
- `internal/params/params_test.go` — Test de Params struct
- `internal/middleware/cors_test.go` — Test de CORS middleware

## Docker

```bash
# Construir servicios
docker compose build

# Iniciar todo
docker compose up -d

# Ver logs
docker compose logs -f

# Reiniciar un servicio
docker compose up -d imager
```

### Dockerfiles

| Dockerfile | Descripción |
|------------|-------------|
| `Dockerfile.imager` | Multi-stage, CGO + libvips, multi-arch (amd64/arm64) |
| `Dockerfile.api` | Multi-stage, Go puro sin CGO |
| `Dockerfile.caddy` | Caddy server para panel SvelteKit (static build) |

El imager requiere CGO para libvips (govips). Usa `tonistiigi/xx` para cross-compilation.
