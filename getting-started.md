---
layout: default
---

# Getting started
This article shows how to create a fully featured [**JSON API**][jsonapi] using **jargo** within a couple of minutes.

# Table of contents

- [Installing jargo](#installing-jargo)
- [Connecting to PostgreSQL](#connecting-to-postgresql)
- [Creating an Application](#creating-an-application)
- [Defining a Resource](#defining-a-resource)
- [Creating a Controller](#creating-a-controller)
- [Serving the Application](#serving-the-application)
- [Using your API](#using-your-api)
- [Conclusion](#conclusion)

## [Installing jargo](#installing-jargo)
*jargo* is installed using `go get`, like any other dependency:
```
go get github.com/crushedpixel/jargo
```

## [Connecting to PostgreSQL](#connecting-to-postgresql)
*jargo* uses [`go-pg`][go-pg] to connect to a [*PostgreSQL*][PostgreSQL] database to store and retrieve resource data.
Create a database connection using `pg.Connect`:

```go
db := pg.Connect(&pg.Options{
    // replace with your connection data
    Addr:     "localhost:5432",
    User:     "jargo",
    Database: "jargo",
    Password: "my-secret-password",
})
```

> If you need help setting up a PostgreSQL database, please refer to [the PostgreSQL wiki][postgres-installation].

## [Creating an Application](#creating-an-application)
The main component of your *jargo* project is a [`jargo.Application`][jargo.Application] instance.  
It is responsible for generating all [*JSON API*][jsonapi] endpoints, which can be served via *HTTP* or *Websockets*.

An *Application* is created using `jargo.NewApplication`, which takes a [`jargo.Options`][jargo.Options] parameter,
used configure your *Application*.  
The only required field is `DB`, which is the database connection we created in the previous step.

```go
app := jargo.NewApplication(jargo.Options{
	DB: db,
})
```

## [Defining a Resource](#defining-a-resource)
In order to expose any resources via [*JSON API*][jsonapi], you have to create a Model definition first.
This is done by defining a `struct` type and annotating it with `jargo` [struct tags][struct-tags]. 

Let's say we want to create a `User` model with the following attributes:

| JSON API Name | Type                     | Description                                                                                               |
| ------------- | ------------------------ | --------------------------------------------------------------------------------------------------------- |
| `id`          | `int64`                  | The model's **unique** primary key. The `id` attribute is always required, and has to be of type `int64`. |
| `name`        | `string`                 | The username. Must be **unique** among all users.                                                         |
| `age`         | `int`                    | The user's age.                                                                                           |
| `joined-at`   | [`time.Time`][time.Time] | The date and time at which the user registered. Should be **automatically set** by the database.          |

This is what our `User` struct could look like:
```go
type User struct {
	Id       int64
	Name     string    `jargo:",unique"`
	Age      int
	JoinedAt time.Time `jargo:",createdAt"`
}
```

> For a more in-depth guide on Resource Model definition, including relationships, visit the dedicated page I'm going to write soon™.

Now that we defined the *Resource Model*, we can register it with our *Application*
to obtain a [`jargo.Resource`][jargo.Resource] instance:

```go
userResource := app.MustRegisterResource(User{})
```

## [Creating a Controller](#creating-a-controller)
Now that we have successfully registered a *Resource*, we can create a [`jargo.Controller`][jargo.Controller],
which handles all requests related to this *Resource*.

For this basic example, we use `Application.NewCRUDController` to create a *Controller*
that handles [*JSON API*][jsonapi] *Index*, *Show*, *Create*, *Update* and *Delete* requests:

```go
userController := app.NewCRUDController(userResource)
```

> For more information about Controllers and how to customize them, visit the dedicated page I'm going to write soon™.

## [Serving the Application](#serving-the-application)

*TODO: show how to serve application via HTTP*

## [Using your API](#using-your-api)

*TODO: give JSON API request examples using `curl`*

## [Conclusion](#conclusion)
In this article, you've created your very first *jargo Application* - amazing!

You're now ready to move on to the more in-depth documentation explaining how you can
customize your *Application* and *Controllers* to make it suit your requirements.

Here's the full source code of the *Application* we just wrote:
```go
// TODO
```

[jsonapi]: http://jsonapi.org/
[PostgreSQL]: https://postgresql.org
[go-pg]: https://github.com/go-pg/pg
[pg.DB]: https://godoc.org/github.com/go-pg/pg#DB
[postgres-installation]: https://wiki.postgresql.org/wiki/Detailed_installation_guides
[jargo.Application]: https://godoc.org/github.com/CrushedPixel/jargo#Application
[jargo.Options]: https://godoc.org/github.com/CrushedPixel/jargo#Options
[struct-tags]: https://golang.org/ref/spec#Tag
[time.Time]: https://golang.org/pkg/time/#Time
[jargo.Resource]: https://godoc.org/github.com/CrushedPixel/jargo#Resource
[jargo.Controller]: https://godoc.org/github.com/CrushedPixel/jargo#Controller