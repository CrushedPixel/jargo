---
layout: default
---

# Introduction
Models are objects that represent the resources your API exposes.

# Id field

# Attribute fields
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

## [Nullable attributes](#nullable-attributes)
*Primitive* attributes have a `NOT NULL` constraint in the database.
If an attribute should be **nullable**, use a **pointer type** instead:

```go
Age  int  // field can't be NULL
Size *int // nil pointers are treated as NULL
```

## [Default values](#default-values)
You may define a `DEFAULT` constraint for pointer attributes.
When inserting a resource instance with an attribute set to `nil`,
the `DEFAULT` constraint will take effect:

```go
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
```

The `default` option must be valid SQL, as it is the constraint itself.
This allows you to use SQL functions, such as `NOW()` in `DEFAULT` constraints:

```go
Expires *time.Time `jargo:",default:NOW() + INTERVAL '1 day'"`
```

### [`NOT NULL` attributes with default values](#not-null-attributes-with-default-values)

You may add the `notnull` option to an attribute with a default value 
to add a `NOT NULL` constraint to the database column:
```go
Size *int `jargo:",notnull,default:170"`
```

Setting the value to `nil` omits it when updating it in the database,
to ensure the `NOT NULL` constraint is never violated.

## Attribute name

## Column name

## Sorting
By default, sorting is enabled for all **non-nullable** attributes and `belongsTo` relations.

[time.Time]: https://golang.org/pkg/time/#Time