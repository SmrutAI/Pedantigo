---
sidebar_position: 4
---

# Collection Constraints

Validation rules for slices, arrays, and maps in Pedantigo. Collection constraints allow you to validate the size, uniqueness, and contents of collections.

## Key Distinction: Size vs. Element Validation

When using `min` and `max` on collections, they validate **item count**, not individual element properties:

- `min=3` on `[]string` means at least 3 items in the slice
- `min=3` on `string` means at least 3 characters in the string

This distinction is crucial for proper collection validation.

## Size Constraints

### `min` / `max` on Collections

Validates the **number of items** (elements) in a slice, array, or map.

```go
type ShoppingCart struct {
    // Must have at least 1 item, at most 100 items
    Items []Product `json:"items" pedantigo:"required,min=1,max=100"`

    // Tags collection: 0 to 10 tags
    Tags []string `json:"tags" pedantigo:"max=10"`
}

type ConfigMap struct {
    // Must have between 2 and 50 configuration entries
    Settings map[string]string `json:"settings" pedantigo:"min=2,max=50"`
}
```

**Behavior:**
- Counts the number of elements in the collection, not the size of individual elements
- Empty collections fail `min` validation if `min > 0`
- Works with slices, arrays, and maps
- Does not validate individual element constraints (use `dive` for that)

### `len`

Validates that a collection has an **exact number of items**.

```go
type ChessTournament struct {
    // Exactly 8 players
    Players []Player `json:"players" pedantigo:"len=8"`
}

type RGBColor struct {
    // Array must have exactly 3 values (red, green, blue)
    Values [3]int `json:"values" pedantigo:"len=3"`
}
```

**Behavior:**
- Validates exact item count, not individual element sizes
- Applies to slices, arrays, and maps
- Fails if the collection has any other number of items

## Uniqueness Constraints

### `unique`

Validates that all elements in a collection are **unique** (no duplicates).

```go
type UserGroup struct {
    // All email addresses must be different
    EmailAddresses []string `json:"emails" pedantigo:"required,unique"`

    // All product IDs must be unique
    ProductIDs []int `json:"product_ids" pedantigo:"required,unique"`
}
```

**Behavior:**
- Compares elements using equality (==)
- Works with strings, numbers, booleans, and any comparable type
- Order doesn't matter; `["a", "b", "a"]` is invalid
- Empty slices pass validation (no items to duplicate)
- For struct slices, see `unique=fieldName` below

**Valid examples:**
```
["apple", "banana", "cherry"] - all unique
[1, 2, 3, 4, 5] - all unique
```

**Invalid examples:**
```
["apple", "banana", "apple"] - "apple" appears twice
[1, 2, 3, 2, 4] - "2" appears twice
```

### `unique=fieldName` for Struct Slices

When validating slices of structs, use `unique=fieldName` to ensure all items have unique values for a specific field.

```go
type UserList struct {
    // All users must have different usernames
    Users []User `json:"users" pedantigo:"required,unique=Username"`
}

type User struct {
    Username string `json:"username" pedantigo:"required"`
    Email    string `json:"email" pedantigo:"required,email"`
}

type InventoryItem struct {
    // All items must have different SKU (stock keeping unit)
    Items []Product `json:"items" pedantigo:"required,unique=SKU"`
}

type Product struct {
    SKU  string `json:"sku"`
    Name string `json:"name"`
}
```

**Behavior:**
- Extracts the specified field from each struct element
- Compares the extracted values for uniqueness
- Field must exist and be comparable
- Fails if any two structs have the same field value
- Field name is case-sensitive

**Valid example:**
```
Users: [
    {Username: "alice", Email: "alice@example.com"},
    {Username: "bob", Email: "bob@example.com"},
    {Username: "charlie", Email: "charlie@example.com"}
]
```

**Invalid example:**
```
Users: [
    {Username: "alice", Email: "alice@example.com"},
    {Username: "bob", Email: "bob@example.com"},
    {Username: "alice", Email: "alice2@example.com"}  // Duplicate username
]
```

## Element Validation with `dive`

### `dive` - Validate Each Element

The `dive` constraint tells Pedantigo to apply constraints to each element in a collection, rather than the collection itself.

```go
type PostRequest struct {
    // Each tag must be lowercase, 3-20 characters
    Tags []string `json:"tags" pedantigo:"required,dive,lowercase,min=3,max=20"`

    // Each score must be between 0 and 100
    Scores []int `json:"scores" pedantigo:"required,dive,min=0,max=100"`
}
```

**Behavior:**
- Constraints after `dive` apply to each element, not the collection
- Collection constraints (`min`, `max`, `unique`) come before `dive`
- Element constraints come after `dive`
- Each element is validated independently
- All errors are collected and reported

### Combining Collection and Element Constraints

You can validate both the collection size and individual element properties:

