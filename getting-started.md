---
layout: default
---

# Getting started
This article shows how to create a fully featured [**JSON API**][jsonapi] using **jargo** within a couple of minutes.

# Table of contents

- [Installing jargo](#installing-jargo)
- [Connecting to PostgreSQL](#connecting-to-postgresql)
- [Creating an Application](#creating-an-application)


## [Installing jargo](#installing-jargo)
**jargo** is installed using `go get`, like any other dependency:
```
go get github.com/crushedpixel/jargo
```

## [Connecting to PostgreSQL](#connecting-to-postgresql)
**jargo** uses [`go-pg`][go-pg] to connect to a [**PostgreSQL**][PostgreSQL] database to store and retrieve resource data.
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

The main component of your **jargo** project is a [`jargo.Application`][jargo.Application] instance.  
It is responsible for generating all [**JSON API**][jsonapi] endpoints, which can be served via *HTTP* or *Websockets*.

An Application is created using `jargo.NewApplication`, which takes a [`jargo.Options`][jargo.Options] parameter,
used configure your Application.  
The only required field is `DB`, which is the database connection we created in the previous step.

```go
app := jargo.NewApplication(jargo.Options{
	DB: db,
})
```

## [Defining Resource Models](#defining-resource-models)

In order to expose any resources via [**JSON API**][jsonapi], you have to define them.  
To define a **jargo** resource model, define a `struct` type and annotate it with `jargo` [struct tags][struct-tags]. 

Let's define a simple `User` model with the following attributes:

| JSON API Name | Type                     | Description                                                                                               |
| ------------- | ------------------------ | --------------------------------------------------------------------------------------------------------- |
| `id`          | `int64`                  | The model's **unique** primary key. The `id` attribute is always required, and has to be of type `int64`. |
| `name`        | `string`                 | The username. Must be **unique** among all users.                                                         |
| `age`         | `int`                    | The user's age.                                                                                           |
| `joined-at`   | [`time.Time`][time.Time] | The date and time at which the user registered. Should be **automatically set** by the database.          |

This is what your `User` struct could look like:
```go
type User struct {
	Id       int64
	Name     string    `jargo:",unique"`
	Age      int
	JoinedAt time.Time `jargo:",createdAt"`
}
```

> For a more in-depth guide on Resource Model definition, including relationships, visit the dedicated page I'm going to write soon™.

## [Creating Resource Controllers](#creating-resource-controllers)

*TODO*

## [Serving the Application](#serving-the-application)

*TODO*

[jsonapi]: http://jsonapi.org/
[PostgreSQL]: https://postgresql.org
[go-pg]: https://github.com/go-pg/pg
[pg.DB]: https://godoc.org/github.com/go-pg/pg#DB
[postgres-installation]: https://wiki.postgresql.org/wiki/Detailed_installation_guides
[jargo.Application]: https://godoc.org/github.com/CrushedPixel/jargo#Application
[jargo.Options]: https://godoc.org/github.com/CrushedPixel/jargo#Options
[struct-tags]: https://golang.org/ref/spec#Tag
[time.Time]: https://golang.org/pkg/time/#Time