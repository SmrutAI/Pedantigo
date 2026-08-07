# Computed/Derived Fields in Go vs. Python

## Short Answer

**Yes, Go fully supports computed fields in JSON** — implement the `MarshalJSON()` interface method.

---

## Why They Work Differently

### Python (Declarative/Decorator-based)

```python
from pydantic import BaseModel, computed_field

class User(BaseModel):
    first_name: str
    last_name: str

    @computed_field
    @property
    def full_name(self) -> str:
        return f"{self.first_name} {self.last_name}"

# Automatically in JSON
user = User(first_name="John", last_name="Doe")
print(user.model_dump())
# {'first_name': 'John', 'last_name': 'Doe', 'full_name': 'John Doe'}
```

**Why it works:**
- Decorators (`@computed_field`, `@property`) mark methods for special treatment
- Pydantic's serializer detects these decorators at runtime
- Automatically includes them in `.model_dump()` and JSON output

### Go (Interface-based)

```go
type User struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

func (u User) MarshalJSON() ([]byte, error) {
    type Alias User  // Prevent infinite recursion
    return json.Marshal(&struct {
        *Alias
        FullName string `json:"full_name"`
    }{
        Alias:    (*Alias)(&u),
        FullName: u.FirstName + " " + u.LastName,
    })
}

// Usage
user := User{FirstName: "John", LastName: "Doe"}
json.Marshal(user)
// {"first_name":"John","last_name":"Doe","full_name":"John Doe"}
```

**Why it works:**
- `json.Marshal()` checks if type implements `json.Marshaler` interface
- Interface: `type Marshaler interface { MarshalJSON() ([]byte, error) }`
- If found, calls custom method instead of default serialization
- Type alias prevents infinite loop (no `MarshalJSON()` on alias)

---

## Go's MarshalJSON Pattern

### Standard Pattern

```go
func (u User) MarshalJSON() ([]byte, error) {
    type Alias User              // 1. Create type alias (no methods attached)
    return json.Marshal(&struct {
        *Alias                   // 2. Embed original fields
        ComputedField Type       // 3. Add computed fields
    }{
        Alias:         (*Alias)(&u),     // 4. Cast to alias (avoids recursion)
        ComputedField: u.ComputeValue(), // 5. Populate computed values
    })
}
```

### Why Type Alias is Required

```go
// ❌ WRONG - Infinite recursion
func (u User) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        User                    // Embeds User type (has MarshalJSON method)
        Extra string
    }{
        User:  u,               // Calls u.MarshalJSON() → infinite loop
        Extra: "value",
    })
}

// ✅ CORRECT - Type alias breaks cycle
func (u User) MarshalJSON() ([]byte, error) {
    type Alias User            // New type, no MarshalJSON method attached
    return json.Marshal(struct {
        *Alias                 // Embeds Alias type (no MarshalJSON)
        Extra string
    }{
        Alias: (*Alias)(&u),   // Safe - uses default JSON marshaling
        Extra: "value",
    })
}
```

---

## Comparison Table

| Feature          | Python (Pydantic)         | Go                       |
|------------------|---------------------------|--------------------------|
| Approach         | Declarative decorators    | Interface method         |
| Syntax           | `@computed_field`         | `MarshalJSON()`          |
| Detection        | Runtime decorator scan    | Interface check          |
| Boilerplate      | Minimal                   | Moderate (type alias)    |
| Recursion Safety | Automatic                 | Manual (type alias)      |
| Flexibility      | Property pattern only     | Full JSON control        |
| Performance Cost | Decorator overhead        | Zero runtime cost        |

---

## Conclusion

Go's **`MarshalJSON()` interface** is the idiomatic equivalent of Python's `@computed_field` decorator. While Go requires more boilerplate (type alias to prevent recursion), it provides **compile-time type safety** and **zero runtime overhead** compared to Python's decorator-based approach.