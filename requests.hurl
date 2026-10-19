# pruebas de la API: se corren con `make requests` o `make test-api` ({{base_url}} lo define docker-compose.yml)
# con Hurl instalado, contra `make run`: hurl --test --variable base_url=http://localhost:8080 requests.hurl
# el id del libro creado se captura y se borra al final: no dependen de lo que ya haya en la tabla
# no se prueban la fecha exacta (la cubre logic/service_test.go con un reloj fijo) ni la concurrencia

# --- POST /books ---

# 1. crear un libro: se recortan los bordes y el estado inicial lo pone el sistema
POST {{base_url}}/books
Content-Type: application/json
{
    "title": "  Rayuela  ",
    "author": "Julio Cortázar"
}
HTTP 201
Content-Type: application/json
[Captures]
book_id: jsonpath "$.id"
[Asserts]
jsonpath "$.id" isInteger
jsonpath "$.title" == "Rayuela"
jsonpath "$.author" == "Julio Cortázar"
jsonpath "$.status" == "no_leido"
jsonpath "$.read_date" == null
jsonpath "$.rating" == null
jsonpath "$.notes" == null

# 2. título vacío
POST {{base_url}}/books
Content-Type: application/json
{
    "title": "",
    "author": "Julio Cortázar"
}
HTTP 400
Content-Type: application/json
[Asserts]
jsonpath "$.error" == "el título no puede estar vacío"

# 3. autor de solo espacios
POST {{base_url}}/books
Content-Type: application/json
{
    "title": "Rayuela",
    "author": "   "
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el autor no puede estar vacío"

# 4. título de 256 caracteres: la lógica lo frena con 400 antes de que el VARCHAR(255) dé un 500
POST {{base_url}}/books
Content-Type: application/json
{
    "title": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "author": "Julio Cortázar"
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el título no puede tener más de 255 caracteres"

# 5. campo que no es del contrato (el estado lo pone el sistema)
POST {{base_url}}/books
Content-Type: application/json
{
    "title": "Rayuela",
    "author": "Julio Cortázar",
    "status": "leido"
}
HTTP 400
[Asserts]
jsonpath "$.error" contains "status"

# 6. cuerpo que no es JSON
POST {{base_url}}/books
Content-Type: application/json
```
esto no es JSON
```
HTTP 400
[Asserts]
jsonpath "$.error" contains "JSON"

# --- GET /books ---

# 7. listar: un array que contiene el libro creado
GET {{base_url}}/books
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$" isCollection
jsonpath "$[*].id" contains {{book_id}}

# --- GET /books/{id} ---

# 8. pedir el libro por su id
GET {{base_url}}/books/{{book_id}}
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$.id" == {{book_id}}
jsonpath "$.title" == "Rayuela"
jsonpath "$.status" == "no_leido"

# 9. id que no existe
GET {{base_url}}/books/2000000000
HTTP 404
Content-Type: application/json
[Asserts]
jsonpath "$.error" == "libro no encontrado"

# 10. id que no es un número
GET {{base_url}}/books/abc
HTTP 400
[Asserts]
jsonpath "$.error" contains "id"

# 11. id que no entra en 32 bits: 400, y no un 404 ni otro libro
GET {{base_url}}/books/99999999999
HTTP 400
[Asserts]
jsonpath "$.error" contains "id"

# --- PUT /books/{id}: transiciones de estado y reglas de los datos ---

# 12. quedarse en no_leido no es una transición permitida
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "no_leido"
}
HTTP 409
[Asserts]
jsonpath "$.error" == "cambio de estado no permitido: de no_leido a no_leido"

# 13. empezar a leerlo: no_leido -> leyendo
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leyendo"
}
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$.id" == {{book_id}}
jsonpath "$.status" == "leyendo"
jsonpath "$.read_date" == null

# 14. quedarse en leyendo no es una transición permitida
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leyendo"
}
HTTP 409
[Asserts]
jsonpath "$.error" == "cambio de estado no permitido: de leyendo a leyendo"

# 15. volver de leyendo a no_leido
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "no_leido"
}
HTTP 409
[Asserts]
jsonpath "$.error" == "cambio de estado no permitido: de leyendo a no_leido"

