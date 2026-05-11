# CommandPM — Descripción General del Proyecto (v1)

> Gestor de proyectos y tareas que vive completamente en la terminal. Diseñado para equipos pequeños que prefieren vivir en la línea de comandos sin renunciar a una experiencia visual agradable, notificaciones en tiempo real y conexión con inteligencia artificial.

---

## Índice

1. [Visión General](#1-visión-general)
2. [Zona 1 — Autenticación](#2-zona-1--autenticación)
3. [Zona 2 — Organizaciones](#3-zona-2--organizaciones)
4. [Zona 3 — Equipo y Roles](#4-zona-3--equipo-y-roles)
5. [Zona 4 — Proyectos](#5-zona-4--proyectos)
6. [Zona 5 — Estados y Plantillas](#6-zona-5--estados-y-plantillas)
7. [Zona 6 — Tareas](#7-zona-6--tareas)
8. [Zona 7 — Interfaz de Terminal (TUI)](#8-zona-7--interfaz-de-terminal-tui)
9. [Zona 8 — Notificaciones por WhatsApp](#9-zona-8--notificaciones-por-whatsapp)
10. [Zona 9 — MCP Local (Integración con IA)](#10-zona-9--mcp-local-integración-con-ia)
11. [Zona 10 — Infraestructura y Base de Datos](#11-zona-10--infraestructura-y-base-de-datos)
12. [Stack Tecnológico](#12-stack-tecnológico)
13. [Tareas de Desarrollo — v1](#13-tareas-de-desarrollo--v1)

---

## 1. Visión General

**CommandPM** es una herramienta de gestión de proyectos que funciona exclusivamente desde la terminal. No hay interfaz web ni aplicación móvil: todo ocurre en la línea de comandos, con una experiencia visual interactiva (TUI) construida con Bubble Tea y Lip Gloss.

El programa permite a equipos de trabajo organizar sus proyectos, definir sus propios flujos de estados, asignar tareas, recibir notificaciones por WhatsApp cuando una tarea se completa, y conectar un agente de inteligencia artificial local para generar tareas de forma automática.

---

## 2. Zona 1 — Autenticación

**¿Qué hace?**
Permite a los usuarios identificarse en el sistema usando su cuenta de GitHub o Google, sin necesidad de crear una contraseña propia.

**Comportamiento:**
- Al ejecutar el programa por primera vez, se muestra una pantalla de bienvenida con las opciones de inicio de sesión.
- El usuario elige GitHub o Google y el programa abre el navegador para completar el flujo de autorización OAuth2.
- Una vez autenticado, la sesión queda guardada localmente. En usos posteriores, el login es automático.
- Si es la primera vez que ese usuario entra al sistema, se le asigna automáticamente el rol de **owner** y se le solicita que cree su organización.

**Tecnologías involucradas:** OAuth2, Goth, Viper (para guardar sesión), SQLite/GORM (para persistir el usuario).

---

## 3. Zona 2 — Organizaciones

**¿Qué hace?**
Es el contenedor principal de todo el trabajo: proyectos, tareas, estados y miembros del equipo viven dentro de una organización.

**Comportamiento:**
- Solo el **owner** puede crear y editar la organización.
- Una organización tiene nombre, descripción y un número de WhatsApp del administrador (usado para las notificaciones).
- Los **users** (empleados) se unen a la organización de un owner; ellos no pueden crear ni editar la organización.
- No existe límite de miembros en v1.

**Datos que maneja:** nombre, descripción, número de WhatsApp del administrador, fecha de creación.

---

## 4. Zona 3 — Equipo y Roles

**¿Qué hace?**
Controla quién puede hacer qué dentro de la organización.

**Roles disponibles en v1:**

| Rol   | Descripción |
|-------|-------------|
| `owner` | Creador de la organización. Tiene acceso total: puede gestionar la organización, crear/editar/eliminar proyectos, estados, tareas y miembros. |
| `user`  | Miembro del equipo. Puede crear, editar, mover y comentar tareas. **No puede** crear ni editar la organización. |

**Comportamiento:**
- El owner puede invitar compañeros al equipo usando su correo electrónico (que debe coincidir con el correo de su cuenta de GitHub o Google).
- El owner puede eliminar miembros del equipo.
- El owner puede ver la lista de todos los miembros activos.
- Los users solo ven los proyectos y tareas relacionados con su organización.

---

## 5. Zona 4 — Proyectos

**¿Qué hace?**
Agrupa estados y tareas bajo un nombre común. Equivale a un "tablero" o "proyecto" en otras herramientas.

**Comportamiento:**
- El owner y los users pueden crear proyectos.
- Al crear un proyecto se le pueden asignar los estados que tendrá (tomados de los estados disponibles en la organización).
- Un proyecto tiene nombre, descripción y fecha de inicio opcional.
- El owner puede eliminar un proyecto. Los users no pueden eliminarlo.
- Un proyecto puede editarse (nombre, descripción, estados asociados) en cualquier momento.

**Datos que maneja:** nombre, descripción, fecha de inicio, estados asociados, creador, organización.

---

## 6. Zona 5 — Estados y Plantillas

**¿Qué hace?**
Define el flujo de trabajo: los pasos por los que pasa una tarea desde que se crea hasta que se completa.

**Comportamiento:**
- Los estados son personalizables: se pueden crear, editar y eliminar.
- Cada estado tiene un nombre y un color (para visualizarlo en el TUI).
- Los estados se asignan a un proyecto concreto.

**Sistema de Plantillas:**
Para no empezar desde cero, el sistema incluye plantillas predefinidas que se pueden aplicar al crear un proyecto. En v1 existe una sola plantilla:

**Plantilla "Estándar":**
```
Backlog → In Progress → In Review → Done
```

Las plantillas están diseñadas para ser fácilmente ampliables en versiones futuras: basta con añadir una nueva entrada al catálogo de plantillas.

---

## 7. Zona 6 — Tareas

**¿Qué hace?**
Es el núcleo del programa. Una tarea representa una unidad de trabajo que alguien del equipo debe completar.

**Datos de una tarea:**
- Título
- Descripción
- Estado (vinculado al proyecto)
- Persona asignada (un miembro del equipo)
- Fecha de inicio
- Fecha de finalización (deadline)
- Comentarios (cualquier miembro puede añadir comentarios)
- Proyecto al que pertenece

**Comportamiento:**
- Cualquier miembro puede crear, editar y ver tareas dentro de su organización.
- Solo el owner puede eliminar tareas.
- Una tarea se puede mover de un estado a otro (por ejemplo, de "In Progress" a "In Review") directamente desde el TUI.
- Cuando una tarea se mueve al estado **Done**, se dispara automáticamente una notificación de WhatsApp al número del administrador con el título de la tarea, el nombre del empleado y los comentarios que dejó.
- Se puede asignar o reasignar la persona responsable en cualquier momento.

---

## 8. Zona 7 — Interfaz de Terminal (TUI)

**¿Qué hace?**
Es la "pantalla" del programa. En lugar de una web o una app, toda la interacción ocurre en la terminal con una interfaz visual construida con Bubble Tea, Lip Gloss y Bubbles.

**Pantallas principales:**

| Pantalla | Descripción |
|----------|-------------|
| **Bienvenida / Login** | Muestra las opciones de autenticación (GitHub / Google). |
| **Crear organización** | Formulario que aparece solo si el usuario es nuevo (owner). |
| **Dashboard / Proyectos** | Lista de proyectos de la organización. Permite crear, editar y seleccionar un proyecto. |
| **Tablero de tareas (Kanban)** | Vista principal de un proyecto: columnas por estado, tareas dentro de cada columna. Permite moverse con el teclado. |
| **Detalle de tarea** | Vista de lectura/edición de una tarea: ver descripción, añadir comentarios, cambiar estado, reasignar persona, editar fechas. |
| **Gestión de estados** | Pantalla para crear, editar y eliminar estados de un proyecto. |
| **Gestión de equipo** | Lista de miembros, opción de invitar y eliminar (solo owner). |
| **Configuración de organización** | Editar nombre, descripción y número de WhatsApp (solo owner). |

**Navegación:** todo se maneja con el teclado. Flechas para moverse, Enter para seleccionar, Esc para volver, teclas de atajo visibles en cada pantalla.

---

## 9. Zona 8 — Notificaciones por WhatsApp

**¿Qué hace?**
Cuando un empleado mueve una tarea al estado **Done**, el sistema envía automáticamente un mensaje de WhatsApp al número del administrador de la organización.

**Contenido del mensaje:**
```
✅ Tarea completada: [Título de la tarea]
👤 Completada por: [Nombre del empleado]
📝 Comentarios: [Últimos comentarios del empleado en la tarea]
📅 Fecha: [Fecha y hora de completado]
```

**Tecnología:** Evolution API (servidor de WhatsApp self-hosted que expone una REST API). El programa realiza una llamada HTTP a esa API cuando detecta el cambio de estado a Done.

**Configuración:** el número de WhatsApp del administrador se define en la configuración de la organización. La URL y token de la Evolution API se configuran en el archivo de configuración del programa (gestionado con Viper).

---

## 10. Zona 9 — MCP Local (Integración con IA)

**¿Qué hace?**
Expone un servidor MCP (Model Context Protocol) local que permite a cualquier agente de IA compatible (como Claude) crear tareas en el sistema de forma programática, sin necesidad de usar el TUI.

**Uso típico:**
El owner conecta su asistente de IA al servidor MCP local de CommandPM. Desde el chat del asistente puede pedirle, por ejemplo: *"Crea las tareas del sprint para el proyecto Alpha"* y el agente las crea directamente en el sistema.

**Herramientas MCP que expone v1:**

| Herramienta | Descripción |
|-------------|-------------|
| `create_task` | Crea una nueva tarea en un proyecto. |
| `list_tasks` | Lista las tareas de un proyecto, opcionalmente filtradas por estado. |
| `list_projects` | Lista los proyectos de la organización. |
| `list_members` | Lista los miembros del equipo (para poder asignar tareas). |
| `list_states` | Lista los estados disponibles de un proyecto. |

**Cómo funciona:** el servidor MCP corre localmente en un puerto configurable. Se conecta a la misma base de datos SQLite que usa el TUI, por lo que los datos son siempre consistentes.

---

## 11. Zona 10 — Infraestructura y Base de Datos

**¿Qué hace?**
Es la "fontanería" del programa: gestión de la base de datos, configuración, logs y migraciones.

**Base de datos:** SQLite. Un único archivo local. No requiere instalar ningún servidor externo. Ideal para uso en equipo pequeño o en una sola máquina.

**Modelos principales:**
- `User` (id, nombre, email, proveedor OAuth, rol, organización)
- `Organization` (id, nombre, descripción, número WhatsApp)
- `Project` (id, nombre, descripción, fecha inicio, organización)
- `State` (id, nombre, color, orden, proyecto)
- `Task` (id, título, descripción, estado, asignado, fechas, proyecto)
- `Comment` (id, contenido, tarea, autor, fecha)
- `TeamMember` (usuario, organización, rol)

**Migraciones:** gestionadas con Goose. Al arrancar el programa, verifica si hay migraciones pendientes y las aplica automáticamente.

**Configuración:** archivo `config.yaml` gestionado con Viper. Contiene: ruta de la base de datos, puerto del servidor MCP, URL y token de Evolution API.

**Logging:** Zap para logs estructurados. En modo desarrollo se muestran en consola; en modo producción se escriben en archivo.

---

## 12. Stack Tecnológico

| Tecnología | Rol en el proyecto |
|------------|-------------------|
| **Go** | Lenguaje principal |
| **Bubble Tea** | Framework para la TUI interactiva |
| **Lip Gloss** | Estilos visuales del TUI (colores, bordes, layout) |
| **Bubbles** | Componentes reutilizables del TUI (listas, tablas, inputs, spinners) |
| **Cobra** | Estructura de comandos CLI (`commandpm start`, `commandpm mcp`, etc.) |
| **SQLite + GORM** | Base de datos local y ORM para Go |
| **Goose** | Migraciones de base de datos |
| **OAuth2 + Goth** | Autenticación con GitHub y Google |
| **Viper** | Gestión de configuración (config.yaml + variables de entorno) |
| **Zap** | Logging estructurado |
| **Evolution API** | Envío de notificaciones por WhatsApp |
| **MCP (Go SDK)** | Servidor MCP local para integración con agentes de IA |
| **Testify** | Framework de testing |

---

## 13. Tareas de Desarrollo — v1

Las tareas están agrupadas por zona. El orden sugiere la secuencia de implementación recomendada.

---

### 🏗️ Infraestructura Base

- [ ] Inicializar repositorio Go con estructura de carpetas (`cmd/`, `internal/`, `db/`, `config/`, `mcp/`)
- [ ] Configurar Cobra con el comando principal `commandpm` y subcomandos `start` y `mcp`
- [ ] Configurar Viper para leer `config.yaml` y variables de entorno
- [ ] Configurar Zap para logging en desarrollo y producción
- [ ] Crear la conexión a SQLite con GORM
- [ ] Definir todos los modelos GORM (`User`, `Organization`, `Project`, `State`, `Task`, `Comment`, `TeamMember`)
- [ ] Configurar Goose y escribir las migraciones iniciales para todos los modelos
- [ ] Verificar y aplicar migraciones al arranque del programa

---

### 🔐 Zona 1 — Autenticación

- [ ] Configurar Goth con los proveedores GitHub y Google
- [ ] Implementar el flujo OAuth2: abrir navegador, recibir callback, obtener token
- [ ] Guardar la sesión del usuario en local (usando Viper o un archivo de sesión)
- [ ] Crear lógica de detección de "primer login": si el usuario no existe en la BD, crearlo con rol `owner`
- [ ] Crear lógica de sesión persistente: si ya hay sesión guardada, saltar el login
- [ ] Implementar logout (comando `commandpm logout`)

---

### 🏢 Zona 2 — Organizaciones

- [ ] Implementar `CreateOrganization` (nombre, descripción, número WhatsApp)
- [ ] Implementar `UpdateOrganization` (solo owner)
- [ ] Implementar `GetOrganization` (lectura de datos de la organización actual)
- [ ] Validar que un owner solo puede tener una organización en v1

---

### 👥 Zona 3 — Equipo y Roles

- [ ] Implementar `InviteMember`: buscar usuario por email y asociarlo a la organización con rol `user`
- [ ] Implementar `RemoveMember`: desasociar un miembro de la organización (solo owner)
- [ ] Implementar `ListMembers`: listar todos los miembros activos de la organización
- [ ] Implementar middleware/guardia de roles: verificar rol antes de ejecutar acciones restringidas

---

### 📁 Zona 4 — Proyectos

- [ ] Implementar `CreateProject` (nombre, descripción, fecha inicio, estados iniciales desde plantilla o personalizado)
- [ ] Implementar `UpdateProject` (nombre, descripción)
- [ ] Implementar `DeleteProject` (solo owner; elimina también sus estados y tareas)
- [ ] Implementar `ListProjects` (todos los proyectos de la organización)
- [ ] Implementar `GetProject` (detalle de un proyecto con sus estados)

---

### 🏷️ Zona 5 — Estados y Plantillas

- [ ] Definir el catálogo de plantillas en código (estructura de datos, no base de datos)
- [ ] Implementar la plantilla "Estándar": Backlog, In Progress, In Review, Done
- [ ] Implementar `CreateState` (nombre, color, orden, proyecto)
- [ ] Implementar `UpdateState` (nombre, color, orden)
- [ ] Implementar `DeleteState` (validar que no tenga tareas asociadas o reasignarlas)
- [ ] Implementar `ListStates` de un proyecto ordenados por campo `orden`
- [ ] Implementar `ApplyTemplate`: crear los estados de una plantilla en un proyecto nuevo

---

### ✅ Zona 6 — Tareas

- [ ] Implementar `CreateTask` (título, descripción, estado, asignado, fecha inicio, fecha fin, proyecto)
- [ ] Implementar `UpdateTask` (todos los campos editables)
- [ ] Implementar `DeleteTask` (solo owner)
- [ ] Implementar `MoveTask`: cambiar el estado de una tarea
- [ ] Implementar `AssignTask`: cambiar la persona asignada
- [ ] Implementar `AddComment`: añadir un comentario a una tarea
- [ ] Implementar `ListComments`: obtener todos los comentarios de una tarea
- [ ] Implementar `ListTasks`: listar tareas de un proyecto, con filtro opcional por estado
- [ ] Implementar `GetTask`: detalle completo de una tarea con sus comentarios
- [ ] Implementar el disparador de notificación: detectar cuando una tarea pasa a estado Done y llamar al servicio de WhatsApp

---

### 📱 Zona 8 — Notificaciones por WhatsApp

- [ ] Implementar cliente HTTP para Evolution API
- [ ] Implementar `SendWhatsAppNotification(phone, message string)` 
- [ ] Construir el mensaje de notificación con los datos de la tarea completada
- [ ] Integrar el envío en el flujo de `MoveTask` cuando el destino es el estado Done
- [ ] Manejar errores de envío sin interrumpir el flujo principal (el fallo de notificación no debe bloquear la app)

---

### 🤖 Zona 9 — Servidor MCP Local

- [ ] Configurar el servidor MCP con el SDK de Go
- [ ] Implementar la herramienta `list_projects`
- [ ] Implementar la herramienta `list_states`
- [ ] Implementar la herramienta `list_members`
- [ ] Implementar la herramienta `list_tasks`
- [ ] Implementar la herramienta `create_task`
- [ ] Añadir autenticación básica al servidor MCP (token en config.yaml)
- [ ] Crear el subcomando `commandpm mcp` que inicia el servidor MCP en el puerto configurado
- [ ] Documentar cómo conectar el servidor MCP a Claude u otro agente compatible

---

### 🖥️ Zona 7 — Interfaz TUI

- [ ] Configurar el entrypoint de Bubble Tea y el modelo raíz de la aplicación
- [ ] Implementar el sistema de navegación entre pantallas (router de vistas)
- [ ] Diseñar y aplicar el tema visual global con Lip Gloss (colores, bordes, tipografía)
- [ ] Implementar pantalla: **Bienvenida / Login** (selección de proveedor OAuth)
- [ ] Implementar pantalla: **Crear organización** (formulario inicial para nuevos owners)
- [ ] Implementar pantalla: **Lista de proyectos** (dashboard principal)
- [ ] Implementar pantalla: **Tablero Kanban** (columnas por estado, navegación con teclado, mover tareas)
- [ ] Implementar pantalla: **Detalle de tarea** (lectura, edición inline, comentarios, cambio de estado)
- [ ] Implementar pantalla: **Gestión de estados** (crear, editar, eliminar estados de un proyecto)
- [ ] Implementar pantalla: **Gestión de equipo** (lista de miembros, invitar, eliminar)
- [ ] Implementar pantalla: **Configuración de organización** (solo owner)
- [ ] Implementar barra de atajos de teclado visible en cada pantalla
- [ ] Implementar mensajes de error y confirmación (modales o línea de estado)

---

### 🧪 Testing

- [ ] Configurar Testify en el proyecto
- [ ] Tests unitarios para la capa de servicios (lógica de negocio de tareas, proyectos, estados)
- [ ] Tests unitarios para la validación de roles
- [ ] Tests de integración para los flujos principales (crear proyecto → crear tarea → mover a Done → notificación)
- [ ] Tests unitarios para las herramientas MCP

---

### 📦 Entrega y Documentación

- [ ] Crear `README.md` con instrucciones de instalación, configuración y uso
- [ ] Crear `config.example.yaml` con todos los campos comentados
- [ ] Documentar cómo configurar los proveedores OAuth (GitHub App, Google OAuth)
- [ ] Documentar cómo instalar y conectar Evolution API
- [ ] Documentar cómo conectar el servidor MCP a un agente de IA
- [ ] Crear script de instalación o instrucciones de build (`go build`)
