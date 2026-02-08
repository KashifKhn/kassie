# Configuration Schema

See the [Configuration Guide](/guide/configuration) for complete documentation.

This page provides a technical schema reference.

## JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "profiles"],
  "properties": {
    "version": {
      "type": "string",
      "const": "1.0"
    },
    "profiles": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/profile"
      }
    },
    "defaults": {
      "$ref": "#/definitions/defaults"
    },
    "clients": {
      "$ref": "#/definitions/clients"
    }
  }
}
```

See the full configuration guide for complete details: [Configuration Guide](/guide/configuration)
