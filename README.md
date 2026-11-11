# Registro de Libros Leídos

Trabajo de cursada — Programación Web.
Aplicación web para llevar un registro personal de libros leídos.

**Entrega TP5 — Renderizado del Lado del Servidor con Templ.** La aplicación
deja de ser una API JSON y pasa a ser una página web renderizada en el
servidor: el servidor de Go arma el HTML completo con [templ](https://templ.guide/),
con los datos ya adentro, y el navegador solo lo muestra. Los libros se cargan,
se cambian de estado y se borran con formularios HTML clásicos, sin JavaScript,
siguiendo el patrón Post/Redirect/Get. Las capas de la entrega anterior
(handler → servicio → repositorio) se mantienen: cambia solo la presentación.


## Requisitos

Probado en Linux. El `Makefile` usa `setsid`, del paquete
util-linux, que viene en todas las distribuciones de Linux.

| Herramienta | Versión mínima | Por qué |
|---|---|---|
| Go | 1.26.5 | Es la versión declarada en el módulo; compila la aplicación y ejecuta el generador de templ |
| sqlc | 1.31.1 | Genera el código de acceso a datos (`sqlc.yaml` usa `version: "2"`) |
| Docker Engine | 29.7.2 | Levanta PostgreSQL y corre Hurl desde su imagen |
| Docker Compose | **v2** 5.5.0 | El flujo usa `docker compose up --wait`, que no existe en el binario v1 `docker-compose` |
| make | — | Automatiza las tareas |

**No hace falta instalar templ, PostgreSQL ni Hurl.** El generador de templ lo
ejecuta el `Makefile` con `go run`, en la misma versión que la biblioteca del
`go.mod` (si templ ya está instalado, no se usa ni se modifica). PostgreSQL y
Hurl corren desde imágenes de Docker con la versión fija (PostgreSQL 16,
Hurl 7.1.0). La primera vez, Go necesita internet para descargar los módulos.

**Los puertos 5432 (base) y 8080 (aplicación) tienen que estar libres.**


## Cómo ejecutar la aplicación

Después de clonar, desde la raíz del proyecto:

    make run-app

Ese comando hace todo, en este orden:

**Tareas previas**

1. Genera el código: el de acceso a datos con `sqlc generate` y el de las
   vistas con `templ generate`.
2. Compila la aplicación (`bin/booklog`).
3. Elimina contenedores y volúmenes de corridas previas.
4. Levanta PostgreSQL en un contenedor y espera a que su `healthcheck` la dé
   por lista.

**Ejecución**

5. Arranca la aplicación en **http://localhost:8080**. La terminal queda
   mostrando el log de cada pedido. Para terminar: **`Ctrl+C`**.

**Tareas posteriores**

6. Al cortar con `Ctrl+C`, elimina contenedores y volúmenes. Corre también si
   la aplicación no pudo arrancar (por ejemplo, con el puerto 8080 ocupado).

Detalles a tener en cuenta:

- Hay que ejecutarlo **desde la raíz del proyecto**: las rutas del
  `docker-compose.yml` y la carpeta `static/` son relativas a esa carpeta.
- **La primera vez**, antes del paso 1, `make` crea el archivo `.env` copiando
  `.env.example` (ver "Credenciales" más abajo). También tarda más: Go descarga
  los módulos y Docker descarga la imagen de PostgreSQL.
- Con un solo `Ctrl+C` alcanza: la limpieza corre en su propia sesión, así que
  apretarlo de nuevo no la corta a la mitad.
- Si se corta **durante las tareas previas** (antes de que arranque la
  aplicación), la limpieza no corre: `make down` borra lo que haya quedado.
  De todos modos, la próxima ejecución empieza borrando todo (paso 3).
- **Cada ejecución empieza con la base vacía** y al terminar se borra el
  volumen: los libros cargados no se conservan entre ejecuciones.


## Pruebas

### Prueba de las rutas

    make test-app

Hace las mismas tareas previas, arranca la aplicación en segundo plano, corre
las 25 peticiones de `requests.hurl` contra ella y después hace las tareas
posteriores (frena la aplicación y elimina contenedores y volúmenes), **aunque
la prueba haya fallado**. En ese caso el comando termina con error.

`requests.hurl` recorre la página como lo haría un usuario: la página vacía,
las hojas de estilos, el alta (con el redirect `303` a `/`), el título de puros
espacios (`400` con el mensaje y el autor conservado), cada cambio de estado,
una transición prohibida (`409`), el puntaje faltante o que no es un número,
la relectura, el borrado y el borrado repetido (`404`), un título que es HTML
(`<script>…</script>`, que se tiene que ver como texto), un id inválido, una
ruta que no existe y los `405`. Empieza con la base vacía y la deja vacía.

**En macOS** el proyecto no está probado y no funciona tal cual: las tareas
posteriores usan `setsid`, que macOS no trae, y `make test-app` además
necesita Docker Desktop 4.34 o posterior, con la sesión iniciada y la opción
*Enable host networking* activada (Settings → Resources → Network), porque el
contenedor de Hurl usa la red de la máquina para llegar a la aplicación (ver
"Decisiones").

### Tests de las reglas del negocio

    make test

Corre `logic/service_test.go`: las reglas de alta, las transiciones de estado,
el puntaje, las notas, la fecha de lectura y los casos de concurrencia. Usan un
repositorio falso en memoria, así que no necesitan la base ni Docker.

### Otras tareas

    make generate   # genera el código con sqlc y templ
    make build      # genera y compila
    make down       # elimina contenedores y volúmenes
    make clean      # elimina el entorno, el código generado y el binario


## La página

Todo pasa en `http://localhost:8080/`: arriba el formulario para agregar un
libro y abajo la tabla con los libros cargados, ordenados por título (o un
mensaje, si todavía no hay ninguno).

| Columna | Qué muestra |
|---|---|
| Título, Autor | Lo que se cargó en el alta |
| Estado | No leído, Leyendo o Leído |
| Fecha de lectura | dd/mm/aaaa, o "Sin fecha" |
| Puntaje | De 1 a 5, o "Sin puntaje" |
| Notas | Las notas, o "Sin notas" |
| Acciones | Solo los cambios de estado que las reglas permiten desde el estado actual, y el borrado |

Acciones según el estado:

| Estado | Acciones |
|---|---|
| No leído | **Empezar a leer**; **Marcar como leído** (con puntaje y notas); **Borrar** |
| Leyendo | **Marcar como leído** (con puntaje y notas); **Borrar** |
| Leído | **Lo volví a leer** (puntaje y notas nuevos, fecha de hoy); **Borrar** |

El formulario de puntaje y notas aparece plegado y se despliega al hacer clic
(con la etiqueta `<details>` de HTML, sin JavaScript). El puntaje se elige de
una lista del 1 al 5.


## Rutas

| Método y ruta | Datos del formulario | Si sale bien | Si falla |
|---|---|---|---|
| `GET /` | — | 200 con la página | 500 |
| `POST /books` | `title`, `author` | 303 a `/` | 400 con la página y el mensaje |
| `POST /books/{id}/status` | `status`; `rating` y `notes` al pasar a `leido` | 303 a `/` | 400, 404 o 409 con la página y el mensaje |
| `POST /books/{id}/delete` | — | 303 a `/` | 404 con la página y el mensaje |
| `GET /static/…` | — | 200 con el archivo (las hojas de estilos) | 404 |

Los formularios se envían como `application/x-www-form-urlencoded`, que es lo
que manda el navegador (por ejemplo, `title=Rayuela&author=Julio+Cort%C3%A1zar`).

| Código | Cuándo |
|---|---|
| 303 | La operación salió bien: el navegador vuelve a pedir `GET /` |
| 400 | Los datos no cumplen una regla, el puntaje no es un número o el id no es válido |
| 404 | El libro no existe (por ejemplo, un doble clic en Borrar), o la ruta no existe |
| 405 | El método no corresponde a la ruta (con la cabecera `Allow`) |
| 409 | El cambio de estado no está permitido, u otro pedido modificó el libro al mismo tiempo |
| 500 | Error interno: el detalle va al log y al usuario solo un mensaje genérico |

Ejemplos con curl (con la aplicación corriendo):

    curl -i http://localhost:8080/
    curl -i http://localhost:8080/books --data-urlencode 'title=Rayuela' --data-urlencode 'author=Julio Cortázar'
    curl -i http://localhost:8080/books/1/status -d 'status=leyendo'
    curl -i http://localhost:8080/books/1/status -d 'status=leido' -d 'rating=5' --data-urlencode 'notes=Muy bueno'
    curl -i -X POST http://localhost:8080/books/1/delete

Con `-d` o `--data-urlencode`, curl ya hace un POST con el cuerpo como
formulario; `--data-urlencode` codifica los espacios y las tildes como lo hace
el navegador.


## Reglas del dominio

- Título y autor son obligatorios: se guardan sin los espacios de los bordes,
  no pueden quedar vacíos y tienen como máximo 255 caracteres.
- Todo libro nuevo empieza en `no_leido`; el estado lo pone el sistema.
- Transiciones permitidas: `no_leido → leyendo`, `no_leido → leido`,
  `leyendo → leido` y `leido → leido` (relectura). Cualquier otra responde 409.
- Al pasar a `leido` el puntaje (1 a 5) es obligatorio, las notas son
  opcionales y la fecha de lectura la pone el sistema (el día de hoy).
- En `no_leido` y `leyendo` no se aceptan puntaje ni notas (400).
- Notas de puros espacios se guardan como "sin notas" (`NULL`).


## Estructura del proyecto

    cmd/server/main.go        Abre la base, arma las capas, registra las rutas y levanta el servidor
    internal/handler/         Capa HTTP: reparte las rutas, lee los formularios, redirige o muestra la página
    views/                    Componentes templ (.templ) y funciones de formato de la página
    static/                   Hojas de estilos (Pico.css y styles.css), servidas como archivos estáticos
    logic/                    Reglas y casos de uso, sin HTTP ni base (y sus tests)
    internal/repository/      Persistencia: envuelve el código de sqlc
    internal/database/        Abre la conexión a partir de DATABASE_URL
    internal/middleware/      Middleware de logging
    db/schema/schema.sql      Definición de la tabla
    db/queries/book.sql       Consultas con anotaciones de sqlc
    db/sqlc/                  Código generado por sqlc (no se versiona)
    requests.hurl             Prueba de las rutas
    docker-compose.yml        PostgreSQL y Hurl
    .env.example              Plantilla de las credenciales (el .env lo crea make)
    Makefile                  Automatización de tareas
    sqlc.yaml                 Configuración de sqlc
    base-actual.png           Diagrama de la tabla

Los componentes de `views/`:

| Archivo | Componente | Qué es |
|---|---|---|
| `layout.templ` | `Layout` | La estructura base (`html`, `head`, `body`) con los links a las hojas de estilos; recibe el contenido como un `templ.Component` |
| `entity_form.templ` | `BookForm` | El formulario de alta |
| `entity_list.templ` | `BookList` | La tabla de libros con las acciones de cada uno |
| `index.templ` | `IndexPage` | La página principal: el layout con el título, el mensaje de error (si hay), el formulario y la lista |

`templ generate` convierte cada `.templ` en un `_templ.go` con el mismo nombre,
que no se versiona (lo genera `make`).

Quién importa a quién:

    handler ──> logic <── repository ──> db/sqlc
       │          ▲
       └──> views ┘

`logic` es el centro y no importa ningún otro paquete del proyecto. Declara la
interfaz `BookRepository` con lo que necesita, y `repository` la implementa.
Por eso las reglas se prueban con un repositorio falso, sin base. Las vistas
reciben los libros de `logic` y no conocen a sqlc. `cmd/server/main.go` es el
único que conoce todas las piezas y las arma de abajo hacia arriba.


## El modelo de datos, en resumen

Una única tabla, `book`, donde cada fila es un libro. Tiene título y autor
obligatorios, un estado (`no_leido`, `leyendo` o `leido`) y tres campos que se
completan al terminar el libro: fecha de lectura, puntaje de 1 a 5 y notas.
No cambió respecto de la entrega anterior.

![Tabla book](base-actual.png)

| Columna | Tipo | Nulo |
|---|---|---|
| `id` | `SERIAL` (PK) | no |
| `title` | `VARCHAR(255)` | no |
| `author` | `VARCHAR(255)` | no |
| `status` | `VARCHAR(20)`, por defecto `no_leido` | no |
| `read_date` | `DATE` | sí |
| `rating` | `INTEGER` | sí |
| `notes` | `TEXT` | sí |

Restricciones (`CHECK` con nombre, en `db/schema/schema.sql`):

| Restricción | Qué exige |
|---|---|
| `book_title_not_blank` | El título no puede ser solo espacios |
| `book_author_not_blank` | El autor no puede ser solo espacios |
| `book_status_valid` | El estado es `no_leido`, `leyendo` o `leido` |
| `book_rating_range` | El puntaje, si está, va de 1 a 5 |
| `book_finished_fields` | En `leido`, fecha y puntaje obligatorios (notas opcionales); en los otros dos estados, fecha, puntaje y notas vacíos |

**Convención de nombres:** la tabla y los tipos van en singular (`book`,
`Book`); las rutas van en plural porque nombran la colección (`/books`).


## Decisiones

**Se mantienen las capas aunque el enunciado describa al handler llamando a
sqlc.** El enunciado pide que el handler de `GET /` obtenga los registros "con
el método List... de sqlc" y que el del formulario guarde "con el método
Create... de sqlc". Hacemos exactamente eso, pero sin que el handler dependa de
sqlc: el handler llama al servicio (`logic`), el servicio al repositorio y el
repositorio a `ListBooks` y `CreateBook` de sqlc. Es una decisión de buena
práctica, para mantener la separación de responsabilidades y la abstracción en
capas que veníamos usando: las reglas del negocio (validación del título y el
autor, transiciones de estado, fecha de lectura, concurrencia) siguen en un
solo lugar y con sus tests, y la presentación se reemplazó sin tocarlas. Tomar
el enunciado literalmente nos parece un retroceso: obligaría a mover esas
reglas al handler, mezclando presentación con lógica del negocio, que es
justamente la dependencia indeseable de que los handlers dependan de sqlc.

**Sin API JSON.** La entrega anterior era una API REST que respondía JSON. Ya
no hay un cliente que la use: la página la arma el servidor, así que los
handlers JSON, los DTO y sus pruebas se sacaron (siguen en la rama `tp3`). La
ruta `POST /books` pasó a ser la del formulario de alta.

**Formularios clásicos y Post/Redirect/Get.** Sin JavaScript, una página solo
puede hacer pedidos con enlaces (GET) y formularios (GET o POST). Cada
formulario hace un POST; si sale bien, el handler responde `303 See Other` con
`Location: /` y el navegador vuelve a pedir la página, que ya muestra el
cambio. Así, recargar la página (F5) repite el GET y no el POST: no se
duplican libros. Usamos `303` y no `302` porque es el código que HTTP define
para "después de un POST, pedí esta otra dirección con GET".

**Cambio de estado y borrado con POST.** Un formulario HTML no puede mandar PUT
ni DELETE (un método que no sea GET o POST se envía como GET), así que cada
acción tiene su ruta: `POST /books/{id}/status` y `POST /books/{id}/delete`.
Las reglas no se reescribieron: los handlers llaman a `Update` y `Delete` del
mismo servicio de la entrega anterior. Editar título o autor no estaba en la
API y tampoco está en la página.

**La página solo ofrece lo que las reglas permiten.** Cada fila muestra solo los
cambios de estado permitidos desde su estado actual. La vista se lo pregunta a
`logic.CanTransition`, la misma función que usa el servicio, en vez de repetir
la tabla de transiciones. El servidor valida igual: alguien puede mandar un
POST armado a mano, o desde una pestaña que quedó desactualizada.

**Errores: la página de nuevo, con el mensaje.** Si una operación falla, el
handler no redirige: responde la página completa con el mensaje arriba, en un
recuadro (`<article>`), la lista actualizada desde la base y el código que
corresponde (400, 404 o 409). Si falla el alta, el formulario vuelve con lo que
se había escrito. El código es para los programas (Hurl, el log) y el cuerpo
para la persona: el navegador muestra la página igual. Para un error interno
(500) se responde un texto genérico, sin volver a armar la página (que también
necesitaría la base), y el detalle va al log.

**La validación del navegador no reemplaza a la del servidor.** Los campos
tienen `required` y `maxlength`, y el puntaje se elige de una lista del 1 al 5:
el navegador avisa antes de enviar, pero se puede saltear (con curl, o editando
la página) y `required` acepta un título de puros espacios. Las reglas que
valen son las de `logic`.

**Los formularios mandan texto.** El puntaje llega como `"4"` y el handler lo
convierte a número. Un campo vacío llega como `""` y no como "ausente": el
handler lo convierte en "sin dato" (`nil`) antes de llamar al servicio; si
no, al empezar a leer un libro el servicio rechazaría unas notas vacías.

**Escape automático del HTML.** templ escapa todo lo que se muestra entre
`{ }`: un título como `<script>…</script>` se ve como texto y no se ejecuta en
el navegador (XSS).

**Componentes.** Los tres archivos de componentes (`layout.templ`,
`entity_form.templ` y `entity_list.templ`) más `index.templ` con la página
principal, para que cada archivo tenga una sola responsabilidad. El `Layout`
recibe su contenido como un parámetro `templ.Component` y lo ubica dentro de
`<main>` con `@`. Las funciones de formato (estado legible, fecha) están en
`views/format.go`: son parte de la presentación.

**Los NULL en la página.** `read_date`, `rating` y `notes` pueden ser `NULL`
(los dos primeros siempre que el libro no está leído). En `logic` son punteros
y `nil` es `NULL`: la vista pregunta si el valor existe antes de mostrarlo
("Sin fecha", "Sin puntaje", "Sin notas"). Seguir un puntero `nil` cortaría la
página a la mitad.

**La fecha se muestra en UTC.** La fecha de lectura se guarda como la
medianoche UTC de ese día. Si se mostrara en hora argentina (UTC−3), sería las
21 h del día anterior y la página mostraría un día menos.

**Pico.css servido por la aplicación.** El estilo es
[Pico.css](https://picocss.com/) 2.1.1, en su versión *classless*, que le da
formato a las etiquetas de HTML sin necesidad de clases. Usamos la variante
fluida, en la que el contenido ocupa todo el ancho de la ventana en vez de una
columna centrada, con el color principal *lime*:
`pico.fluid.classless.lime.min.css`. El archivo está en `static/` y lo
sirve la aplicación con `http.FileServer`, en vez de cargarlo desde un CDN: así
la página funciona sin internet, la versión queda fija en el repositorio y no
dependemos de un tercero en cada visita. Es un archivo de terceros (licencia
MIT, aviso en su primera línea): no se edita.

**Modo claro fijo.** Pico sigue por defecto el modo claro u oscuro del sistema
operativo. Con `data-theme="light"` en `<html>`, la página se ve igual en
cualquier máquina.

**Tipografía con serifa, en un CSS propio.** Pico usa la fuente del sistema
operativo. `static/styles.css` cambia esa fuente por una con serifa con una
sola regla, que reemplaza la variable `--pico-font-family` de Pico; por eso se
enlaza después de Pico. Usa Georgia si está instalada y, si no, la fuente con
serifa del sistema. Es el único CSS nuestro: el resto del aspecto es el de
Pico.

**templ con la versión del `go.mod`.** El código que genera templ usa la
biblioteca `github.com/a-h/templ` del `go.mod`, así que el generador tiene que
ser de la misma versión. El `Makefile` lee esa versión con `go list -m` y
ejecuta el generador con `go run github.com/a-h/templ/cmd/templ@<versión>`:
la versión queda escrita en un solo lugar y no hace falta instalar templ. Los
`_templ.go` no se versionan, igual que `db/sqlc`.

**La aplicación corre en la máquina; la base, en Docker.** `make run-app`
compila la aplicación y ejecuta el binario, que se conecta a PostgreSQL por el
puerto publicado del contenedor. Así `Ctrl+C` corta directamente la
aplicación y su log se ve en la terminal. Ya no hay imagen de la aplicación:
se sacaron el `Dockerfile` y el servicio `api` del compose.

**`Ctrl+C` y las tareas posteriores.** `Ctrl+C` le llega a todos los procesos
de la terminal, también a `make`, que por sí solo cortaría sin ejecutar las
tareas posteriores. Por eso la receta de `run-app` usa `trap`: al salir (de
cualquier forma) corre `post-tasks`. El `trap` de `INT` hace falta porque
`/bin/sh` (dash, en Ubuntu y derivados) no ejecuta el `trap` de `EXIT` cuando
lo corta una señal. La limpieza se ejecuta con `setsid`, en otra sesión: un
segundo `Ctrl+C` no le llega y no queda a medias (`docker compose` atiende la
señal aunque el shell la ignore).

**Hurl desde Docker, con la red de la máquina.** Hurl corre desde su imagen,
con la versión fija. Como la aplicación corre en la máquina y no en un
contenedor, el servicio `hurl` usa `network_mode: host`: así `localhost:8080`
dentro del contenedor es la aplicación (sin eso, `localhost` sería el propio
contenedor). En Linux el pedido viaja por la interfaz local y ningún firewall
lo bloquea. En Docker Desktop (macOS, Windows) esa opción hay que activarla
(ver "Pruebas").

**Esperar a la aplicación sin espera activa.** En `make test-app` la aplicación
arranca en segundo plano y tarda unos milisegundos en escuchar. El primer
pedido de `requests.hurl` se reintenta cada medio segundo, hasta 20 veces,
hasta que responde: una consulta periódica con pausa y con límite, como el
`healthcheck` de Docker. Una pausa fija (`sleep`) dependería de la velocidad
de la máquina.

**Una sola tabla para el libro y su lectura.** `book` guarda los datos del
libro (título, autor) y los de su lectura (estado, fecha, puntaje, notas) en
la misma fila: cada libro tiene una sola lectura y alcanza con una tabla.
Separarlas en dos (libro y lectura) queda para una versión futura del
proyecto; en esta entrega cambiaba solo la presentación.

**Validación en dos barreras.** Las reglas se validan en `logic`, que es donde
se arma el mensaje claro para el usuario (400). Además, la tabla tiene `CHECK`
con nombre para cada regla: si la validación de Go fallara, la base igual
rechaza el dato.

**Un tipo de error por clase de respuesta.** Cada regla es un valor distinto
(`ErrEmptyTitle`, `ErrMissingRating`…) del tipo `*logic.ValidationError`, y
los conflictos son del tipo `*logic.ConflictError`. El handler pregunta con
`errors.As` de qué tipo es el error y elige 400 o 409 sin conocer la lista de
reglas: una regla nueva se agrega solo en `logic`.

**Concurrencia en el cambio de estado.** Para validar una transición, el
servicio lee el estado actual y después escribe. Si llegan dos pedidos a la vez
sobre el mismo libro, los dos podrían validar contra el mismo estado viejo
(race condition): por ejemplo, un libro que un pedido acaba de pasar a `leido`
podría terminar en `leyendo`, que es una transición prohibida. Por eso la
escritura es un `UPDATE` condicional (`UpdateBookStatus` en
`db/queries/book.sql`): `... WHERE id = $1 AND status = $6 RETURNING *`, donde
`$6` es el estado que se leyó antes de validar. Si otro pedido cambió el estado
en el medio, no se actualiza nada y se responde 409. Cuando dos `UPDATE`
llegan a la misma fila, PostgreSQL hace esperar al segundo hasta que termine el
primero y después vuelve a evaluar el `WHERE` sobre la fila ya modificada: la
espera la resuelve la base, sin bucles en nuestro código.

**El 404 sale de la misma operación que modifica.** `UpdateBookStatus` y
`DeleteBook` son `:one` con `RETURNING *`: la sentencia devuelve la fila que
tocó, y si no tocó ninguna, el método generado responde `sql.ErrNoRows`, que el
repositorio traduce a "libro no encontrado". Así no hay que consultar antes si
el libro existe.

**Un único handler reparte las rutas.** `BookHandler.ServeHTTP` está registrado
en `/` y recibe todo lo que no es `/static/`. Mira la ruta y el método: `/`
solo con GET, `/books` solo con POST, y `/books/{id}/status` o
`/books/{id}/delete` solo con POST. Cualquier otra ruta da 404 (el patrón `/`
del mux atrapa todas las rutas, así que sin ese chequeo cualquier dirección
mostraría la página). Un 405 lleva la cabecera `Allow`, como exige HTTP.

**Credenciales en un solo lugar.** El `.env` define usuario, contraseña y
nombre de la base. Lo leen `docker-compose.yml` y el `Makefile`, que arma la
variable `DATABASE_URL` con la que arranca la aplicación.

**El `.env` no se versiona; se versiona la plantilla.** El repositorio tiene
`.env.example` y el `Makefile` crea el `.env` copiándolo la primera vez (una
regla de make sin prerrequisitos, así nunca pisa un `.env` que ya existe).
Clonar y ejecutar funciona, y no hay un `.env` con credenciales en el
repositorio.

**Apagado ordenado.** Al recibir `Ctrl+C` (o `SIGTERM`), el servidor deja de
aceptar pedidos nuevos, espera hasta 10 segundos a que terminen los que están
en curso y `run` vuelve normalmente: así se ejecuta el `defer db.Close()` y se
cierra la conexión a la base. Para eso `ListenAndServe` corre en su propia
gorutina y `run` espera lo primero que pase: la señal o un error del servidor
(por ejemplo, el puerto 8080 ocupado), que llega por un canal. `main` llama a
`run() error` en vez de usar `log.Fatal` adentro, porque `log.Fatal` termina
con `os.Exit`, que no ejecuta los `defer`.

**Fecha de lectura.** La pone el sistema con el día de hoy según el reloj del
servidor. El `Makefile` ejecuta la aplicación con
`TZ=America/Argentina/Buenos_Aires` (y el binario incluye la base de zonas
horarias, por si el sistema no la tiene): sin eso usaría la zona de la máquina
y, si fuera UTC, después de las 21 h guardaría el día siguiente. El servicio
recibe el reloj como un campo, así los tests fijan la hora.

**`db/sqlc` no se versiona.** Se genera a partir de `db/schema` y
`db/queries`, que son la única fuente de verdad. `make` lo genera antes de
compilar.

**Espera por `healthcheck`.** El contenedor de la base arranca en un segundo
pero PostgreSQL tarda varios más en aceptar conexiones. El `healthcheck` usa
`pg_isready -h 127.0.0.1`: durante la inicialización PostgreSQL solo escucha
en el socket local, así que exigir una conexión TCP garantiza que el esquema
ya se haya aplicado. `docker compose up --wait` espera a ese `healthcheck`
antes de arrancar la aplicación.

**Esquema aplicado por el entrypoint.** El `schema.sql` se monta en
`/docker-entrypoint-initdb.d/`, una carpeta de la imagen de PostgreSQL donde
todo archivo `.sql` se ejecuta al inicializar la base. Solo corre con el
volumen vacío, que es exactamente el caso en cada ejecución.


## Integrantes

- Campos Madarieta, Simón
- Herrera, David
- Urquiza, Carlos David
