# Forward References in Go vs. Python

## Short Answer

**We don't need Pydantic's `.model_rebuild()` in Go** — the language has built-in solutions.

---

## Why Forward References Work Differently

### Python (Dynamic/Interpreted)

```python
# Annotations stored as strings, evaluated later
class Author(BaseModel):
    books: list['Book']  # String stored, evaluated on .model_rebuild()

class Book(BaseModel):
    author: Author
```

### Go (Static/Compiled)

```go
// ❌ This FAILS - Book not defined yet
type Author struct {
    Books []Book  // Compiler error: undefined: Book
}

type Book struct {
    Author Author
}
```

**Why it fails:** Go's compiler needs complete type information to:
1. Calculate struct sizes (memory layout)
2. Generate machine code
3. Perform type checking at compile time

---

## Go's Built-in Solutions

### Solution 1: Pointers (Most Common)

```go
// ✅ Works perfectly
type Author struct {
    Name  string
    Books []*Book  // Pointer allows forward reference
}

type Book struct {
    Title  string
    Author *Author  // Pointer allows circular reference
}
```

**Why pointers work:**
- Pointer size is always known (8 bytes on 64-bit)
- Compiler doesn't need to know `Book`'s full size yet
- Breaks infinite size recursion

**Python equivalent:**
```python
class Author(BaseModel):
    books: list['Book']  # Forward ref via string

Author.model_rebuild()  # Resolve at runtime
```

---

### Solution 2: Type Aliases + Package-level Organization

```go
// types.go
package myapp

// Forward declare via type alias
type BookList []Book

type Author struct {
    Books BookList
}

// Later in same file or different file in same package
type Book struct {
    Title  string
    Author *Author
}
```

**Why this works:**
- All files in a package are compiled together
- Compiler makes multiple passes
- Order doesn't matter within same package

---

## Comparison Table

| Feature          | Python              | Go                  |
|------------------|---------------------|---------------------|
| Type Resolution  | Runtime             | Compile-time        |
| Forward Refs     | String annotations  | Pointers solve it   |
| Circular Deps    | `.model_rebuild()`  | Pointers (built-in) |
| Performance Cost | Runtime overhead    | Zero runtime cost   |
| Complexity       | Need manual rebuild | Automatic           |
| Type Safety      | Runtime errors      | Compile-time errors |

---

## Conclusion

Go's **pointers** are the idiomatic equivalent of Python's `.model_rebuild()` — they solve circular dependencies at **compile-time** with **zero runtime cost** and **type safety guarantees**.