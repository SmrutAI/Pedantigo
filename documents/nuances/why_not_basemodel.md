# Why Pedantigo Doesn't Have a BaseModel

## Short Answer

**External validators over BaseModel embedding** because BaseModel adds complexity with minimal benefit. The implementation requires initialization boilerplate, violates Go idioms, and provides only syntax changes—not new features.

---

## The Pydantic Pattern

```python
from pydantic import BaseModel

class User(BaseModel):
    name: str
    email: str

user = User(name="Alice", email="alice@example.com")
user.model_validate(data)      # Methods on instances
schema = User.model_json_schema()
```

---

## The Core Problem in Go

**Go's embedding cannot access parent struct fields:**

```go
type BaseModel struct {}

func (b BaseModel) Validate() error {
    // Receiver is BaseModel, NOT User
    // Can't see User's Name or Email fields
    // Reflection can't traverse "up" to container
}

type User struct {
    BaseModel
    Name  string
    Email string
}

user.Validate()  // Calls BaseModel.Validate(), but it can't access Name/Email
```

**Why:** When you call `user.Validate()`, Go invokes `BaseModel.Validate()` with receiver type `BaseModel`. The method only sees `BaseModel`'s fields, not `User`'s.

---

## Possible Implementation Approaches

### Option 1: Explicit Initialization

```go
type BaseModel struct {
    self interface{}  // Store reference to parent
}

func Init[T any](obj *T) {
    // Use reflection to set self reference
}

func (b *BaseModel) Validate() error {
    return validateViaReflection(b.self)
}

// Usage
user := &User{Name: "Alice"}
pedantigo.Init(user)  // ← MUST call this
user.Validate()
```

**Trade-off:** Works, but requires Init() on every instance. Zero value is broken.

### Option 2: Interface Contract

```go
type Validatable interface {
    GetValidationTarget() interface{}
}

// Every struct needs this boilerplate
func (u *User) GetValidationTarget() interface{} {
    return u
}
```

**Trade-off:** Works, but verbose. Defeats the purpose of "convenience."

### Option 3: Global Cache + Delegation

```go
var cache sync.Map

func (b *BaseModel) Validate() error {
    validator := getOrCreateValidator(reflect.TypeOf(b.self))
    return validator.Validate(b.self)  // Just wraps external validator
}
```

**Trade-off:** Works, but BaseModel is just a thin wrapper. Still needs Init().

---

## What Features Would BaseModel Actually Add?

### Syntax Changes (Not New Features)

| Feature | External Validator | BaseModel |
|---------|-------------------|-----------|
| Validate | `validator.Validate(&user)` | `user.Validate()` |
| Schema | `validator.Schema()` | `user.Schema()` |
| Marshal | `validator.Marshal(&user)` | `user.Marshal()` |
| Dict | `validator.Dict(&user)` | `user.Dict()` |

**All of these are already possible** - just different syntax.

### Features Already Supported

**Computed fields:**
```go
// No BaseModel needed - implement MarshalJSON()
func (u User) MarshalJSON() ([]byte, error) {
    type Alias User
    return json.Marshal(&struct {
        *Alias
        FullName string `json:"full_name"`
    }{
        Alias:    (*Alias)(&u),
        FullName: u.FirstName + " " + u.LastName,
    })
}
```

### Features That Don't Work in Go

**Immutability (Freeze()):**
- Can't intercept struct field writes in Go
- Would require getters/setters for all fields
- Not idiomatic

**Change tracking:**
- Same problem - can't intercept field assignments
- Would need setters for everything

---

## The Decision

### Complexity Analysis

**BaseModel implementation requires:**
- Init() function with reflection logic
- Global type-based cache with thread safety
- Documentation on initialization requirement
- Error handling for missing Init()
- Testing initialization edge cases

**External validator is:**
- Already implemented
- Well-tested
- Straightforward to use

### Go Idioms

**External validators align with Go:**
- Explicit over implicit (validator creation is visible)
- Separation of concerns (data vs behavior)
- Zero value is useful (structs work without Init)
- Composition (works with ANY struct, even from other packages)

**BaseModel violates:**
- Implicit (global cache, hidden initialization)
- Mixed concerns (data has methods it can't implement cleanly)
- Zero value broken (must call Init)
- Doesn't compose with external types

### Actual Usage Comparison

**BaseModel:**
```go
user := &User{Name: "Alice"}
pedantigo.Init(user)  // ← Initialize EVERY instance
user.Validate()
```

**External Validator:**
```go
validator := pedantigo.New[User]()  // ← Initialize ONCE per type
user := &User{Name: "Alice"}
validator.Validate(&user)
```

**Both require initialization. External is more efficient and clearer.**

---

## What We Provide

```go
validator := pedantigo.New[User]()

// Full API
validator.Validate(&user)
validator.Unmarshal(data)
validator.Marshal(&user)
validator.Schema()
validator.SchemaJSON()

// Works with ANY struct
type ThirdPartyStruct struct { ... }
validator := pedantigo.New[ThirdPartyStruct]()  // ✅ Works!
```

**Computed fields:** Document `MarshalJSON()` pattern (standard Go)

**Custom validation:** Document `Validate() error` interface (optional)

---

## For Pydantic Developers

Pydantic's BaseModel works because Python has:
- True inheritance (not embedding)
- Runtime type resolution
- Class-level initialization hooks

Go has:
- Composition (not inheritance)
- Compile-time type checking
- No automatic parent access from embedded types

**The external validator pattern IS the Go way.**

```go
// Idiomatic Go
validator := pedantigo.New[User]()
validator.Validate(&user)
```

Not `user.Validate()` - that's the Python way.

---

## Conclusion

BaseModel adds:
- ❌ Initialization boilerplate (Init on every instance)
- ❌ Global state (global cache)
- ❌ Violated Go idioms (zero value broken)
- ✅ Syntax sugar (methods on instances)

External validators provide:
- ✅ All the same features
- ✅ Idiomatic Go design
- ✅ Works with any struct
- ✅ Simpler implementation

**Syntax change doesn't justify the complexity.**
