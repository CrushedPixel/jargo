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
~~~
go get github.com/crushedpixel/jargo
~~~

## [Connecting to PostgreSQL](#connecting-to-postgresql)
*jargo* uses [*go-pg*][go-pg] to connect to a [*PostgreSQL*][PostgreSQL] database to store and retrieve resource data.
Create a database connection using `pg.Connect`:

~~~go
db := pg.Connect(&pg.Options{
    // replace with your connection data
    Addr:     "localhost:5432",
    User:     "jargo",
    Database: "jargo",
    Password: "my-secret-password",
})
~~~

> If you need help setting up a PostgreSQL database, please refer to [the PostgreSQL wiki][postgres-installation].

## [Creating an Application](#creating-an-application)
The main component of your *jargo* project is a [`jargo.Application`][jargo.Application] instance.  
It is responsible for generating all [*JSON API*][jsonapi] endpoints.

An *Application* is created using `jargo.NewApplication`, which takes a [`jargo.Options`][jargo.Options] parameter,
used configure your *Application*.  
The only required field is `DB`, which is the database connection we created in the previous step.

~~~go
app := jargo.NewApplication(jargo.Options{
    DB: db,
})
~~~

## [Defining a Resource](#defining-a-resource)
In order to expose any resources via [*JSON API*][jsonapi], you have to create a Model definition first.
This is done by defining a `struct` type and annotating it with `jargo` [struct tags][struct-tags]. 

Let's say we want to create a `User` model with the following attributes:

| JSON API Name | Type                     | Description                                                                                               |
|---------------|--------------------------|-----------------------------------------------------------------------------------------------------------|
| `id`          | `int64`                  | The model's **unique** primary key. The `id` attribute is always required, and has to be of type `int64`. |
| `name`        | `string`                 | The username. Must be **unique** among all users.                                                         |
| `age`         | `int`                    | The user's age.                                                                                           |
| `joined-at`   | [`time.Time`][time.Time] | The date and time at which the user registered. Should be **automatically set** by the database.          |

This is what our `User` struct could look like:
~~~go
type User struct {
    Id       int64
    Name     string     `jargo:",unique"`
    Age      int
    JoinedAt *time.Time `jargo:",createdAt"`
}
~~~

> For a more in-depth guide on Resource Model definition, including relationships, 
  please refer to the dedicated page: [Model definition][model-definition]

Now that we defined the *Resource Model*, we can register it with our *Application*
to obtain a [`jargo.Resource`][jargo.Resource] instance:

~~~go
userResource := app.MustRegisterResource(User{})
~~~

Registering the *Resource* automatically creates a database table for the *Resource*.

## [Creating a Controller](#creating-a-controller)
Now that we have successfully registered a *Resource*, we can create a [`jargo.Controller`][jargo.Controller],
which handles all requests related to this *Resource*.

For this basic example, we use `Application.NewCRUDController` to create a *Controller*
that handles [*JSON API*][jsonapi] *Index*, *Show*, *Create*, *Update* and *Delete* requests:

~~~go
userController := app.NewCRUDController(userResource)
~~~

> For more information about Controllers and how to customize them, 
  please refer to the dedicated page I'm going to write soon™.

## [Serving the Application](#serving-the-application)
We can now serve our *Application* via *HTTP*. 
This is as straight-forward as providing a serve address and API namespace:

~~~go
log.Println("Serving Application...")
err := app.ServeHTTP("127.0.0.1:8080", "/api")
if err != nil {
    log.Fatalf("Error serving Application: %s\n", err.Error())
}
~~~

Installing and running your program should now host your [*JSON API*][jsonapi] on *localhost*, Port *8080*.

> For a more in-depth guide on serving an Application, including how to serve it via Websocket, 
  please refer to the dedicated page I'm going to write soon™.

## [Using your API](#using-your-api)
Let's make some requests to our API. 
Note that all requests are made to the `/api` namespace we set earlier.

