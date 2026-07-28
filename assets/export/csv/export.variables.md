# CSV Export Spec

The CSV export uses a JSON spec instead of Handlebars.

Supported key:
- `columns`: ordered objects with a `header` label and `field` name

Legacy specs with separate `headers` and string `columns` arrays remain supported.

Common row fields:
- `sessionId`
- `issueId`
- `issueTitle`
- `repo`
- `stream`
- `startTime`
- `endTime`
- `durationSeconds`
- `summary`
