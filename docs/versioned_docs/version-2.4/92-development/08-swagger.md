# Swagger, API Spec and Code Generation

Woodpecker uses [gin-swagger](https://github.com/swaggo/gin-swagger) middleware to automatically
generate Swagger v2 API specifications and a nice looking Web UI from the source code.
Also, the generated spec will be transformed into Markdown, using [go-swagger](https://github.com/go-swagger/go-swagger)
and then being using on the community's website documentation.

It's paramount important to keep the gin handler function's godoc documentation up-to-date,
to always have accurate API documentation.
Whenever you change, add or enhance an API endpoint, please update the godocs.

You don't require any extra tools on your machine, all Swagger tooling is automatically fetched by standard Go tools.

## Gin-Handler API documentation guideline

Here's a typical example of how annotations for Swagger documentation look like...

```go title="server/api/user.go"
// @Summary  Get a user
// @Description Returns a user with the specified login name. Requires admin rights.
// @Router   /users/{login} [get]
// @Produce  json
// @Success  200 {object} User
// @Tags   Users
// @Param   Authorization header string true "Insert your personal access token" default(Bearer <personal access token>)
// @Param   login   path string true "the user's login name"
// @Param   foobar  query   string false "optional foobar parameter"
// @Param   page    query int  false "for response pagination, page offset number" default(1)
// @Param   perPage query int  false "for response pagination, max items per page" default(50)
```

```go title="server/model/user.go"
type User struct {
  ID int64 `json:"id" xorm:"pk autoincr 'user_id'"`
// ...
} // @name User
```

These guidelines aim to have consistent wording in the swagger doc:

- first word after `@Summary` and `@Summary` are always uppercase
- `@Summary` has no `.` (dot) at the end of the line
- model structs shall use custom short names, to ease life for API consumers, using `@name`
- `@Success` object or array declarations shall be short, this means the actual `model.User` struct must have a `@name` annotation, so that the model can be renderend in Swagger
- when pagination is used, `@Parame page` and `@Parame perPage` must be added manually
- `@Param Authorization` is almost always present, there are just a few un-protected endpoints

There are many examples in the `server/api` package, which you can use a blueprint.
More enhanced information you can find here <https://github.com/swaggo/swag/blob/master/README.md#declarative-comments-format>

### Manual code generation

```bash title="generate the server's Go code containing the Swagger"
make generate-swagger
```

```bash title="update the Markdown in the ./docs folder"
make docs
```

```bash title="auto-format swagger related godoc"
go run github.com/swaggo/swag/cmd/swag@latest fmt -g server/api/z.go
```
