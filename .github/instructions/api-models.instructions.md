---
applyTo: "volumeleaders/**/*.go"
---

# VolumeLeaders API model review instructions

- Structs should match captured VolumeLeaders JSON field names and avoid silently dropping important response fields.
- Public request parameter structs should model captured form, query, path, and DataTables parameters accurately.
- Changes to request payload structs need tests that prove the emitted JSON, form body, or query shape.
- Avoid speculative fields unless captured browser traffic, fixtures, or observed VolumeLeaders behavior show they are needed.
- Keep raw DataTables field names and VolumeLeaders form keys behind typed Go APIs where possible.
