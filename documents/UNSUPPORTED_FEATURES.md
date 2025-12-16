# Unsupported Features

Features from Pydantic v2 and go-playground/validator that Pedantigo does not support, with explanations.

---

## Quick Reference

| Feature                         | Source    | Reason                                 | Workaround                                       |
|---------------------------------|-----------|----------------------------------------|--------------------------------------------------|
| Immutable structs               | Pydantic  | Go can't intercept field writes        | Use unexported fields + getters                  |
| Validate on assignment          | Pydantic  | Go can't intercept field writes        | Call `Validate()` after mutations                |
| Before validators               | Pydantic  | Go has no annotation system            | Transform in `NewModel()` or custom deserializer |
| Wrap validators                 | Pydantic  | Go has no annotation system            | Implement `Validatable` interface                |
| Type coercion                   | Pydantic  | Go is statically typed                 | Use correct types in JSON                        |
| Generic[T] BaseModel            | Pydantic  | Go can't construct types at runtime    | Define concrete types per variant                |
| ORM mode                        | Pydantic  | Python-specific (SQLAlchemy)           | Manual struct conversion                         |
| BaseSettings                    | Pydantic  | Python-specific (dotenv, env vars)     | Use `envconfig` or `viper`                       |
| TypeAdapter                     | Pydantic  | Python's dynamic typing                | Use `Validator[T]` directly                      |
| RootModel                       | Pydantic  | Python-specific pattern                | Use wrapper struct                               |
| ~~Cross-struct validation~~     | validator | ✅ **NOW SUPPORTED** (v1.x)            | Use `eqfield=Inner.Field` dotted notation        |
| Validator context               | validator | Adds complexity                        | Use struct fields for context                    |

---

## Detailed Explanations

### 1. Immutable/Frozen Structs

**Source:** Pydantic `frozen=True`

**Pydantic:**
```python
class User(BaseModel):
    model_config = ConfigDict(frozen=True)
    name: str

user = User(name="Alice")
user.name = "Bob"  # ❌ Raises ValidationError
```

**Why not in Go:**
- Go has no mechanism to intercept struct field writes
- Would require getter/setter for every field (massive boilerplate)
- Not idiomatic Go - breaks expectations

**Workaround:**
```go
type User struct {
    name string  // unexported
}

func (u *User) Name() string { return u.name }  // read-only access
```

**See:** [documents/nuances/why_not_basemodel.md](nuances/why_not_basemodel.md)

---

### 2. Validate on Assignment

**Source:** Pydantic `validate_assignment=True`

**Pydantic:**
```python
class User(BaseModel):
    model_config = ConfigDict(validate_assignment=True)
    age: int = Field(ge=0)

user = User(age=25)
user.age = -5  # ❌ Raises ValidationError immediately
```

**Why not in Go:**
- Same as immutable structs - Go can't intercept field assignments
- Would need setters for every field
- Defeats purpose of struct tags

**Workaround:**
```go
user.Age = -5
if err := validator.Validate(&user); err != nil {
    // Handle validation error
}
```

---

### 3. Before/Wrap Validators

**Source:** Pydantic `mode='before'`, `mode='wrap'`

**Pydantic:**
```python
class User(BaseModel):
    name: str

    @field_validator('name', mode='before')
    @classmethod
    def strip_name(cls, v):
        return v.strip() if isinstance(v, str) else v
```

**Why not in Go:**
- Go has no decorator/annotation system
- Can't inject code before field assignment during unmarshaling
- Reflection can observe but not modify unmarshaling process

**Workaround:**
```go
// Use NewModel() which can transform data
user, err := validator.NewModel(input)

// Or implement custom UnmarshalJSON
func (u *User) UnmarshalJSON(data []byte) error {
    // Transform before setting fields
}
```

**Note:** String transformations (`strip_whitespace`, `to_lower`, `to_upper`) ARE supported via tags. In `Unmarshal()`/`NewModel()` they transform the data; in `Validate()` they check format.

---

### 4. Type Coercion

**Source:** Pydantic automatic coercion

