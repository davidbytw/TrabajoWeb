# Registro de Libros Leídos

Trabajo de cursada — Programación Web.
Aplicación web para llevar un registro personal de libros leídos.

**Entrega TP2 — Persistiendo el dominio.** Esta entrega define el esquema de la
base de datos y la capa de acceso a datos generada con sqlc. Todavía no se
conecta al servidor web: el objetivo es tener el paquete de base de datos listo
para el próximo práctico.

![Estado actual de la base de datos](base-actual.png)


## Requisitos

| Herramienta | Versión mínima | Por qué |
|---|---|---|
| Go | 1.26.5 | Es la versión declarada en el módulo |
| Docker Engine | 29.7.2 | Levanta PostgreSQL en un contenedor |
| Docker Compose | **v2** 5.5.0 | El flujo usa `docker compose up --wait`, que no existe en el binario v1 `docker-compose` |
| sqlc | 1.31.1 | `sqlc.yaml` usa la configuración `version: "2"` |

**No hace falta tener PostgreSQL instalado**: corre dentro de un contenedor que
se crea y se destruye en cada ejecución de los tests.

**El puerto 5432 tiene que estar libre.** Es el puerto por defecto de
PostgreSQL, así que si hay una instancia local corriendo hay que detenerla antes.

## Cómo ejecutar los tests

Desde la raíz del proyecto:

    make test

Ese comando realiza el ciclo completo, en este orden:

1. Genera el código de acceso a datos con `sqlc generate`.
2. Compila el proyecto.
3. Elimina contenedores y volúmenes de corridas previas.
4. Levanta PostgreSQL y espera a que el `healthcheck` lo dé por sano.
5. Ejecuta los tests.
6. Elimina contenedores y volúmenes.


Hay que ejecutarlo **desde la raíz del proyecto**: las rutas del
`docker-compose.yml` son relativas a la ubicación de ese archivo.

## Otras tareas disponibles

    make up        # levanta la base de datos y espera a que esté sana
    make down      # elimina contenedores y volúmenes
    make generate  # genera el código de acceso a datos con sqlc
    make build     # compila (genera primero)
    make clean     # elimina el entorno y el código generado

## Estructura del proyecto

    .env                     Variables de entorno del contenedor de PostgreSQL
    docker-compose.yml       Entorno de base de datos para desarrollo y pruebas
    Makefile                 Automatización de tareas
    sqlc.yaml                Configuración de sqlc
    go.mod y go.sum          Módulo y dependencias
    db/schema/schema.sql     Definición de la tabla
    db/queries/queries.sql   Consultas CRUD con anotaciones de sqlc
    db_test.go               Pruebas del ciclo CRUD


## El modelo de datos, en resumen

Una única tabla, `books`, donde cada fila es un libro. Tiene título y
autor obligatorios, un estado restringido a `no_leido`, `leyendo` o `leido`, y
tres campos opcionales que se completan al terminar el libro: fecha de lectura,
puntaje de 1 a 5 y notas.


## Decisiones sobre el entorno

**Espera por `healthcheck`.** El contenedor arranca en un segundo pero
PostgreSQL tarda varios más en aceptar conexiones. Se descartó una espera fija
por arbitraria, y un bucle propio por no tener que escribirlo. El `healthcheck`
usa `pg_isready`, que ya viene en la imagen, con `-h 127.0.0.1`: durante la
inicialización PostgreSQL solo escucha en el socket local, así que exigir una
conexión TCP garantiza que el esquema ya se haya aplicado cuando el contenedor
se declara sano. Se usa period de 120s para dar una ventana de testeo a la imagen de docker.

**Esquema aplicado por el entrypoint.** El `schema.sql` se monta en
`/docker-entrypoint-initdb.d/`, una carpeta de la imagen de PostgreSQL donde
todo archivo `.sql` se ejecuta al inicializar la base. Solo corre con el volumen
vacío, que es exactamente el caso en cada ejecución de los tests.

**El volumen y las pruebas.** El volumen `pgdata` conserva los datos entre
ejecuciones normales (`make down` sin `-v`). El flujo de pruebas lo elimina
deliberadamente antes y después, porque cada corrida tiene que partir de una
base limpia y reproducible.

**Credenciales en `.env`** Decidimos agregar un .env sin ser excepcionado en el
.gitignore, decidimos dejarlo simplemente para que corra los testeos y no tengamos que mandar un correo con el .env.


## Integrantes

- Campos Madarieta, Simón
- Herrera, David
- Urquiza, Carlos David
