use serde::{Serialize, Deserialize};
use std::collections::HashMap;
use std::sync::{OnceLock, RwLock};

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct RouteDoc {
    pub method: String,
    pub path: String,
    pub summary: String,
    pub description: String,
    pub tags: Vec<String>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub params: Vec<ParamDoc>,
    #[serde(skip_serializing_if = "Option::is_none", rename = "requestBody", default)]
    pub request_body: Option<RequestBodyDoc>,
    pub responses: HashMap<String, ResponseDoc>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ParamDoc {
    pub name: String,
    pub r#in: String, // "query", "path", "header"
    #[serde(skip_serializing_if = "String::is_empty", default)]
    pub description: String,
    pub required: bool,
    pub r#type: String, // e.g. "string", "integer"
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct RequestBodyDoc {
    pub content: HashMap<String, MediaTypeDoc>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct MediaTypeDoc {
    pub schema: SchemaDoc,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct SchemaDoc {
    pub r#type: String,
    #[serde(skip_serializing_if = "HashMap::is_empty", default)]
    pub properties: HashMap<String, Property>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Property {
    pub r#type: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ResponseDoc {
    pub description: String,
}

pub struct APIRegistry {
    pub title: RwLock<String>,
    pub description: RwLock<String>,
    routes: RwLock<HashMap<String, RouteDoc>>,
}

impl APIRegistry {
    pub fn global() -> &'static Self {
        static INSTANCE: OnceLock<APIRegistry> = OnceLock::new();
        INSTANCE.get_or_init(|| Self {
            title: RwLock::new("Zeno API".to_string()),
            description: RwLock::new("Auto-generated API Documentation".to_string()),
            routes: RwLock::new(HashMap::new()),
        })
    }

    pub fn register(&self, method: &str, path: &str, doc: RouteDoc) {
        let mut routes = self.routes.write().unwrap();
        let key = format!("{}:{}", method, path);
        routes.insert(key, doc);
    }

    pub fn get_routes(&self) -> Vec<RouteDoc> {
        let routes = self.routes.read().unwrap();
        routes.values().cloned().collect()
    }

    pub fn to_json(&self) -> Result<String, serde_json::Error> {
        let spec = self.generate_openapi();
        serde_json::to_string_pretty(&spec)
    }

    pub fn generate_openapi(&self) -> serde_json::Value {
        let title = self.title.read().unwrap().clone();
        let description = self.description.read().unwrap().clone();
        let routes = self.routes.read().unwrap();

        let mut paths_obj = serde_json::Map::new();

        for route in routes.values() {
            // Find or create path entry in paths
            let path_entry = paths_obj
                .entry(route.path.clone())
                .or_insert_with(|| serde_json::Value::Object(serde_json::Map::new()));

            if let serde_json::Value::Object(methods_map) = path_entry {
                let method_lower = route.method.to_lowercase();

                let mut operation = serde_json::Map::new();
                operation.insert("summary".to_string(), serde_json::Value::String(route.summary.clone()));
                operation.insert("description".to_string(), serde_json::Value::String(route.description.clone()));
                operation.insert("tags".to_string(), serde_json::to_value(&route.tags).unwrap_or(serde_json::Value::Array(Vec::new())));

                if !route.params.is_empty() {
                    operation.insert("parameters".to_string(), serde_json::to_value(&route.params).unwrap());
                }

                if let Some(ref body) = route.request_body {
                    operation.insert("requestBody".to_string(), serde_json::to_value(body).unwrap());
                }

                operation.insert("responses".to_string(), serde_json::to_value(&route.responses).unwrap_or(serde_json::Value::Object(serde_json::Map::new())));

                methods_map.insert(method_lower, serde_json::Value::Object(operation));
            }
        }

        serde_json::json!({
            "openapi": "3.0.0",
            "info": {
                "title": title,
                "version": "1.0.0",
                "description": description
            },
            "paths": paths_obj
        })
    }
}

pub fn swagger_ui_html(swagger_json_url: &str) -> String {
    format!(
        r#"<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
<script>
window.onload = function() {{
  window.ui = SwaggerUIBundle({{
    url: "{}",
    dom_id: '#swagger-ui',
  }});
}};
</script>
</body>
</html>"#,
        swagger_json_url
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_apidoc_registration_and_openapi_generation() {
        let registry = APIRegistry::global();
        
        let mut responses = HashMap::new();
        responses.insert("200".to_string(), ResponseDoc {
            description: "Success".to_string(),
        });

        registry.register("GET", "/api/v1/ping", RouteDoc {
            method: "GET".to_string(),
            path: "/api/v1/ping".to_string(),
            summary: "Ping endpoint".to_string(),
            description: "Returns 200 pong".to_string(),
            tags: vec!["System".to_string()],
            params: vec![ParamDoc {
                name: "verbosity".to_string(),
                r#in: "query".to_string(),
                description: "Debug verbosity".to_string(),
                required: false,
                r#type: "string".to_string(),
            }],
            request_body: None,
            responses,
        });

        let openapi_json = registry.to_json().unwrap();
        println!("Generated OpenAPI Spec:\n{}", openapi_json);

        assert!(openapi_json.contains("\"openapi\": \"3.0.0\""));
        assert!(openapi_json.contains("\"title\": \"Zeno API\""));
        assert!(openapi_json.contains("\"/api/v1/ping\""));
        assert!(openapi_json.contains("\"get\""));
        assert!(openapi_json.contains("\"verbosity\""));
        assert!(openapi_json.contains("\"System\""));
    }

    #[test]
    fn test_swagger_ui_html() {
        let html = swagger_ui_html("/swagger.json");
        assert!(html.contains("url: \"/swagger.json\""));
        assert!(html.contains("https://unpkg.com/swagger-ui-dist"));
    }
}