**Pydantic:**
```python
class Config(BaseModel):
    port: int

config = Config(port="8080")  # ✅ Converts "8080" to 8080
```

**Why not in Go:**
- Go is statically typed - `"8080"` is a string, not an int
- `encoding/json` errors on type mismatch
- Implicit coercion violates Go's explicit philosophy

**Workaround:**
```go
// Provide correct types in JSON
json := `{"port": 8080}`  // number, not string

// Or use json.Number for flexible parsing
type Config struct {
    Port json.Number `json:"port"`
}
```

---

### 5. Generic Structs (Pydantic's Dynamic Pattern)

**Source:** Pydantic `Generic[T]`

**Pydantic:**
```python
from typing import Generic, TypeVar
T = TypeVar('T')

class Response(BaseModel, Generic[T]):
    data: T
    status: int

# Construct type at runtime
response = Response[User](data=user, status=200)
schema = Response[User].model_json_schema()  # Schema for User
schema = Response[Order].model_json_schema() # Schema for Order - same class!
```

**Why not in Go:**
- Go requires concrete types at compile time
- Can't construct `Response[SomeType]` dynamically from a string/variable
- Each `Response[User]` and `Response[Order]` is a distinct type at compile time

**Note:** Go generics DO work with reflection. Pedantigo's `Validator[T]` uses this. The limitation is runtime type construction, not reflection.

**Workaround:**
```go
// Define concrete types (Go's approach)
type UserResponse struct {
    Data   User `json:"data"`
    Status int  `json:"status"`
}

type OrderResponse struct {
    Data   Order `json:"data"`
    Status int   `json:"status"`
}

// Both work with Pedantigo
userValidator := pedantigo.New[UserResponse]()
orderValidator := pedantigo.New[OrderResponse]()
```

---

### 6. ORM Mode / from_attributes

**Source:** Pydantic `from_attributes=True`

**Pydantic:**
```python
# In Python: SQLAlchemy model and Pydantic model are SEPARATE
class UserORM(Base):  # SQLAlchemy
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    name = Column(String)
    email = Column(String)

class UserOut(BaseModel):  # Pydantic - separate class!
    model_config = ConfigDict(from_attributes=True)
    id: int
    name: str
    email: str

# Need conversion layer
db_user = session.query(UserORM).first()
user_out = UserOut.model_validate(db_user)  # from_attributes reads object attrs
```

**Why Go doesn't need this:**

In Python, SQLAlchemy models and Pydantic models are fundamentally different class hierarchies - you MUST define separate classes and convert between them.

In Go, struct tags allow ONE type to serve multiple purposes:

```go
// GORM, JSON, and Pedantigo use the SAME struct
type User struct {
    ID    uint   `gorm:"primaryKey" json:"id" pedantigo:"required"`
    Name  string `gorm:"column:name" json:"name" pedantigo:"min=1,max=100"`
    Email string `gorm:"uniqueIndex" json:"email" pedantigo:"required,email"`
}

// Query from database
var user User
db.First(&user, 1)

// Validate directly - no conversion!
err := validator.Validate(&user)

// Marshal directly - no conversion!
data, _ := validator.Marshal(user)

// Unmarshal from API, validate, save to DB - same struct throughout
var newUser User
validator.Unmarshal(jsonData)  // Validates during unmarshal
db.Create(&newUser)
```

**Go's advantage:** Struct tags eliminate the "two model" problem. GORM, sqlx, ent, and other Go ORMs all use struct tags, making them directly compatible with Pedantigo.

**If you truly need conversion** (rare - usually means design issue):
```go
// Manual field copy
func ToUserOut(orm *UserORM) UserOut {
    return UserOut{ID: orm.ID, Name: orm.Name}
}

// Or use a struct copying library like copier
copier.Copy(&userOut, &userORM)
```

---

### 7. BaseSettings / Environment Variables

**Source:** Pydantic `BaseSettings`

**Pydantic:**
```python
class Settings(BaseSettings):
    database_url: str
    debug: bool = False

    class Config:
        env_file = '.env'

settings = Settings()  # Auto-loads from environment
```