```go
type DocumentUpload struct {
    // Must have 1-10 files, each with 3-50 character filename
    Files []string `json:"files" pedantigo:"required,min=1,max=10,dive,min=3,max=50"`

    // Must have 2-5 categories, each category must be lowercase letters only
    Categories []string `json:"categories" pedantigo:"required,min=2,max=5,dive,alpha,lowercase"`
}
```

**Behavior:**
- `min=1,max=10` validates the collection has 1-10 items
- `dive` marks the start of per-element constraints
- `min=3,max=50` validates each filename is 3-50 characters
- All collection constraints must come before `dive`
- All element constraints must come after `dive`

### Nested Struct Validation with `dive`

Use `dive` to validate each struct in a collection:

```go
type OrderRequest struct {
    // Each order item must be valid
    Items []OrderItem `json:"items" pedantigo:"required,min=1,max=100,dive"`
}

type OrderItem struct {
    SKU      string `json:"sku" pedantigo:"required,len=10"`
    Quantity int    `json:"quantity" pedantigo:"required,min=1,max=1000"`
    Price    float64 `json:"price" pedantigo:"required,gt=0"`
}
```

**Behavior:**
- `dive` applies struct validation to each OrderItem
- Each OrderItem's fields are validated according to its own constraints
- Errors include the index of the invalid element: `Items[0].SKU`
- Nested validation works recursively (struct within struct within collection)

**Example error output:**
```
Items[0].SKU: length must be exactly 10 characters
Items[2].Price: must be greater than 0
Items[5].Quantity: must be between 1 and 1000
```

## Map Key Validation

### `keys` / `endkeys` - Validate Map Keys

Use `keys` and `endkeys` to apply constraints specifically to map keys (not values).

```go
type AppConfig struct {
    // All keys must be lowercase alphanumeric
    Settings map[string]string `json:"settings" pedantigo:"required,keys,lowercase,alphanum,endkeys"`

    // All keys must be valid email addresses
    UserPreferences map[string]any `json:"preferences" pedantigo:"keys,email,endkeys"`
}
```

**Behavior:**
- `keys` marks the start of key-specific constraints
- `endkeys` marks the end of key constraints
- Constraints between `keys` and `endkeys` apply to map keys only
- Values are not validated (add separate constraints after `endkeys`)
- Works only with string keys (Go maps with string key types)

### Validating Both Keys and Values

```go
type EnvironmentVars struct {
    // Keys must be uppercase, values must be non-empty strings
    Variables map[string]string `json:"variables" pedantigo:"required,keys,uppercase,alphanum,endkeys,dive,min=1"`
}
```

**Behavior:**
- `keys` ... `endkeys` validates map keys
- Constraints after `endkeys` validate values
- Use `dive` before value constraints to apply per-value validation

## Default Values

### `default` on Collections

Provides a default value when a collection field is missing from JSON.

```go
type UserPreferences struct {
    // If tags are not provided, default to empty array
    Tags []string `json:"tags,omitempty" pedantigo:"default="`

    // If not provided, defaults to a predefined list
    Regions []string `json:"regions" pedantigo:"default=us-east-1,us-west-2"`
}
```

**Behavior:**
- For slices/arrays: `default=` creates an empty collection
- For slices with values: `default=value1,value2,value3` creates a collection with default items
- Only applied when the field is missing from JSON (not when empty)
- Works with `omitempty` in JSON tags

## Complete Example

Here's a comprehensive example showing multiple collection constraints working together:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type BlogPost struct {
    Title string `json:"title" pedantigo:"required,min=5,max=200"`

    // 1-1000 tags, each 2-30 lowercase characters
    Tags []string `json:"tags" pedantigo:"min=1,max=1000,dive,min=2,max=30,lowercase"`

    // 1-10 authors, all with different emails
    Authors []Author `json:"authors" pedantigo:"required,min=1,max=10,dive,unique=Email"`

    // Map of translations: keys must be language codes (2 chars uppercase)
    // Values must be 10+ characters
    Translations map[string]string `json:"translations" pedantigo:"keys,len=2,uppercase,alphanum,endkeys,dive,min=10"`

    // Each comment must be valid
    Comments []Comment `json:"comments" pedantigo:"max=500,dive"`

    // Related post URLs: all must be unique, each must be valid URL
    RelatedPosts []string `json:"related_posts" pedantigo:"max=5,dive,unique,url"`
}

type Author struct {
    Name  string `json:"name" pedantigo:"required,min=2,max=100"`
    Email string `json:"email" pedantigo:"required,email"`
}

type Comment struct {
    Author  string `json:"author" pedantigo:"required,min=1,max=50"`
    Text    string `json:"text" pedantigo:"required,min=10,max=5000"`
    Rating  int    `json:"rating" pedantigo:"required,min=1,max=5"`
}

