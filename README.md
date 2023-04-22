# KB - a tiny Kanban Board API
`kb` is a simple WIP api for kanban boards

## Usage
```
go run ./src
```
now it should be running on [localhost:8080](http://localhost:8080)

There are these endpoints:
  - `POST /` - create a new board, the body should be of this form
    ```json
    {
      "title": "My great title"
    }
    ```
  - `GET /` - return all the boards
  - `GET /:id` - return board with `:id`
  - `POST /:id/card` - create a new card, the body should be of this form
    ```json
    {
      "content": "This could be anything you like, **markdown** ?"
    }
    ```
  - `GET /:id/card` - get the cards associated with the board with `:id`
  - `DELETE /:id` - delete a board (this deletes cards attached)
