# Catálogo de Endpoints de la API (`gama_api`)

Este archivo documenta los endpoints HTTP implementados en la API de Go y sirve como referencia para pruebas en Postman, Bruno o Insomnia.

---

## 🔑 Autenticación

### 1. Iniciar Sesión (`POST /api/v1/auth/login`)
*   **Acceso**: Público
*   **Descripción**: Recibe las credenciales y devuelve un token JWT firmado con algoritmo asimétrico **RS256** (llave privada RSA), válido por 8 horas.
*   **Cuerpo (JSON)**:
    ```json
    {
      "user_nick": "saalcazar",
      "password": "password123"
    }
    ```
*   **Respuesta Exitosa (200 OK)**:
    ```json
    {
      "token": "eyJhbGciOiJSUzI1Ni...",
      "user": {
        "user": {
          "id": "b1b2c3d4-...",
          "user_name": "Samuel Alejandro Alcazar",
          "user_nick": "saalcazar",
          "user_principal_role": "ADMIN"
        },
        "permissions": ["USUARIO_VER", "USUARIO_CREAR", "AUDITORIA_VER"]
      }
    }
    ```

### 2. Obtener Perfil Actual (`GET /api/v1/auth/me`)
*   **Acceso**: Protegido (Requiere cabecera `Authorization: Bearer <token>`)
*   **Descripción**: Retorna la información y lista de permisos del usuario autenticado a partir del contenido del token.
*   **Respuesta Exitosa (200 OK)**:
    ```json
    {
      "id": "b1b2c3d4-...",
      "user_nick": "saalcazar",
      "role": "ADMIN",
      "permissions": ["USUARIO_VER", "USUARIO_CREAR", "AUDITORIA_VER"]
    }
    ```

---

## 🏢 Departamentos

### 3. Listar Departamentos (`GET /api/v1/departments`)
*   **Acceso**: Protegido (Requiere cabecera `Authorization: Bearer <token>`)
*   **Descripción**: Retorna el listado jerárquico de todos los departamentos registrados, ordenados por nivel y nombre.

### 4. Crear Departamento (`POST /api/v1/departments`)
*   **Acceso**: Protegido (Requiere rol `ADMIN` en el token)
*   **Descripción**: Crea una nueva área en el organigrama municipal. Si se asigna un `parent_department_id`, el nivel jerárquico se calcula automáticamente (`level = parent.level + 1`).
*   **Cuerpo (JSON)**:
    ```json
    {
      "name": "Unidad de Desarrollo Tecnológico",
      "sigla": "UDT",
      "parent_department_id": "d0000000-0005-4000-8000-000000000002" 
    }
    ```
    *(Nota: `parent_department_id` puede ser nulo o vació para departamentos principales de Nivel 1).*

---

## 👤 Usuarios

### 5. Listar Usuarios (`GET /api/v1/users`)
*   **Acceso**: Protegido (Requiere permiso `USUARIO_VER`)
*   **Descripción**: Retorna el listado de todos los funcionarios del municipio.

### 6. Ver Detalles de Usuario (`GET /api/v1/users/{id}`)
*   **Acceso**: Protegido (Requiere permiso `USUARIO_VER`)
*   **Descripción**: Obtiene los datos detallados de un usuario específico por su UUID.

### 7. Registrar Nuevo Usuario (`POST /api/v1/users`)
*   **Acceso**: Protegido (Requiere permiso `USUARIO_CREAR`)
*   **Descripción**: Registra un nuevo funcionario. Valida que el departamento exista, encripta la contraseña usando `bcrypt` antes de persistirla y define `active = true` por defecto.
*   **Cuerpo (JSON)**:
    ```json
    {
      "user_name": "Diana Martinez Soto",
      "user_ci": "8877665",
      "user_email": "diana.martinez@municipio.gob.bo",
      "user_phone": "+591 76543210",
      "department_id": "d0000000-0005-4000-8000-000000000002",
      "charge": "Responsable de Infraestructura",
      "user_nick": "dmartinez",
      "password": "mi_password_segura",
      "user_principal_role": "TECNICO"
    }
    ```

### 8. Editar Usuario (`PUT /api/v1/users/{id}`)
*   **Acceso**: Protegido (Requiere permiso `USUARIO_EDITAR`)
*   **Descripción**: Modifica la información básica, cargo, rol, departamento o estado de activación de un usuario existente.
*   **Cuerpo (JSON)**:
    ```json
    {
      "user_name": "Diana Martinez Soto",
      "user_ci": "8877665",
      "user_email": "diana.soto@municipio.gob.bo",
      "user_phone": "+591 76543211",
      "department_id": "d0000000-0005-4000-8000-000000000002",
      "charge": "Jefa de Infraestructura TI",
      "user_nick": "dmartinez",
      "user_principal_role": "DIRECTOR",
      "active": true,
      "requires_password_change": false
    }
    ```