### Listing resources
First of all, let's do an *Index Request* to the `user` resource:
~~~
curl http://127.0.0.1:8080/api/users -H "Content-Type: application/vnd.api+json"
~~~
~~~json
{
  "data": []
}                                   
~~~
Since we haven't created any users yet, the response contains an empty resource collection.

### Creating a resource
Let's do a *Create Request*, creating a user with name `Bob` and age `25`.

~~~
curl http://127.0.0.1:8080/api/users -H "Content-Type: application/vnd.api+json" -d '{"data":{"type":"users","attributes":{"name":"Bob","age":25}}}'
~~~
~~~json
{
  "data": {
    "type": "users",
    "id": "1",
    "attributes": {
      "age": 25,
      "joined-at": "2018-03-06T21:49:57Z",
      "name": "Bob"
    }
  }
}
~~~
The response contains the newly created user resource.

As you can see, the `joined-at` attribute was automatically set by the database 
due to the `createdAt` struct tag in the model definition.

### Showing a resource
We can now request the user with id `1` using a *Show Request*:
~~~
curl http://127.0.0.1:8080/api/users/1 -H "Content-Type: application/vnd.api+json"
~~~
~~~json
{
  "data": {
    "type": "users",
    "id": "1",
    "attributes": {
      "age": 25,
      "joined-at": "2018-03-06T21:49:57Z",
      "name": "Bob"
    }
  }
}
~~~

### Deleting a resource
Let's delete the user again by making a *Delete Request*:
~~~
curl http://127.0.0.1:8080/api/users/1 -H "Content-Type: application/vnd.api+json" -X DELETE
~~~
The response is empty, with status code `204 No Content`.

## [Conclusion](#conclusion)
In this article, you've created and used a simple [*JSON API*][jsonapi] using *jargo* - amazing!

You're now ready to move on to the more in-depth documentation explaining how you can
customize your *Application* and *Controllers* to make it suit your requirements.

Here's the full source code of the *Application* we just created:
~~~go
package main

import (
    "github.com/crushedpixel/jargo"
    "github.com/go-pg/pg"
    "log"
    "time"
)

type User struct {
    Id       int64
    Name     string     `jargo:",unique"`
    Age      int
    JoinedAt *time.Time `jargo:",createdAt"`
}

func main() {
    db := pg.Connect(&pg.Options{
        // replace with your connection data
        Addr:     "localhost:5432",
        User:     "test",
        Database: "jargo",
        Password: "my-secret-password",
    })
    
    // create application
    app := jargo.NewApplication(jargo.Options{
        DB: db,
    })
    
    // create user resource
    userResource := app.MustRegisterResource(User{})
    
    // create controller for user resource
    app.NewCRUDController(userResource)
    
    // serve application
    log.Println("Serving Application...")
    err := app.ServeHTTP("127.0.0.1:8080", "/api")
    if err != nil {
        log.Fatalf("Error serving Application: %s\n", err.Error())
    }
}
~~~

[jsonapi]: http://jsonapi.org/
[PostgreSQL]: https://postgresql.org
[go-pg]: https://github.com/go-pg/pg
[pg.DB]: https://godoc.org/github.com/go-pg/pg#DB
[postgres-installation]: https://wiki.postgresql.org/wiki/Detailed_installation_guides
[jargo.Application]: https://godoc.org/github.com/CrushedPixel/jargo#Application
[jargo.Options]: https://godoc.org/github.com/CrushedPixel/jargo#Options
[struct-tags]: https://golang.org/ref/spec#Tag
[time.Time]: https://golang.org/pkg/time/#Time
[model-definition]: model-definition
[jargo.Resource]: https://godoc.org/github.com/CrushedPixel/jargo#Resource
[jargo.Controller]: https://godoc.org/github.com/CrushedPixel/jargo#Controller
[ferry]: https://github.com/CrushedPixel/ferry
[http_bridge]: https://github.com/CrushedPixel/http_bridge