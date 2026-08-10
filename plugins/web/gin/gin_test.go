package gin

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	ginpkg "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginjson "github.com/gin-gonic/gin/codec/json"
	"github.com/stretchr/testify/require"

	"github.com/SmrutAI/pedantigo/v2/validator"
)

type ginTestRequest struct {
	Name  string `json:"name" form:"name" header:"X-Name" binding:"required,min=1,max=100"`
	Email string `json:"email" form:"email" header:"X-Email" binding:"required,email"`
}

var _ = validator.Register(validator.New[ginTestRequest](validator.Options{
	TagName:             "binding",
	StrictMissingFields: true,
}))

type panicValidateRequest struct {
	Name  string `form:"name" binding:"required"`
	Email string `form:"email" binding:"required,email"`
}

func (*panicValidateRequest) Validate() error {
	panic("boom")
}

var _ = validator.Register(validator.New[panicValidateRequest](validator.Options{TagName: "binding"}))

func TestMain(m *testing.M) {
	ginpkg.SetMode(ginpkg.TestMode)
	NewBinder()
	os.Exit(m.Run())
}

func TestBindJSONValid(t *testing.T) {
	c, _ := newJSONContext(http.MethodPost, `{"name":"Alice","email":"alice@example.com"}`)
	var req ginTestRequest

	err := c.ShouldBindJSON(&req)
	require.NoError(t, err)
	require.Equal(t, "Alice", req.Name)
	require.Equal(t, "alice@example.com", req.Email)
}

func TestBindJSONMissingRequiredField(t *testing.T) {
	c, _ := newJSONContext(http.MethodPost, `{"name":"Alice"}`)
	var req ginTestRequest

	err := c.ShouldBindJSON(&req)
	require.Error(t, err)

	ve, ok := AsValidationError(err)
	require.True(t, ok)
	require.NotNil(t, ve)
}

func TestBindQueryValid(t *testing.T) {
	c, _ := newQueryContext("/?name=Alice&email=alice@example.com")
	var req ginTestRequest

	err := c.ShouldBindQuery(&req)
	require.NoError(t, err)
	require.Equal(t, "Alice", req.Name)
	require.Equal(t, "alice@example.com", req.Email)
}

func TestBindQueryMissingRequiredField(t *testing.T) {
	c, _ := newQueryContext("/?name=Alice")
	var req ginTestRequest

	err := c.ShouldBindQuery(&req)
	// Gin has already populated the struct for query binding, so Pedantigo cannot
	// distinguish "key absent" from "key present with zero value" here.
	require.NoError(t, err)
}

func TestBindFormMissingRequiredField(t *testing.T) {
	c, _ := newFormContext("name=Alice")
	var req ginTestRequest

	err := c.ShouldBind(&req)
	// Gin has already populated the struct for form binding, so missing-field
	// required checks are not recoverable at Pedantigo's validation hook.
	require.NoError(t, err)
}

func TestBindFormInvalidEmail(t *testing.T) {
	c, _ := newFormContext("name=Alice&email=not-an-email")
	var req ginTestRequest

	err := c.ShouldBind(&req)
	require.Error(t, err)

	ve, ok := AsValidationError(err)
	require.True(t, ok)
	require.NotNil(t, ve)
}

func TestBindHeaderMissingRequiredField(t *testing.T) {
	c, _ := newHeaderContext("Alice", "")
	var req ginTestRequest

	err := c.ShouldBindHeader(&req)
	// Gin has already populated the struct for header binding, so missing-field
	// required checks are not recoverable at Pedantigo's validation hook.
	require.NoError(t, err)
}

func TestBindHeaderInvalidEmail(t *testing.T) {
	c, _ := newHeaderContext("Alice", "not-an-email")
	var req ginTestRequest

	err := c.ShouldBindHeader(&req)
	require.Error(t, err)

	ve, ok := AsValidationError(err)
	require.True(t, ok)
	require.NotNil(t, ve)
}

func TestPedantigoCodecMarshalPassThrough(t *testing.T) {
	codec := &pedantigoCodec{fallback: ginjson.API}
	payload := map[string]string{"name": "Alice"}

	got, err := codec.Marshal(payload)
	require.NoError(t, err)

	want, wantErr := ginjson.API.Marshal(payload)
	require.NoError(t, wantErr)
	require.Equal(t, want, got)
}

func TestPedantigoCodecUnmarshalPassThrough(t *testing.T) {
	codec := &pedantigoCodec{fallback: ginjson.API}
	data := []byte(`{"name":"Alice"}`)

	var got map[string]string
	err := codec.Unmarshal(data, &got)
	require.NoError(t, err)

	var want map[string]string
	wantErr := ginjson.API.Unmarshal(data, &want)
	require.NoError(t, wantErr)
	require.Equal(t, want, got)
}

