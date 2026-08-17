package pedantigoecho

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

type TestRequest struct {
	Name  string `json:"name" query:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" query:"email" validate:"required,email"`
}

var _ = validator.Register(validator.New[TestRequest]())

type unregisteredRequest struct {
	Name string `json:"name"`
}

func TestBind_PostValidBody(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Alice","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.NoError(t, err)
	assert.Equal(t, "Alice", out.Name)
	assert.Equal(t, "alice@example.com", out.Email)
}

func TestBind_PostMissingRequiredField(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.Error(t, err)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr, "expected *echo.HTTPError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestBind_PostInvalidEmail(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Alice","email":"not-an-email"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.Error(t, err)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr, "expected *echo.HTTPError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestBind_PostEmptyBody(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.NoError(t, err)
	assert.Equal(t, TestRequest{}, out)
}

func TestBind_PutValidBody(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Bob","email":"bob@example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.NoError(t, err)
	assert.Equal(t, "Bob", out.Name)
	assert.Equal(t, "bob@example.com", out.Email)
}

func TestBind_GetFallsBackToDefaultBinder(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	req := httptest.NewRequest(http.MethodGet, "/?name=Carol&email=carol@example.com", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.NoError(t, err)
	assert.Equal(t, "Carol", out.Name)
	assert.Equal(t, "carol@example.com", out.Email)
}

func TestBind_DeleteFallsBackToDefaultBinder(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	req := httptest.NewRequest(http.MethodDelete, "/?name=Dave&email=dave@example.com", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.NoError(t, err)
	assert.Equal(t, "Dave", out.Name)
}

type pathParamTestRequest struct {
	ID   int    `json:"id" param:"id"`
	Name string `json:"name" validate:"required"`
}

var _ = validator.Register(validator.New[pathParamTestRequest]())

// TestBind_PathParamBindError covers the BindPathParams error branch: a
// non-numeric path value bound to an int field fails type conversion.
func TestBind_PathParamBindError(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-number")

	var out pathParamTestRequest
	err := c.Bind(&out)

	require.Error(t, err)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr, "expected *echo.HTTPError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

// errReader is an io.Reader whose Read always fails, used to exercise the
// io.ReadAll error branch in Bind.
type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated read error")
}

func TestBind_BodyReadError(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	req := httptest.NewRequest(http.MethodPost, "/", errReader{})
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out TestRequest
	err := c.Bind(&out)

	require.Error(t, err)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr, "expected *echo.HTTPError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, "failed to read body", httpErr.Message)
}

func TestBind_UnregisteredTypePanics(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := `{"name":"Eve"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var out unregisteredRequest
	assert.Panics(t, func() {
		_ = c.Bind(&out)
	})
}

// TestRegister_Behavior verifies Register() using a type that is NOT registered
// at package level (unlike TestRequest, which is registered once via the
// package-level var above and shared by the Bind tests). Both assertions run
// as ordered subtests within a single Test function so the "first registration
// succeeds" step is a verified precondition of "second registration panics",
// not an assumption about package-init timing or file-level test ordering.
func TestRegister_Behavior(t *testing.T) {
	type registrationTestRequest struct {
		Name string `json:"name" validate:"required"`
	}

	var registered *validator.Validator[registrationTestRequest]

	t.Run("first registration succeeds", func(t *testing.T) {
		require.NotPanics(t, func() {
			registered = validator.Register(validator.New[registrationTestRequest]())
		})
		require.NotNil(t, registered)
	})

	t.Run("second registration for the same type panics", func(t *testing.T) {
		assert.Panics(t, func() {
			validator.Register(validator.New[registrationTestRequest]())
		})
	})
}
