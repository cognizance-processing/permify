// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/permissions/check": {
            "post": {
                "description": "Check subject is authorized",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Permission"
                ],
                "summary": "Permission",
                "operationId": "check",
                "parameters": [
                    {
                        "description": "''",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/permission.Check"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Check"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        },
        "/relationships/delete": {
            "post": {
                "description": "delete relation tuple",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Relationship"
                ],
                "summary": "Relationship",
                "operationId": "delete",
                "parameters": [
                    {
                        "description": "''",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/relationship.Delete"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Message"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        },
        "/relationships/write": {
            "post": {
                "description": "create new relation tuple",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Relationship"
                ],
                "summary": "Relationship",
                "operationId": "write",
                "parameters": [
                    {
                        "description": "''",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/relationship.Write"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Message"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        },
        "/schemas/replace": {
            "post": {
                "description": "replace your authorization model",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Schema"
                ],
                "summary": "Schema",
                "operationId": "replace",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Message"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        },
        "/status/ping": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Server"
                ],
                "summary": "Server",
                "operationId": "ping",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Message"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        },
        "/status/version": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Server"
                ],
                "summary": "Server",
                "operationId": "version",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.Message"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.HTTPErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "permission.Check": {
            "type": "object",
            "properties": {
                "body": {
                    "description": "*\n\t * Body",
                    "type": "object",
                    "properties": {
                        "action": {
                            "type": "string"
                        },
                        "depth": {
                            "type": "integer"
                        },
                        "object": {
                            "type": "string"
                        },
                        "user": {
                            "type": "string"
                        }
                    }
                },
                "pathParams": {
                    "description": "*\n\t * PathParams",
                    "type": "object"
                },
                "queryParams": {
                    "description": "*\n\t * QueryParams",
                    "type": "object"
                }
            }
        },
        "relationship.Delete": {
            "type": "object",
            "properties": {
                "body": {
                    "description": "*\n\t * Body",
                    "type": "object",
                    "properties": {
                        "entity": {
                            "type": "string"
                        },
                        "object_id": {
                            "type": "string"
                        },
                        "relation": {
                            "type": "string"
                        },
                        "userset_entity": {
                            "type": "string"
                        },
                        "userset_object_id": {
                            "type": "string"
                        },
                        "userset_relation": {
                            "type": "string"
                        }
                    }
                },
                "pathParams": {
                    "description": "*\n\t * PathParams",
                    "type": "object"
                },
                "queryParams": {
                    "description": "*\n\t * QueryParams",
                    "type": "object"
                }
            }
        },
        "relationship.Write": {
            "type": "object",
            "properties": {
                "body": {
                    "description": "*\n\t * Body",
                    "type": "object",
                    "properties": {
                        "entity": {
                            "type": "string"
                        },
                        "object_id": {
                            "type": "string"
                        },
                        "relation": {
                            "type": "string"
                        },
                        "userset_entity": {
                            "type": "string"
                        },
                        "userset_object_id": {
                            "type": "string"
                        },
                        "userset_relation": {
                            "type": "string"
                        }
                    }
                },
                "pathParams": {
                    "description": "*\n\t * PathParams",
                    "type": "object"
                },
                "queryParams": {
                    "description": "*\n\t * QueryParams",
                    "type": "object"
                }
            }
        },
        "responses.Check": {
            "type": "object",
            "properties": {
                "can": {
                    "type": "boolean"
                },
                "decisions": {},
                "remaining_depth": {
                    "type": "integer"
                }
            }
        },
        "responses.HTTPErrorResponse": {
            "type": "object",
            "properties": {
                "errors": {}
            }
        },
        "responses.Message": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/v1",
	Schemes:     []string{},
	Title:       "Permify API",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}