func TestPedantigoCodecMarshalIndentPassThrough(t *testing.T) {
	codec := &pedantigoCodec{fallback: ginjson.API}
	payload := map[string]string{"name": "Alice"}

	got, err := codec.MarshalIndent(payload, "", "  ")
	require.NoError(t, err)

	want, wantErr := ginjson.API.MarshalIndent(payload, "", "  ")
	require.NoError(t, wantErr)
	require.Equal(t, want, got)
}

func TestPedantigoCodecNewEncoderPassThrough(t *testing.T) {
	codec := &pedantigoCodec{fallback: ginjson.API}
	payload := map[string]string{"name": "Alice"}

	var got bytes.Buffer
	err := codec.NewEncoder(&got).Encode(payload)
	require.NoError(t, err)

	var want bytes.Buffer
	wantErr := ginjson.API.NewEncoder(&want).Encode(payload)
	require.NoError(t, wantErr)
	require.Equal(t, want.String(), got.String())
}

func TestPedantigoCodecNewDecoderReturnsPedantigoDecoder(t *testing.T) {
	codec := &pedantigoCodec{fallback: ginjson.API}
	require.IsType(t, &pedantigoDecoder{}, codec.NewDecoder(strings.NewReader(`{}`)))
}

func TestDecodeUnregisteredTypeReturnsErrorNotPanic(t *testing.T) {
	type unregisteredRequest struct {
		Name string `json:"name"`
	}

	var req unregisteredRequest
	var err error
	require.NotPanics(t, func() {
		err = (&pedantigoDecoder{reader: strings.NewReader(`{"name":"Eve"}`)}).Decode(&req)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unregisteredRequest")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated read error")
}

func TestDecodeBodyReadError(t *testing.T) {
	var req ginTestRequest
	var err error
	require.NotPanics(t, func() {
		err = (&pedantigoDecoder{reader: errReader{}}).Decode(&req)
	})
	require.Error(t, err)
	require.EqualError(t, err, "simulated read error")
}

func TestValidateStructPanicInsideCustomValidateRecovered(t *testing.T) {
	c, _ := newQueryContext("/?name=Alice&email=alice@example.com")
	var req panicValidateRequest

	var err error
	require.NotPanics(t, func() {
		err = c.ShouldBindQuery(&req)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
}

func TestNewBinderPanicsOnEnableDecoderUseNumber(t *testing.T) {
	binding.EnableDecoderUseNumber = true
	defer func() {
		binding.EnableDecoderUseNumber = false
	}()

	require.Panics(t, func() {
		NewBinder()
	})
}

func TestNewBinderPanicsOnEnableDecoderDisallowUnknownFields(t *testing.T) {
	binding.EnableDecoderDisallowUnknownFields = true
	defer func() {
		binding.EnableDecoderDisallowUnknownFields = false
	}()

	require.Panics(t, func() {
		NewBinder()
	})
}

func TestNewBinderReinstallWithMatchingTagName(t *testing.T) {
	require.NotPanics(t, func() {
		NewBinder(WithTagName("binding"))
	})
}

func TestNewBinderReinstallWithConflictingTagName(t *testing.T) {
	require.Panics(t, func() {
		NewBinder(WithTagName("validate"))
	})
}

func TestAsValidationErrorTrue(t *testing.T) {
	c, _ := newFormContext("name=Alice&email=not-an-email")
	var req ginTestRequest

	err := c.ShouldBind(&req)
	ve, ok := AsValidationError(err)
	require.True(t, ok)
	require.NotNil(t, ve)
}

func TestAsValidationErrorFalse(t *testing.T) {
	ve, ok := AsValidationError(errors.New("x"))
	require.False(t, ok)
	require.Nil(t, ve)
}

func newJSONContext(method, body string) (*ginpkg.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := ginpkg.CreateTestContext(w)
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func newQueryContext(target string) (*ginpkg.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := ginpkg.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, http.NoBody)
	return c, w
}

func newFormContext(body string) (*ginpkg.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := ginpkg.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", binding.MIMEPOSTForm)
	c.Request = req
	return c, w
}

func newHeaderContext(name, email string) (*ginpkg.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := ginpkg.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	if name != "" {
		req.Header.Set("X-Name", name)
	}
	if email != "" {
		req.Header.Set("X-Email", email)
	}
	c.Request = req
	return c, w
}

var _ io.Reader = errReader{}