**Why not in Go:**
- Python-specific pattern
- Go has mature alternatives
- Different philosophy (explicit vs magic)

**Workaround:**
```go
// Use envconfig
import "github.com/kelseyhightower/envconfig"

type Settings struct {
    DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
    Debug       bool   `envconfig:"DEBUG" default:"false"`
}

envconfig.Process("", &settings)

// Or use viper for complex configs
import "github.com/spf13/viper"
```

---

### 8. TypeAdapter

**Source:** Pydantic `TypeAdapter`

**Pydantic:**
```python
adapter = TypeAdapter(list[int])
result = adapter.validate_python([1, 2, "3"])  # Validates + coerces
```

**Why not in Go:**
- Python's dynamic typing allows runtime type construction
- Go types are fixed at compile time
- `Validator[T]` already handles this case

**Workaround:**
```go
// Define type explicitly
type IntList []int

validator := pedantigo.New[IntList]()
```

---

### 9. RootModel

**Source:** Pydantic `RootModel`

**Pydantic:**
```python
class UserList(RootModel[list[User]]):
    pass

users = UserList.model_validate([{"name": "Alice"}, {"name": "Bob"}])
```

**Why not in Go:**
- Can't have methods on `[]T` in Go
- Would need wrapper struct anyway
- Validator already supports slices directly

**Workaround:**
```go
type UserList struct {
    Users []User `json:"users"`
}

// Or validate slice directly
validator := pedantigo.New[[]User]()
```

---

### 10. Cross-Struct Validation ✅ NOW SUPPORTED

**Source:** go-playground/validator `eqcsfield`, `necsfield`, etc.

**Pedantigo now supports nested field references via dotted notation:**
```go
type Outer struct {
    Inner Inner
    Max   int `pedantigo:"gtfield=Inner.Value"`  // Compare across nested structs
    Check int `pedantigo:"eqfield=Inner.MinValue"` // Works with deep nesting too
}
```

**Features:**
- Supports any nesting depth: `eqfield=A.B.C.Field`
- All cross-field operators: `eqfield`, `nefield`, `gtfield`, `gtefield`, `ltfield`, `ltefield`
- Handles pointers automatically
- Validates field paths at `New()` time (fail-fast)

---

### 11. Validator Context

**Source:** go-playground/validator `FieldLevel`

**validator:**
```go
validate.RegisterValidationCtx("custom", func(ctx context.Context, fl validator.FieldLevel) bool {
    userID := ctx.Value("user_id").(string)
    // Use context in validation
})
```

**Why not in Pedantigo:**
- Adds complexity to validation API
- Context belongs at service layer, not field level
- Can use struct fields to pass context

**Workaround:**
```go
type Request struct {
    UserID string  // Include context as field
    Data   string `pedantigo:"required"`
}

func (r *Request) Validate() error {
    // Use r.UserID for context-aware validation
}
```

---

## Features That May Be Added Later

These are not currently supported but could be added if demand exists:

| Feature | Complexity | Notes |
|---------|------------|-------|
| Strict types (StrictInt, etc.) | Medium | Planned for Phase 12 |
| Set/Tuple types | Low | Not idiomatic in Go |
| Alias generators | Medium | Could use go:generate |
| i18n/l10n | High | Community contribution welcome |

**Recently Implemented** (removed from this list):
- ✅ Secret types (`SecretStr`, `SecretBytes`) - Masks sensitive data in String/JSON
- ✅ Path validation (`filepath`, `dirpath`, `file`, `dir`) - Syntax and existence checks
- ✅ `time.Duration` - Parses duration strings like "1h30m", "500ms"
- ✅ ISO codes (`iso3166_alpha2`, `iso4217`, `bcp47`, `postcode`) - Country, currency, language, postal codes

---

## Philosophy

Pedantigo prioritizes:

1. **Go idioms** over Pydantic patterns
2. **Compile-time safety** over runtime magic
3. **Explicit behavior** over implicit coercion
4. **Simplicity** over feature completeness

If a Pydantic feature requires fighting Go's design, we document the workaround rather than add non-idiomatic code.