### 9. Dar de Baja / Desactivar Usuario (`DELETE /api/v1/users/{id}`)
*   **Acceso**: Protegido (Requiere permiso `USUARIO_DESACTIVAR`)
*   **Descripción**: Realiza una desactivación lógica de la cuenta del usuario (pone `active = false`). No borra el registro de la base de datos para mantener la trazabilidad e integridad referencial.

---

## 🔍 Auditoría (Pruebas)

### 10. Consultar Logs de Auditoría (`GET /api/v1/admin/audit`)
*   **Acceso**: Protegido (Requiere permiso `AUDITORIA_VER`)
*   **Descripción**: Ruta de prueba del sistema de autorización para corroborar que el middleware de permisos detallados restringe las solicitudes correctamente.

---

## 📑 Solicitantes / Interesados

### 11. Registrar Solicitante (`POST /api/v1/applicants`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_CREAR`)
*   **Descripción**: Registra un nuevo ciudadano o empresa demandante en el sistema.
*   **Cuerpo (JSON)**:
    ```json
    {
      "full_name": "Empresa Comercializadora Los Andes S.R.L.",
      "ci_nit": "10293847012",
      "email": "contacto@losandes.bo",
      "phone": "+591 78901234"
    }
    ```

### 12. Listar Solicitantes (`GET /api/v1/applicants`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_VER_BANDEJA`)
*   **Descripción**: Retorna la lista completa de personas e interesados registrados.

---

## 📜 Hojas de Ruta y Recorrido del Expediente

### 13. Crear Nueva Hoja de Ruta (`POST /api/v1/roadmaps`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_CREAR`)
*   **Descripción**: Registra una nueva Hoja de Ruta. Genera automáticamente el número correlativo anual (ej. `HR-0002/2026`), crea el primer paso de derivación inicial y permite asociar un solicitante existente (`applicant_id`) o registrar uno nuevo inline (`new_applicant`).
*   **Cuerpo (JSON)**:
    ```json
    {
      "procedure_code": "LIC-2026-05",
      "pages_count": 8,
      "subject": "Solicitud de Licencia de Funcionamiento para Actividad Comercial de Restobar",
      "priority": "ALTA",
      "new_applicant": {
        "full_name": "Empresa Comercializadora Los Andes S.R.L.",
        "ci_nit": "10293847012",
        "email": "contacto@losandes.bo",
        "phone": "+591 78901234"
      },
      "destination_department_id": "d0000000-0004-4000-8000-000000000002",
      "instruction": "Remítase a la DAF para verificación de pago de patentes municipales."
    }
    ```

### 14. Listar Hojas de Ruta Visibles (`GET /api/v1/roadmaps`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_VER_BANDEJA`)
*   **Descripción**: Devuelve los trámites visibles según el rol del usuario autenticado. Si el usuario es Alcalde, Admin o Sec. General (`TRAMITE_VER_TODOS`), retorna todos los trámites del municipio; de lo contrario, muestra solo los vinculados a su unidad o puesto.

### 15. Consultar Bandeja de Entrada Activa (`GET /api/v1/roadmaps/inbox`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_VER_BANDEJA`)
*   **Descripción**: Muestra los expedientes físicos/digitales pendientes que se encuentran actualmente recepcionados o asignados en la dependencia del usuario.

### 16. Ver Expediente e Historial de Recorrido (`GET /api/v1/roadmaps/{id}`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_VER_BANDEJA`)
*   **Descripción**: Retorna la cabecera completa del trámite junto con su historial ordenado de derivaciones (pasos 1, 2, 3...), con sellos, firmas e instrucciones.

### 17. Derivar Trámite a Otra Dependencia (`POST /api/v1/roadmaps/{id}/movements`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_DERIVAR`)
*   **Descripción**: Cierra el paso de derivación actual estampando fecha/hora de salida y firma del funcionario, y crea el siguiente paso numerado (`step_number + 1`) con estado `PENDIENTE`.
*   **Cuerpo (JSON)**:
    ```json
    {
      "destination_department_id": "d0000000-0005-4000-8000-000000000002",
      "assigned_user_id": "b1b2c3d4-0001-4000-8000-000000000001",
      "instruction": "Se verificó la patente. Derívese a Soporte Técnico para habilitación de cuenta."
    }
    ```

### 18. Actualizar Estado Global de Trámite (`PATCH /api/v1/roadmaps/{id}/status`)
*   **Acceso**: Protegido (Requiere permiso `TRAMITE_RESOLVER`)
*   **Descripción**: Cambia el estado del expediente (`RESUELTO`, `CONCLUIDO`, `ARCHIVADO`, `RECHAZADO`).
*   **Cuerpo (JSON)**:
    ```json
    {
      "status": "CONCLUIDO"
    }
    ```

