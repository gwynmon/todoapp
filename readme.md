Stack: sqlx RabbitMQ goose

https://github.com/evrone/go-clean-template/blob/master/README_RU.md

### Endpoints
#### Account Endpoints
- [POST] **/api/register** - accepts JSON {name, email, password}
- [POST] **/api/login** - accepts JSON {email, password} and returns Token
#### Tasks Endpoints
- [GET] **/api/tasks** - returns **cached** list of tasks
- [POST] **/api/tasks** - creates task
- [GET] **/api/tasks/{id}** - returns specified task
- [PUT] **/api/tasks/{id}** - updates specified task
- [DELETE] **/api/tasks/{id}** - deletes specified task
#### Notes Endpoints
- [POST] **/api/tasks/{id}/notes** - adds a note for specified task
- [GET] **/api/tasks/{id}/notes** - returns a note from specified task
- [DELETE] **/api/notes/{noteId}** - deletes specified note
