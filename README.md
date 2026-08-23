# Registro de Libros Leídos

Trabajo de cursada — Programación Web.
Aplicación web para llevar un registro personal de libros leídos.

## El dominio

La aplicación registra libros con los siguientes atributos:

| Atributo | Obligatorio | Observaciones |
|---|---|---|
| Título | Sí | |
| Autor | Sí | |
| Estado | Sí | `pendiente`, `leyendo` o `leido` |
| Fecha de lectura | No | Se completa al marcar como leído |
| Puntaje | No | Se completa al marcar como leído |
| Notas | No | Se completa al marcar como leído |

## Decisiones de diseño

**Instancia del 23-08-2026**

- Título, autor y estado son obligatorios. El resto de los atributos se
  completan recién al marcar un libro como leído.
- El estado admite tres valores fijos: `pendiente`, `leyendo` y `leido`.
- La fecha de lectura todavía no tiene definido su tratamiento temporal
  (formato y validaciones), queda pendiente para la próxima instancia.
- La carga de los datos de finalización se hará por formulario.

**Implementación**

- Se usa `http.FileServer` para servir los archivos estáticos, en lugar de
  escribir la respuesta a mano. Además de resolver el ruteo a archivos y el
  404, deriva el `Content-Type` de la extensión de cada archivo (`.html`,
  `.css`, `.js`), evitando fijar un tipo único para todas las respuestas.
- Se agregó una hoja de estilos mínima en `static/css/`, que sirve además
  para verificar lo anterior: el HTML se entrega como `text/html` y el CSS
  como `text/css`.

## Requisitos

- Go 1.25 o superior

## Cómo ejecutar

Ambas opciones se ejecutan **desde la raíz del proyecto**, ya que el servidor
busca la carpeta `static/` de forma relativa al directorio de trabajo.

### 1. Desde el código fuente

    go run .

### 2. Desde el binario ya compilado

Se incluye un binario compilado en el repositorio como alternativa, por si la
versión de Go disponible fuera anterior a la declarada en `go.mod`:

    ./miapp

Nota: el binario fue compilado para Linux (amd64) con Go 1.25.12, mediante:

    go build -o miapp .

## Acceso

Una vez levantado el servidor, abrir en el navegador:

http://localhost:8080/

## Integrantes

- Campos Madarieta, Simón
- Herrera, David
- Urquiza, Carlos David