# 16. puntaje fuera de leido
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leyendo",
    "rating": 4
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el puntaje y las notas solo se cargan al marcar el libro como leído"

# 17. notas fuera de leido
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leyendo",
    "notes": "Voy por la mitad"
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el puntaje y las notas solo se cargan al marcar el libro como leído"

# 18. read_date no es del contrato: la fecha la pone el sistema
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leido",
    "rating": 5,
    "read_date": "2020-01-01"
}
HTTP 400
[Asserts]
jsonpath "$.error" contains "read_date"

# 19. los rechazos anteriores no cambiaron nada
GET {{base_url}}/books/{{book_id}}
HTTP 200
[Asserts]
jsonpath "$.status" == "leyendo"
jsonpath "$.read_date" == null
jsonpath "$.rating" == null
jsonpath "$.notes" == null

# 20. leido sin puntaje
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leido"
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el puntaje es obligatorio al marcar el libro como leído"

# 21. puntaje fuera de rango
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leido",
    "rating": 6
}
HTTP 400
[Asserts]
jsonpath "$.error" == "el puntaje tiene que estar entre 1 y 5"

# 22. terminarlo: leyendo -> leido, con las notas recortadas y la fecha del sistema
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leido",
    "rating": 5,
    "notes": "  Se puede leer en cualquier orden.  "
}
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$.status" == "leido"
jsonpath "$.rating" == 5
jsonpath "$.notes" == "Se puede leer en cualquier orden."
jsonpath "$.read_date" matches /^\d{4}-\d{2}-\d{2}$/

# 23. el cambio quedó guardado
GET {{base_url}}/books/{{book_id}}
HTTP 200
[Asserts]
jsonpath "$.status" == "leido"
jsonpath "$.rating" == 5

# 24. volver de leido a leyendo
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leyendo"
}
HTTP 409
[Asserts]
jsonpath "$.error" == "cambio de estado no permitido: de leido a leyendo"

# 25. volver de leido a no_leido
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "no_leido"
}
HTTP 409
[Asserts]
jsonpath "$.error" == "cambio de estado no permitido: de leido a no_leido"

# 26. releerlo: leido -> leido con otro puntaje y sin notas
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "leido",
    "rating": 4
}
HTTP 200
[Asserts]
jsonpath "$.status" == "leido"
jsonpath "$.rating" == 4
jsonpath "$.notes" == null

# 27. estado que no existe
PUT {{base_url}}/books/{{book_id}}
Content-Type: application/json
{
    "status": "terminado"
}
HTTP 400
[Asserts]
jsonpath "$.error" contains "el estado tiene que ser no_leido, leyendo o leido"

# 28. actualizar un id que no existe
PUT {{base_url}}/books/2000000000
Content-Type: application/json
{
    "status": "leyendo"
}
HTTP 404
[Asserts]
jsonpath "$.error" == "libro no encontrado"

# --- rutas y métodos ---

# 29. método que la ruta de un libro no acepta: 405 con la cabecera Allow
PATCH {{base_url}}/books/{{book_id}}
HTTP 405
Allow: GET, PUT, DELETE
[Asserts]
jsonpath "$.error" == "método no permitido"

# 30. la colección no acepta DELETE: no se borran todos los libros de una vez
DELETE {{base_url}}/books
HTTP 405
Allow: GET, POST
[Asserts]
jsonpath "$.error" == "método no permitido"

# 31. ruta que no existe: 404 en JSON
GET {{base_url}}/books/{{book_id}}/autores
HTTP 404
Content-Type: application/json
[Asserts]
jsonpath "$.error" == "ruta no encontrada"

# --- DELETE /books/{id} ---

# 32. borrar el libro creado en la 1: 204 sin cuerpo
DELETE {{base_url}}/books/{{book_id}}
HTTP 204
[Asserts]
bytes count == 0

# 33. ya no se puede pedir
GET {{base_url}}/books/{{book_id}}
HTTP 404

# 34. borrarlo de nuevo
DELETE {{base_url}}/books/{{book_id}}
HTTP 404
Content-Type: application/json
[Asserts]
jsonpath "$.error" == "libro no encontrado"
