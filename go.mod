module github.com/SmrutAI/pedantigo/v2

go 1.21

retract v2.0.0 // published with a stale layout (pre validator/ restructure) that was permanently cached on the module proxy before the fix landed; use v2.0.1 or later
retract v2.1.0 // README/docs pointed at plugins/web/echo and plugins/web/gin v2.1.0 tags whose module paths broke Go's major-version-suffix rule (v2 not at the end of the path); those tags have been deleted, so following this version's own install instructions fails; use v2.1.1 or later with plugins/web/pedantigoecho/v2 and plugins/web/pedantigogin/v2

require (
	github.com/invopop/jsonschema v0.13.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/stretchr/testify v1.11.1
	golang.org/x/text v0.14.0
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
