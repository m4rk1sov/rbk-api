package handler

import "net/http"

// DocsHandler serves a minimal Swagger UI page pointing to /swagger.yaml.
func DocsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <title>RBK Fitness API — Swagger UI</title>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
  <style>
    body { margin: 0; background: #0b1020; }
    .topbar { display: none; }
    #swagger-ui { max-width: 1200px; margin: 0 auto; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/swagger.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}

// RedocHandler serves a minimal Redoc page pointing to /swagger.yaml.
func RedocHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <title>RBK Fitness API — Redoc</title>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
  <redoc spec-url="/swagger.yaml"></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js" crossorigin></script>
</body>
</html>`
	_, _ = w.Write([]byte(html))
}

// SwaggerYAMLHandler serves the swagger.yaml file from the project root.
func SwaggerYAMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	http.ServeFile(w, r, "./swagger.yaml")
}
