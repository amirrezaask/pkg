package openapi

type Document struct {
	OpenAPI           string
	Info              Info
	jsonSchemaDialect string
	Servers           []Server
	Paths             map[string]PathItem
	Components        Components
	// Webhooks
	Security []SecurityRequirement
}

type Info struct {
	Title          string
	Summary        string
	Description    string
	TermsOfService string
	Contact        Contact
	License        License
	Version        string
}

type Contact struct {
	Name  string
	URL   string
	Email string
}

type License struct {
	Name       string
	Identifier string
	URL        string
}

type Server struct {
	URL         string
	Description string
	Variables   map[string]ServerVariable
}

type ServerVariable struct {
	Enum        []string
	Default     string
	Description string
}

type PathItem struct {
	Ref         string `json:"$ref"`
	Summary     string
	Description string
	Get         Operation
	Put         Operation
	Post        Operation
	Delete      Operation
	Options     Operation
	Head        Operation
	Patch       Operation
	Trace       Operation
	Servers     []Server
	Parameters  []Parameter
}

type Parameter struct {
	Name            string
	In              string
	Description     string
	Required        bool
	Depreacated     bool
	AllowEmptyValue bool
}

type SecurityRequirement struct {
}

type ExternalDocumentation struct{}

type Operation struct {
	Tags         []string
	Summary      string
	Description  string
	ExternalDocs ExternalDocumentation
	OperationId  string
	Parameters   []Parameter
	RequestBody  RequestBody
	Responses    map[string]Response
}

type MediaType struct {
	Schema   Schema
	Example  any
	Examples map[string]Example
	Encoding map[string]Encoding
}
type Response struct {
	Description string
	Headers     map[string]Header
	Content     map[string]MediaType
	Links       map[string]Link
}

type Schema struct {
}
type Encoding struct {
	ContentType   string
	Headers       map[string]Header
	Style         string
	Explode       string
	AllowReserved string
}
type Example struct{}

type RequestBody struct{}

type Header struct{}

type SecuritySchema struct{}

type Link struct{}

type Components struct {
	Schemas         map[string]Schema
	Responses       map[string]Response
	Parameters      map[string]Parameter
	Examples        map[string]Example
	RequestBodies   map[string]RequestBody
	Headers         map[string]Header
	SecuritySchemas map[string]SecuritySchema
	Links           map[string]Link
	PathItems       map[string]PathItem
}
