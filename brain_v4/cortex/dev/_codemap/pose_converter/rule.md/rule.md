---
author: GeminiCLI
---
# Pose Converter Codemap Rule

## PROVIDES
- 3D Joint data to JSON conversion
- PoseData struct definition

## DEPENDS
- encoding/json (std)
- strings (test only)

## CONSTRAINTS
- Coordinates must be float64
- Error handling is mandatory in callers
- Joint names must not be empty (validation required)

