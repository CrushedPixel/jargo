---
layout: default
---

# [Introduction](#introduction)
Resource models are defined by creating `struct` types, configured via `jargo` [struct tags][struct-tags].

# [Id field](#id-field)
Every model definition has an **id field**, which serves as the **primary key**.
 
The id field **must** be named `Id`.
Currently, it has to be of type `int64`, however support for `UUID` and `string` values is planned.

~~~go
Id int64
~~~

## [Resource name](#resource-name)
By default, a resource's **JSON API name** is the **underscored** and **pluralized** version of the struct name.
~~~go
type CookieFlavor struct {
    Id int64 // resource name: cookie_flavors
}
~~~  
To override the resource name, set the first value in the *id field's* `jargo` struct tag:
~~~go
type CookieFlavor struct {
    Id int64 `jargo:"flavors"`
}
~~~

*All resource names must adhere to the [JSON API specification][member-names].*

To keep the default resource name, but specify other options, 
simply omit the name and start the struct tag with a comma:
~~~go
Id int64 `jargo:",table:my_table"`
~~~

## [Table name](#table-name)
By default, a resource's **database table name** is the same as its [resource name](#resource-name).

To override the table name, use the `table` option on the *id field*:
~~~go
type CookieFlavor struct {
    Id int64 `jargo:",table:gusto"`
}
~~~

## [Table alias](#table-alias)
By default, a resource's **table alias** to use in *SQL queries* is the **underscored** version of the struct name.
~~~go
type CookieFlavor struct {
    Id int64 // table alias: cookie_flavor
}
~~~
To override the table alias, use the `alias` option on the *id field*:
~~~go
type CookieFlavor struct {
    Id int64 `jargo:",alias:flavor"`
}
~~~

# [Attribute fields](#attribute-fields)
A *resource model* may have any number of *attributes* representing primitive data types stored in the database.

Supported attribute types are:

| Go type                                    | PostgreSQL type    |
|--------------------------------------------|--------------------|
| `int8`, `uint8`, `int16`                   | `smallint`         |
| `uint16`, `int32`                          | `integer`          |
| `uint32`, `int64`, `int`, `uint`, `uint64` | `bigint`           |
| `float32`                                  | `real`             |
| `float64`                                  | `double precision` |
| `bool`                                     | `boolean`          |
| `string`                                   | `text`             |
| [`time.Time`][time.Time]                   | `timestamptz`      |

## [JSON API member name](#json-api-member-name)
By default, an attribute's **JSON API member name** is the **dasherized** version of the field name.  
~~~go
Age      int    // name: age
UserName string // name: user-name
FooBar   int    // name: foo-bar
~~~
To override the member name, set the first value in the `jargo` struct tag:
~~~go
Age      int    `jargo:"age_in_years"`
UserName string `jargo:"name"`
FooBar   int    `jargo:"foo_bar"`
~~~

*All member names must adhere to the [JSON API specification][member-names].*

To keep the default member name, but specify other options, 
simply omit the name and start the struct tag with a comma:
~~~go
FooBar int `jargo:",unique"`
~~~

### [Unexported attributes](#unexported-attributes)
Sometimes, you don't want to expose certain attributes via *JSON API*, but still keep them in your database.  
To exclude an attribute from the generated API, simply set its member name to `-`:
~~~go
PasswordHash string `jargo:"-"`
~~~ 

## [Column name](#column-name)
By default, an attribute's **database column name** is the **underscored** version of the field name.
~~~go
Age      int    // column: age
UserName string // column: user_name
FooBar   int    // column: foo_bar
~~~
To override the column name, use the `column` option:
~~~go
Age      int    `jargo:",column:attr_age"`
UserName string `jargo:",column:name"`
FooBar   int    `jargo:",column:FooBarColumn"`
~~~

## [Nullable attributes](#nullable-attributes)
*Primitive* attributes have a `NOT NULL` constraint in the database.
If an attribute should be **nullable**, use a **pointer type** instead:

~~~go
Age  int  // field can't be NULL
Size *int // nil pointers are treated as NULL
~~~

## [Default values](#default-values)
You may define a `DEFAULT` constraint for pointer attributes.
When inserting a resource instance with the attribute set to `nil`,
the `DEFAULT` constraint will take effect:

~~~go
type Person struct {
    Id   int64
    Name *string `jargo:",default:'John Doe'"`
}

// insert person without name
res, err := req.Resource().InsertInstance(req.DB(), &Person{}).Result()
if err != nil {
    return jargo.NewErrorResponse(err)
}
person := res.(*Person)
log.Printf("Name: %s", person.Name) // person's name is "John Doe"
~~~

*The `default` option must be valid SQL*, as it is the constraint itself.
This allows you to use SQL functions, such as `NOW()` in `DEFAULT` constraints:

~~~go
Expires *time.Time `jargo:",default:NOW() + INTERVAL '1 day'"`
~~~

### [`NOT NULL` attributes with default values](#not-null-attributes-with-default-values)
You may add the `notnull` option to an attribute with a default value 
to add a `NOT NULL` constraint to the database column:
~~~go
Size *int `jargo:",notnull,default:170"`
~~~

Setting the value to `nil` omits it when updating it in the database,
to ensure the `NOT NULL` constraint is never violated.