func main() {
    // Valid blog post
    postJSON := []byte(`{
        "title": "Getting Started with Pedantigo",
        "tags": ["validation", "golang", "pedantigo"],
        "authors": [
            {
                "name": "Alice Smith",
                "email": "alice@example.com"
            },
            {
                "name": "Bob Jones",
                "email": "bob@example.com"
            }
        ],
        "translations": {
            "en": "This is a detailed blog post about validation",
            "es": "Este es un artículo detallado sobre validación",
            "fr": "Ceci est un article détaillé sur la validation"
        },
        "comments": [
            {
                "author": "Reader One",
                "text": "Great article, very helpful!",
                "rating": 5
            }
        ],
        "related_posts": [
            "https://example.com/post1",
            "https://example.com/post2"
        ]
    }`)

    post, err := pedantigo.Unmarshal[BlogPost](postJSON)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    fmt.Printf("Valid post: %s by %d authors\n", post.Title, len(post.Authors))

    // Invalid blog post - multiple constraint violations
    invalidJSON := []byte(`{
        "title": "Hi",
        "tags": ["a", "validation", "validation"],
        "authors": [
            {
                "name": "Alice",
                "email": "alice@example.com"
            },
            {
                "name": "Bob",
                "email": "alice@example.com"
            }
        ],
        "translations": {
            "english": "Too long key",
            "es": "short"
        },
        "comments": [
            {
                "author": "Reader",
                "text": "Short",
                "rating": 10
            }
        ],
        "related_posts": [
            "not-a-url",
            "https://example.com/post1"
        ]
    }`)

    _, err = pedantigo.Unmarshal[BlogPost](invalidJSON)
    if err != nil {
        fmt.Printf("Validation errors:\n%v\n", err)
        // Output will show:
        // - title: length must be between 5 and 200 characters
        // - tags[0]: length must be between 2 and 30 characters
        // - tags[2]: duplicate value "validation"
        // - authors[1].email: duplicate value "alice@example.com"
        // - translations: key "english" must be exactly 2 characters
        // - translations["es"]: must be between 10 and 5000 characters
        // - comments[0].rating: must be between 1 and 5
        // - related_posts[0]: must be a valid URL
    }
}
```

## Validation Behavior Notes

- **Empty collections:** Collections validate as empty by default. Use `min=1` to require at least one item.
- **Order independence:** `unique` doesn't care about order; duplicates are detected regardless of position.
- **Performance:** `dive` is efficient; validation is performed in a single pass with element index tracking.
- **Nested errors:** When using `dive`, error messages include the element index, e.g., `Items[2].Price`.
- **Nil/zero slices:** A nil slice vs an empty slice `[]` both have length 0 for size validation purposes.
- **Map considerations:** Maps have no guaranteed order; uniqueness is based on map keys, not iteration order.

## Quick Reference Table

| Constraint | Example | Applies To | Description |
|------------|---------|-----------|-------------|
| `min=N` | `min=1` | Collections | Minimum number of items |
| `max=N` | `max=100` | Collections | Maximum number of items |
| `len=N` | `len=8` | Collections | Exact number of items |
| `unique` | `unique` | Collections | All items must be unique |
| `unique=field` | `unique=Email` | Struct slices | All structs must have unique field value |
| `dive` | `dive` | Collections | Validate each element (constraints follow) |
| `keys` | `keys` | Maps | Start of key constraints |
| `endkeys` | `endkeys` | Maps | End of key constraints |
| `default=val` | `default=` | Collections | Default value when missing |

## Advanced Patterns

### Validating Complex Nested Structures

```go
type Project struct {
    // Tasks with assigned users, validated deeply
    Tasks []Task `json:"tasks" pedantigo:"required,min=1,max=500,dive"`
}

type Task struct {
    Title       string   `json:"title" pedantigo:"required,min=3,max=100"`
    AssignedTo  []User   `json:"assigned_to" pedantigo:"max=10,dive"`
    SubTasks    []Task   `json:"subtasks" pedantigo:"max=50,dive"`  // Recursive validation
}

type User struct {
    Username string `json:"username" pedantigo:"required,alphanum,min=3,max=20"`
}
```

### Set Semantics with Unique Collections

```go
type DataSet struct {
    // Unique identifiers - enforces set semantics
    IDs []string `json:"ids" pedantigo:"required,unique,dive,uuid"`

    // Unique enum values
    Statuses []string `json:"statuses" pedantigo:"required,unique,dive,oneof=active inactive pending"`
}
```

### Size-Constrained Collections with Format Validation

```go
type NotificationConfig struct {
    // 1-5 unique email addresses for notifications
    EmailRecipients []string `json:"recipients" pedantigo:"required,min=1,max=5,unique,dive,email"`

    // 1-10 webhook URLs, all unique and valid
    WebhookURLs []string `json:"webhooks" pedantigo:"min=1,max=10,unique,dive,url,startswith=https://"`
}
```
