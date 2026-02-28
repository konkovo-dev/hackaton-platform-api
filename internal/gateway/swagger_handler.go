package gateway

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"
)

//go:embed swagger
var swaggerFiles embed.FS

// serviceSwaggerFiles are the proto-generated swagger files to merge into one spec.
var serviceSwaggerFiles = []string{
	"swagger/auth/v1/auth_service.swagger.json",
	"swagger/hackathon/v1/hackathon_service.swagger.json",
	"swagger/identity/v1/users_service.swagger.json",
	"swagger/identity/v1/me_service.swagger.json",
	"swagger/identity/v1/skills_service.swagger.json",
	"swagger/participationandroles/v1/staff_service.swagger.json",
	"swagger/participationandroles/v1/participation_service.swagger.json",
	"swagger/team/v1/teams_service.swagger.json",
	"swagger/team/v1/team_members_service.swagger.json",
	"swagger/team/v1/team_inbox_service.swagger.json",
	"swagger/mentors/v1/mentors_service.swagger.json",
}

var (
	mergedSwaggerOnce sync.Once
	mergedSwaggerJSON []byte
)

// mergedSwagger reads all service swagger files and merges their paths and
// definitions into a single Swagger 2.0 document.
func mergedSwagger() []byte {
	mergedSwaggerOnce.Do(func() {
		merged := map[string]interface{}{
			"swagger": "2.0",
			"info": map[string]interface{}{
				"title":   "Hackathon Platform API",
				"version": "1.0",
			},
			"consumes": []string{"application/json"},
			"produces": []string{"application/json"},
			"securityDefinitions": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":        "apiKey",
					"in":          "header",
					"name":        "Authorization",
					"description": "JWT Bearer token. Example: \"Bearer <token>\"",
				},
			},
			"security": []interface{}{
				map[string]interface{}{"BearerAuth": []interface{}{}},
			},
			"paths":       map[string]interface{}{},
			"definitions": map[string]interface{}{},
		}

		paths := merged["paths"].(map[string]interface{})
		defs := merged["definitions"].(map[string]interface{})

		for _, filePath := range serviceSwaggerFiles {
			data, err := fs.ReadFile(swaggerFiles, filePath)
			if err != nil {
				log.Printf("swagger merge: skipping %s: %v", filePath, err)
				continue
			}

			var spec map[string]interface{}
			if err := json.Unmarshal(data, &spec); err != nil {
				log.Printf("swagger merge: failed to parse %s: %v", filePath, err)
				continue
			}

			if p, ok := spec["paths"].(map[string]interface{}); ok {
				for k, v := range p {
					paths[k] = v
				}
			}
			if d, ok := spec["definitions"].(map[string]interface{}); ok {
				for k, v := range d {
					defs[k] = v
				}
			}
		}

		out, err := json.MarshalIndent(merged, "", "  ")
		if err != nil {
			log.Printf("swagger merge: failed to marshal: %v", err)
			out = []byte(`{"swagger":"2.0","info":{"title":"Hackathon Platform API","version":"1.0"},"paths":{}}`)
		}
		mergedSwaggerJSON = out
	})

	return mergedSwaggerJSON
}

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Hackathon Platform API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
<script>
  window.onload = function() {
    SwaggerUIBundle({
      url: "/swagger/api.json",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [SwaggerUIBundle.plugins.DownloadUrl],
      layout: "StandaloneLayout",
      tryItOutEnabled: true
    });
  };
</script>
</body>
</html>`

func newSwaggerHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/swagger/api.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(mergedSwagger())
	})

	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerUIHTML))
	})

	return mux
}
