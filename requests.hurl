# pruebas de las rutas de la pagina (make test-app): empiezan con la base vacia y la dejan vacia
# base_url llega con --variable; los textos se buscan con xpath sobre el HTML de la respuesta


# 1. la pagina responde; se reintenta mientras la aplicacion arranca
GET {{base_url}}/
[Options]
retry: 20
retry-interval: 500
HTTP 200
[Asserts]
header "Content-Type" == "text/html; charset=utf-8"

# 2. la pagina sin libros: el titulo y el mensaje de la lista vacia
GET {{base_url}}/
HTTP 200
[Asserts]
xpath "string(//title)" == "Registro de Libros Leídos"
xpath "normalize-space(//main/section[2]/p)" == "Todavía no cargaste ningún libro."

# 3. las hojas de estilos se sirven como archivos estaticos
GET {{base_url}}/static/pico.fluid.classless.lime.min.css
HTTP 200
[Asserts]
header "Content-Type" == "text/css; charset=utf-8"

GET {{base_url}}/static/styles.css
HTTP 200
[Asserts]
header "Content-Type" == "text/css; charset=utf-8"
body contains "--pico-font-family"


# 4. alta: guarda el libro y redirige a la pagina con 303 (Post/Redirect/Get)
POST {{base_url}}/books
[FormParams]
title: Rayuela
author: Julio Cortázar
HTTP 303
[Asserts]
header "Location" == "/"

# 5. el libro aparece sin leer, con las acciones de no_leido; el id sale del formulario de borrar
GET {{base_url}}/
HTTP 200
[Captures]
book_id: xpath "string(//tr[td[1]='Rayuela']//form[button='Borrar']/@action)" regex "/books/(\\d+)/delete"
[Asserts]
xpath "string(//tr[td[1]='Rayuela']/td[2])" == "Julio Cortázar"
xpath "string(//tr[td[1]='Rayuela']/td[3])" == "No leído"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[4])" == "Sin fecha"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[5])" == "Sin puntaje"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[6])" == "Sin notas"
xpath "count(//tr[td[1]='Rayuela']//button[.='Empezar a leer'])" == 1
xpath "count(//tr[td[1]='Rayuela']//summary[.='Marcar como leído'])" == 1

# 6. alta con un titulo de puros espacios (el required del navegador lo deja pasar): 400,
# la pagina con el mensaje, el autor que se habia escrito y sin guardar nada
POST {{base_url}}/books
Content-Type: application/x-www-form-urlencoded
`title=%20%20%20&author=Borges`
HTTP 400
[Asserts]
xpath "normalize-space(//main/article/strong)" == "el título no puede estar vacío"
xpath "string(//input[@name='author']/@value)" == "Borges"
xpath "count(//tbody/tr)" == 1


# 7. empezar a leer
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leyendo
HTTP 303
[Asserts]
header "Location" == "/"

# 8. ahora esta leyendo: ya no ofrece "Empezar a leer"
GET {{base_url}}/
HTTP 200
[Asserts]
xpath "string(//tr[td[1]='Rayuela']/td[3])" == "Leyendo"
xpath "count(//tr[td[1]='Rayuela']//button[.='Empezar a leer'])" == 0
xpath "count(//tr[td[1]='Rayuela']//summary[.='Marcar como leído'])" == 1

# 9. transicion prohibida, como desde una pagina que quedo desactualizada: 409 con el mensaje
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leyendo
HTTP 409
[Asserts]
xpath "normalize-space(//main/article/strong)" == "cambio de estado no permitido: de leyendo a leyendo"

# 10. marcar como leido sin puntaje: 400
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leido
HTTP 400
[Asserts]
xpath "normalize-space(//main/article/strong)" == "el puntaje es obligatorio al marcar el libro como leído"

# 11. puntaje que no es un numero: 400
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leido
rating: abc
HTTP 400
[Asserts]
xpath "normalize-space(//main/article/strong)" == "el puntaje tiene que ser un número del 1 al 5"

# 12. marcar como leido con puntaje y notas
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leido
rating: 4
notes: Muy bueno
HTTP 303
[Asserts]
header "Location" == "/"

# 13. leido: fecha de hoy como dd/mm/aaaa, puntaje, notas y la opcion de relectura
GET {{base_url}}/
HTTP 200
[Asserts]
xpath "string(//tr[td[1]='Rayuela']/td[3])" == "Leído"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[4])" matches "^\\d{2}/\\d{2}/\\d{4}$"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[5])" == "4"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[6])" == "Muy bueno"
xpath "count(//tr[td[1]='Rayuela']//summary[.='Lo volví a leer'])" == 1


# 14. relectura: un libro leido se vuelve a marcar como leido, con puntaje y notas nuevos
POST {{base_url}}/books/{{book_id}}/status
[FormParams]
status: leido
rating: 2
notes: Segunda lectura
HTTP 303
[Asserts]
header "Location" == "/"

# 15. la fila muestra el puntaje y las notas de la relectura
GET {{base_url}}/
HTTP 200
[Asserts]
xpath "string(//tr[td[1]='Rayuela']/td[3])" == "Leído"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[5])" == "2"
xpath "normalize-space(//tr[td[1]='Rayuela']/td[6])" == "Segunda lectura"


# 16. borrar
POST {{base_url}}/books/{{book_id}}/delete
HTTP 303
[Asserts]
header "Location" == "/"

# 17. borrar de nuevo, como con un doble clic: 404 con el mensaje y la lista vacia
POST {{base_url}}/books/{{book_id}}/delete
HTTP 404
[Asserts]
xpath "normalize-space(//main/article/strong)" == "libro no encontrado"
xpath "normalize-space(//main/section[2]/p)" == "Todavía no cargaste ningún libro."


# 18. alta con un titulo que es HTML
POST {{base_url}}/books
[FormParams]
title: <script>alert(1)</script>
author: Prueba de escape
HTTP 303

# 19. templ lo escapa: se ve como texto y la pagina no tiene ningun <script>
GET {{base_url}}/
HTTP 200
[Captures]
html_id: xpath "string(//tr[td[2]='Prueba de escape']//form[button='Borrar']/@action)" regex "/books/(\\d+)/delete"
[Asserts]
xpath "string(//tr[td[2]='Prueba de escape']/td[1])" == "<script>alert(1)</script>"
xpath "count(//script)" == 0
body contains "&lt;script&gt;alert(1)&lt;/script&gt;"

# 20. se borra para dejar la base vacia
POST {{base_url}}/books/{{html_id}}/delete
HTTP 303


# 21. id que no es un numero: 400
POST {{base_url}}/books/abc/status
HTTP 400
[Asserts]
xpath "normalize-space(//main/article/strong)" == "el id del libro no es válido"

# 22. ruta que no existe: 404
GET {{base_url}}/no-existe
HTTP 404

# 23. metodo no permitido en la pagina: 405 con la cabecera Allow
DELETE {{base_url}}/
HTTP 405
[Asserts]
header "Allow" == "GET"

# 24. GET a la ruta del alta: 405, solo acepta POST
GET {{base_url}}/books
HTTP 405
[Asserts]
header "Allow" == "POST"
