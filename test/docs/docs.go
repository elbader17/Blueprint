package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "test API",
	Description:      "This is a sample server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

var docTemplate = `{
    "swagger": "2.0",
    "info": {
        "description": "test API",
        "title": "test API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/",
    "paths": {}
}`

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
