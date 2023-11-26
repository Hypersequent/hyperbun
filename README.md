# Hyperbun

Hyperbun is a utility library built on top of [Bun](https://github.com/uptrace/bun) to reduce boilerplate code.

Queries like this in bun:

```go
func Foo() {
  // ...
	
  var user User
  if err := db.NewSelect().Model(&user).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      // ...
    }
	// ...
  }
		
  db.NewDelete().Model(&user).Where("id = ?", id).Exec(c.Request.Context())
}
```

can be written like this with Hyperbun:

```go
func Foo() {
  hdb := hyperbun.NewContext(c.Request.Context(), db)

  user, err := hyperbun.ByID(hdb, "users", id)
  if err != nil {
	// ...	
  }	
  if user == nil {
    // ...	
  }
	
  hyperbun.DeleteByID(hdb, "users", id)
}
```

It also makes working with transactions easier and tx and hdb can be used interchangeably. This promotes code reuse.

```go
func Foo(db *hyperbun.DB) {
  hyperbun.DeleteByID(db, "users", id)
}

func main() {
  hyperbun.RunInTx(hdb, func(tx *hyperbun.Tx) error {
    Foo(tx)
    return nil
  })

  Foo(hdb)
}
```

## License

- Distributed under MIT License, please see license file within the code for more details.

## Credits

- [Bun](https://github.com/uptrace/bun)
- 