## [Unique attributes](#unique-attributes)
The `unique` option adds a `UNIQUE` constraint to the database column.
~~~go
Name  string `jargo:",unique"`
Email string `jargo:",unique"`
~~~
A [Query][queries] violating this constraint returns an [`UniqueViolationError`][UniqueViolationError]:
~~~go
type User struct {
    Id   int64
    Name string `jargo:",unique"`
}

// insert user with name "Marius"
res, err := req.Resource().InsertInstance(req.DB(), &User{Name: "Marius"}).Result()
if err != nil {
    return jargo.NewErrorResponse(err)
}

// insert another user with same name, violating the unique constraint
res, err = req.Resource().InsertInstance(req.DB(), &User{Name: "Marius"}).Result()
if err != nil {
    if uve, ok := err.(*jargo.UniqueViolationError); ok {
        log.Printf("Unique constraint violation on field %s\n", uve.Field)
    }
    return jargo.NewErrorResponse(err)
}
~~~

## [Readonly attributes](#readonly-attributes)
By default, all of a model's attributes can be set in **create** or **update requests**.  
To disallow setting the value of a specific attribute, add the `readonly` option:
~~~go
Money float64 `jargo:",readonly"`
~~~

## [Omitting empty attributes](#omitting-empty-attributes)
If you want to omit attributes with their type's [zero value][zero-value] from *JSON API* payloads,
you can add the `omitempty` option:
~~~go
Name string `jargo:",omitempty"` // omitted if value is ""
Age  int    `jargo:",omitempty"` // omitted if value is 0
Pets *int   `jargo:",omitempty"` // omitted if value is <nil>
~~~

## [Sorting](#sorting)
By default, [sorting][sorting] is enabled for all **non-nullable** attributes and `belongsTo` relations.  
To disable sorting by an attribute, add the `nosort` option:
~~~go
Age int `jargo:",nosort"`
~~~

## [Filtering](#filtering)
By default, [filtering][filtering] is enabled for all attributes and `belongsTo` relations.  
To disable filtering by an attribute, add the `nofilter` option:
~~~go
Age int `jargo:",nofilter"`
~~~

## [Automatic Timestamps](#automatic-timestamps)
Often times, you want to keep track of when a model has been **created** in the database 
and when it was **last updated**.  
This can be achieved using the `createdAt` and `updatedAt` options on a [`*time.Time`][time.Time] attribute:
~~~go
type User struct {
    Id          int64
    Name        string
    Joined      *time.Time `jargo:",createdAt"`
    LastUpdated *time.Time `jargo:",updatedAt"`
}

// insert new user instance
res, err := req.Resource().InsertInstance(req.DB(), &User{Name: "Marius"}).Result()
if err != nil {
    return jargo.NewErrorResponse(err)
}
user := res.(*User)
log.Printf("Joined: %s", user.Joined) // user.Joined is the time of creation in the database

// update user instance
user.Name = "Jane"
res, err = req.Resource().UpdateInstance(req.DB(), user).Result()
if err != nil {
    return jargo.NewErrorResponse(err)
}
user = res.(*User)
log.Printf("Joined: %s", user.LastUpdated) // user.LastUpdated is the time of last update in the database
~~~

# [Relations](#relations)
*TODO: write this section*

# [Options cheat sheet](#options-cheat-sheet)
Here's a list of all options in `jargo` struct tags:

| Option                                                | Field types                    | Description                                                            |
|-------------------------------------------------------|--------------------------------|------------------------------------------------------------------------|
| [`table`](#table-name)                                | [Id](#id-field)                | Sets the resource's **table name**.                                    |
| [`alias`](#table-alias)                               | [Id](#id-field)                | Sets the resource's **table alias**.                                   |
| [`column`](#column-name)                              | [Attribute](#attribute-fields) | Sets the attribute's **column name**.                                  |
| [`default`](#default-values)                          | [Attribute](#attribute-fields) | Sets the attribute's **default value**.                                |
| [`notnull`](#not-null-attributes-with-default-values) | [Attribute](#attribute-fields) | Adds a **`NOT NULL` constraint** to a nullable type.                   |
| [`unique`](#unique-attributes)                        | [Attribute](#attribute-fields) | Adds a **`UNIQUE` constraint** to an attribute.                        |
| [`readonly`](#readonly-attributes)                    | [Attribute](#attribute-fields) | Disallows **setting** the attribute via JSON API.                      |
| [`omitempty`](#omitting-empty-attributes)             | [Attribute](#attribute-fields) | **Omits** the attribute from JSON API if empty.                        |
| [`nosort`](#sorting)                                  | [Attribute](#attribute-fields) | **Disables sorting** by an attribute.                                  |
| [`nofilter`](#filtering)                              | [Attribute](#attribute-fields) | **Disables filtering** by an attribute.                                |
| [`createdAt`](#automatic-timestamps)                  | [Attribute](#attribute-fields) | Automatically sets the attribute's value to the **time of insertion**. |
| [`updatedAt`](#automatic-timestamps)                  | [Attribute](#attribute-fields) | Automatically sets the attribute's value to the **time of updating**.  |

[struct-tags]: https://golang.org/ref/spec#Tag
[time.Time]: https://golang.org/pkg/time/#Time
[member-names]: http://jsonapi.org/format/#document-member-names
[queries]: #TODO
[UniqueViolationError]: https://godoc.org/github.com/CrushedPixel/jargo#UniqueViolationError
[zero-value]: https://golang.org/ref/spec#The_zero_value
[sorting]: http://jsonapi.org/format/#fetching-sorting
[filtering]: http://jsonapi.org/format/#fetching-filtering