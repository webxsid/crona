# Daily Report Template Variables

The daily report templates use Handlebars-style variables.

Supported blocks in v1:
- `{{variable}}`
- `{{#if variable}}...{{/if}}`
- `{{#each items}}...{{/each}}`

Top-level variables:
- `{{date}}`
- `{{generatedAt}}`
- `{{summary.*}}`
- `{{dayHealth}}`
- `{{#each highlights}} ... {{/each}}`
- `{{#each risks}} ... {{/each}}`

Nested groups:
- `{{#each repos}} -> {{#each streams}} -> issues`
- `{{#each habitRepos}} -> {{#each streams}} -> habits`

Optional objects:
- `{{checkIn.*}}`
- `{{metrics.*}}`

Flat collections:
- `{{#each issues}} ... {{/each}}`
- `{{#each sessions}} ... {{/each}}`
- `{{#each habits}} ... {{/each}}`

Time by tag (empty when no worked issue was tagged; a session counts toward every tag on its issue):
- `{{#each tagTime}} -> {{name}}, {{workedTime}}, {{workedSeconds}}, {{percent}}`
