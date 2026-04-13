# Example: Clean Function Design

## Principles Demonstrated

- Clear naming reveals intent
- Early returns reduce nesting
- Errors wrapped with context
- Resources cleaned up with defer
- Small, focused scope

## Before (problematic)

```
func process(d []byte, t string, f bool) (interface{}, error) {
    if d != nil {
        if t == "json" {
            var result map[string]interface{}
            err := json.Unmarshal(d, &result)
            if err == nil {
                if f {
                    // validate
                    if _, ok := result["id"]; !ok {
                        return nil, errors.New("error")
                    }
                }
                return result, nil
            } else {
                return nil, err
            }
        }
    }
    return nil, errors.New("bad input")
}
```

## After (clean)

```
func parseAndValidateJSON(rawPayload []byte, requireID bool) (map[string]any, error) {
    if len(rawPayload) == 0 {
        return nil, errors.New("empty payload")
    }

    var parsed map[string]any
    if err := json.Unmarshal(rawPayload, &parsed); err != nil {
        return nil, fmt.Errorf("invalid JSON payload: %w", err)
    }

    if requireID {
        if _, exists := parsed["id"]; !exists {
            return nil, errors.New("missing required field: id")
        }
    }

    return parsed, nil
}
```

## What Changed
- Function name describes exactly what it does
- Parameter names reveal their purpose (`rawPayload` not `d`)
- Boolean parameter is named (`requireID` not `f`)
- Early return for empty input eliminates nesting
- Error messages describe what went wrong specifically
- Original parse error is wrapped with context using `%w`
- Return type is concrete (`map[string]any`) not `interface{}`
- No unnecessary else branches
