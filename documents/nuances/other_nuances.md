# Pydantic vs. Pedantigo: Key Nuances

A focused guide for developers migrating from Python's Pydantic to Go's Pedantigo.

**Note:** Examples use Pydantic v2 syntax. For v1 migration, see notes in each section.

**Pydantic Documentation References:**
- [Fields](https://docs.pydantic.dev/latest/concepts/fields/)
- [Validators](https://docs.pydantic.dev/latest/concepts/validators/)
- [Models](https://docs.pydantic.dev/latest/concepts/models/)
- [Unions](https://docs.pydantic.dev/latest/concepts/unions/)
- [Errors](https://docs.pydantic.dev/latest/errors/errors/)

---

## 1. Field Defaults & Optional Values

**Pydantic:** Static defaults via assignment, dynamic via `Field(default_factory=callable)` ([docs](https://docs.pydantic.dev/latest/concepts/fields/#default-values))

```python
from pydantic import BaseModel, Field
from datetime import datetime

class User(BaseModel):
    role: str = "user"                              # Static
    created_at: datetime = Field(default_factory=datetime.now)  # Dynamic
```

**Pedantigo:** Static via `default=` tag, dynamic via `defaultFactory=MethodName` tag

```go
type User struct {
    Role      string    `validate:"default=user"`              // Static
    CreatedAt time.Time `validate:"defaultFactory=Now"`        // Dynamic
}

// Method must have signature: func (t *T) MethodName() (FieldType, error)
func (u *User) Now() (time.Time, error) {
    return time.Now(), nil
}
```

**Optional fields:** Pydantic uses `Optional[T]` or `T | None`, Pedantigo uses pointer types `*T`.

---

## 2. Validators (Field-level Constraints)

**Pydantic:** Custom validation logic using `@field_validator` decorator ([docs](https://docs.pydantic.dev/latest/concepts/validators/#field-validators))

```python
from pydantic import BaseModel, field_validator

class User(BaseModel):
    email: str

    @field_validator('email')
    @classmethod
    def validate_email(cls, v: str) -> str:
        if '@' not in v:
            raise ValueError('invalid email')
        return v
```

*Note: Pydantic v1 used `@validator` - this is legacy syntax.*

**Pedantigo:** Declarative struct tags using built-in constraints

```go
type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"required,min=0,max=120"`
}
```

Built-in constraints include: `email`, `url`, `uuid`, `min`, `max`, `len`, `regexp`, and many others.

---

## 3. Cross-field Validation

**Pydantic:** Write custom validation logic using `@model_validator` decorator ([docs](https://docs.pydantic.dev/latest/concepts/validators/#model-validators))

```python
from pydantic import BaseModel, model_validator

class DateRange(BaseModel):
    start_date: datetime
    end_date: datetime

    @model_validator(mode='after')
    def check_dates(self) -> Self:
        if self.end_date < self.start_date:
            raise ValueError('end must be after start')
        return self
```

*Note: Pydantic v1 used `@root_validator` - this is legacy syntax.*

**Pedantigo:** Simple comparisons use declarative tags, complex logic uses `Validate()` method

```go
// Simple field comparison - use declarative tags
type DateRange struct {
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`
}

// Complex conditional logic - implement Validate()
type User struct {
    AuthMethod string `json:"auth_method"`
    Password   string `json:"password"`
    Email      string `json:"email"`
    Phone      string `json:"phone"`
}

func (u *User) Validate() error {
    if u.AuthMethod == "email" && u.Password == "" {
        return errors.New("password required when using email auth")
    }

    if u.Email == "" && u.Phone == "" {
        return errors.New("must provide email or phone")
    }

    return nil
}
```

**Key difference:** Pydantic requires custom code for all cross-field validation. Pedantigo offers declarative tags (`gtfield`, `gtefield`, `ltfield`, `ltefield`) for simple field comparisons, reducing boilerplate.

**Note:** `Validate()` is completely optional in Pedantigo. Only implement when struct tags can't express your validation logic.

---

## 4. Type Coercion

**Pydantic:** Automatically coerces compatible types ([docs](https://docs.pydantic.dev/latest/concepts/models/#data-conversion))

```python
from pydantic import BaseModel

class Config(BaseModel):
    port: int

config = Config(port="8080")  # ✅ Converts "8080" to 8080
```

**Pedantigo:** Strict typing - no automatic coercion

```go
type Config struct {
    Port int `json:"port"`
}

json := `{"port": "8080"}`
validator.Unmarshal([]byte(json))  // ❌ Error: cannot unmarshal string into int
```

You must provide the correct type in JSON. Use `json.Number` if you need flexible numeric parsing.

---

## 5. Immutability

**Pydantic:** `frozen=True` in model config makes models immutable ([docs](https://docs.pydantic.dev/latest/concepts/models/#frozen-models))

```python
from pydantic import BaseModel, ConfigDict

class User(BaseModel):
    model_config = ConfigDict(frozen=True)

    name: str

user = User(name="Alice")
user.name = "Bob"  # ❌ Raises ValidationError
```

*Note: Pydantic v1 used `class Config: frozen = True` - this is legacy syntax.*

**Pedantigo:** Go has no built-in struct immutability

```go
user := User{Name: "Alice"}
user.Name = "Bob"  // ✅ Allowed - cannot prevent at compile-time
```

Workaround: Use unexported fields with getter methods, but mutations cannot be prevented at compile-time.

---

## 6. Computed/Derived Fields

**Pydantic:** `@computed_field` decorator automatically includes property in serialization ([docs](https://docs.pydantic.dev/latest/concepts/fields/#computed-fields))

```python
from pydantic import BaseModel, computed_field

class User(BaseModel):
    first_name: str
    last_name: str

    @computed_field
    @property
    def full_name(self) -> str:
        return f"{self.first_name} {self.last_name}"

# Automatically appears in .model_dump() and JSON output
```

**Pedantigo:** Implement custom `MarshalJSON()` method to include computed fields

```go
type User struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

func (u User) MarshalJSON() ([]byte, error) {
    type Alias User  // Avoid recursion
    return json.Marshal(&struct {
        *Alias
        FullName string `json:"full_name"`
    }{
        Alias:    (*Alias)(&u),
        FullName: u.FirstName + " " + u.LastName,
    })
}
```

**Key difference:** Pydantic uses declarative decorator, Go requires implementing `MarshalJSON()` interface. Both achieve the same result - computed fields in JSON output.

---

## 7. Private Attributes

**Pydantic:** Use `PrivateAttr()` for fields excluded from validation/serialization ([docs](https://docs.pydantic.dev/latest/api/fields/#pydantic.fields.PrivateAttr))

```python
from pydantic import BaseModel, PrivateAttr

class User(BaseModel):
    name: str
    _internal_id: int = PrivateAttr(default=0)
```

**Pedantigo:** Use unexported fields (lowercase first letter)

```go
type User struct {
    Name       string `json:"name"`
    internalID int    // Automatically excluded from JSON and validation
}
```

---

## 8. Discriminated Unions

**Pydantic:** Built-in support with `Union` and `Field(discriminator=...)` ([docs](https://docs.pydantic.dev/latest/concepts/unions/#discriminated-unions))

```python
from typing import Literal, Union
from pydantic import BaseModel, Field

class Cat(BaseModel):
    pet_type: Literal["cat"]
    meow_volume: int

class Dog(BaseModel):
    pet_type: Literal["dog"]
    bark_pitch: int

class Model(BaseModel):
    pet: Union[Cat, Dog] = Field(discriminator='pet_type')
```

**Pedantigo:** Currently requires manual implementation with interfaces

```go
type Pet interface {
    GetType() string
}

type Cat struct {
    Type       string `json:"type"`
    MeowVolume int    `json:"meow_volume"`
}

type Dog struct {
    Type      string `json:"type"`
    BarkPitch int    `json:"bark_pitch"`
}

func (c Cat) GetType() string { return "cat" }
func (d Dog) GetType() string { return "dog" }
```

Discriminated union support is now available via `NewUnion()`:

```go
validator, _ := pedantigo.NewUnion[any](pedantigo.UnionOptions{
    DiscriminatorField: "pet_type",
    Variants: []pedantigo.UnionVariant{
        pedantigo.VariantFor[Cat]("cat"),
        pedantigo.VariantFor[Dog]("dog"),
    },
})

result, _ := validator.Unmarshal(jsonData)
```

---

## 9. Model Configuration

**Pydantic:** Use `model_config = ConfigDict(...)` for model-level settings ([docs](https://docs.pydantic.dev/latest/concepts/models/#model-config))

```python
from pydantic import BaseModel, ConfigDict

class User(BaseModel):
    model_config = ConfigDict(
        validate_assignment=True,
        arbitrary_types_allowed=True
    )

    name: str
```

*Note: Pydantic v1 used `class Config:` - this is legacy syntax.*

**Pedantigo:** Sensible defaults built into library - no config class needed

```go
validator := pedantigo.New[User]()  // Uses sensible defaults
```

---

## 10. Serialization Aliases

**Pydantic:** Use `Field(alias=...)` to specify JSON key names ([docs](https://docs.pydantic.dev/latest/concepts/fields/#field-aliases))

```python
from pydantic import BaseModel, Field

class User(BaseModel):
    name: str = Field(alias='userName')
```

**Pedantigo:** Use standard Go `json` struct tags

```go
type User struct {
    Name string `json:"userName"`  // JSON key is "userName", Go field is "Name"
}
```

---

## 11. Error Accumulation

Both libraries accumulate all validation errors and return them together (don't fail on first error).

**Pydantic:** Returns `ValidationError` exception with `.errors()` method ([docs](https://docs.pydantic.dev/latest/errors/errors/))

```python
from pydantic import BaseModel, ValidationError

class User(BaseModel):
    email: str
    age: int

try:
    User(email="invalid", age=-5)
except ValidationError as e:
    for error in e.errors():
        print(f"{error['loc']}: {error['msg']}")
```

**Pedantigo:** Returns `*ValidationError` with `Errors []FieldError`

```go
_, err := validator.Unmarshal(jsonData)
if ve, ok := err.(*pedantigo.ValidationError); ok {
    for _, fe := range ve.Errors {
        fmt.Printf("%s: %s (code: %s)\n", fe.Field, fe.Message, fe.Code)
    }
}
```

---

## 12. Inheritance vs. Composition

**Pydantic:** Uses class inheritance

```python
class BaseUser(BaseModel):
    name: str
    email: str

class AdminUser(BaseUser):  # Inherits name and email
    role: str = "admin"
```

**Pedantigo:** Uses struct embedding (composition, not inheritance)

```go
type BaseUser struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type AdminUser struct {
    BaseUser  // Embedded - fields are "promoted" to outer struct
    Role string `json:"role" validate:"default=admin"`
}
```

Embedded fields are promoted to the outer struct, achieving similar ergonomics without true inheritance.

---

## Summary Table

| Feature | Pydantic v2 | Pedantigo |
|---------|-------------|-----------|
| **Defaults** | `Field(default=...)`, `Field(default_factory=...)` | `validate:"default=..."`, `validate:"defaultFactory=Method"` |
| **Validators** | `@field_validator` decorator | `validate:"email,min=5,max=100"` tags |
| **Cross-field** | `@model_validator` decorator | Cross-field tags or `Validate() error` method |
| **Type Coercion** | Automatic | Strict (no coercion) |
| **Optional** | `Optional[T]` or `T \| None` | `*T` (pointer types) |
| **Immutability** | `model_config = ConfigDict(frozen=True)` | Not supported |
| **Computed Fields** | `@computed_field` decorator | `MarshalJSON()` method |
| **Private Attrs** | `PrivateAttr()` | Unexported fields (lowercase) |
| **Unions** | `Union[A, B]` with discriminator | Interfaces + manual handling (Phase 5 planned) |
| **Config** | `model_config = ConfigDict(...)` | Library defaults (no config needed) |
| **Aliases** | `Field(alias=...)` | `json:"alias"` tags |
| **Errors** | `ValidationError.errors()` | `[]ValidationError` slice |
| **Inheritance** | Class inheritance | Struct embedding |

---

## Migration Tips

1. **Start with struct tags** - Most Pydantic validators map to Pedantigo tags (`email`, `url`, `min`, `max`, etc.)
2. **Use pointers for optional fields** - Replace `Optional[T]` with `*T`
3. **Implement `Validate()` for complex logic** - Only when struct tags can't express your validation
4. **Remember: Go is strict** - No automatic type coercion like Pydantic
5. **Composition over inheritance** - Use struct embedding instead of class inheritance
6. **Check error slices** - Pedantigo returns all errors at once like Pydantic