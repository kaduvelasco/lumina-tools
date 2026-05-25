# Moodle Plugin Development

Standards and conventions for developing Moodle plugins in this project.
Covers file structure, API usage, architecture patterns, and common mistakes.

---

## Development Environment

- **Moodle Version:** {{MOODLE_VERSION}}
- **Installation Path:** {{MOODLE_PATH}}
- **MCP Server:** `lumina-mdle-dev` is configured and indices have been generated.

---

## lumina-mdle-dev — Usage Guide

Always load the plugin context before starting work:

```text
Load context for plugin local_myplugin.
```

### Available Tools

- `get_plugin_info` — loads the complete plugin context into the current session.
- `search_api` — searches for Moodle core API functions.
- `generate_plugin_context` — generates `PLUGIN_*.md` documentation files for a plugin.
- `update_indexes` — regenerates global indices after installing new plugins.
- `doctor` — runs health checks on the `moodle-dev-mcp` environment.

### Recommended Workflow

1. **Initialize:** Load context using `get_plugin_info` before working on a plugin.
2. **Research:** Use `search_api` before suggesting or implementing core functions.
3. **Document:** Run `generate_plugin_context` after significant code changes.
4. **Sync:** Execute `update_indexes` whenever new plugins are added to the environment.

---

## Target Moodle Version

This project targets **Moodle {{MOODLE_VERSION}}+** (`requires = {{MOODLE_FULLVERSION}}`).

- Use only APIs compatible with Moodle {{MOODLE_VERSION}} or later.
- **Strictly avoid** functions deprecated in previous versions.
- **Note:** Hook API is **NOT** available in this version — use `lib.php` callbacks instead.

---

## Required Files

Every plugin must include these files at a minimum:

```text
version.php
lang/en/[component].php
db/install.xml
db/access.php
```

**Optional (recommended):** `settings.php`, `lib.php`, `index.php`, `classes/`, `templates/`, `amd/src/`, `tests/`, `tests/behat/`

---

## File Standards

### version.php

```php
defined('MOODLE_INTERNAL') || die();

$plugin->component = 'local_example';
$plugin->version   = 2025010100;
$plugin->requires  = 2022112800; // Matches {{MOODLE_FULLVERSION}}
$plugin->maturity  = MATURITY_STABLE;
$plugin->release   = '1.0';
```

### Language File

**Path:** `lang/en/local_example.php`

```php
defined('MOODLE_INTERNAL') || die();

$string['pluginname'] = 'Example Plugin';
```

All user-facing strings must be defined here. Never hardcode strings in PHP or Mustache templates.

### db/install.xml

- Tables must use the Moodle prefix (handled by XMLDB).
- Every table **must** have a primary key.
- Define indexes for columns used in `WHERE` or `JOIN` clauses.
- Supported types: `INT`, `CHAR`, `TEXT`, `NUMBER`, `FLOAT`.

### db/access.php

All capabilities must be declared in this file.

```php
defined('MOODLE_INTERNAL') || die();

$capabilities = [
    'local/example:view' => [
        'riskbitmask'  => RISK_PERSONAL,
        'captype'      => 'read',
        'contextlevel' => CONTEXT_SYSTEM,
        'archetypes'   => [
            'user'    => CAP_ALLOW,
            'manager' => CAP_ALLOW,
        ],
    ],
];
```

---

## PHP Namespaces

All classes must use the plugin namespace. File paths must match namespaces under `classes/`.

```php
namespace local_example\service;
namespace local_example\repository;
namespace local_example\output;
namespace local_example\external;
namespace local_example\task;
namespace local_example\event;
```

---

## Recommended Plugin Architecture

```text
classes/
  service/        → Business logic
  repository/     → Database access
  output/         → Rendering logic (renderers and renderables)
  external/       → Web service endpoints
  task/           → Scheduled and ad-hoc tasks
  event/          → Event classes
db/
  events.php      → Event observers
templates/        → Mustache templates
amd/src/          → AMD JavaScript modules
tests/            → PHPUnit tests
tests/behat/      → Behat acceptance tests
```

### Entry Point Pattern (index.php)

```php
require_once(__DIR__ . '/../../config.php');

require_login();
require_capability('local/example:view', context_system::instance());

$PAGE->set_url(new moodle_url('/local/example/index.php'));
$PAGE->set_context(context_system::instance());
$PAGE->set_title(get_string('pluginname', 'local_example'));

$service = new \local_example\service\example_service();
$data    = $service->get_data();

$output  = $PAGE->get_renderer('local_example');
echo $output->header();
echo $output->render_from_template('local_example/main', $data);
echo $output->footer();
```

---

## Common Mistakes

### Quick Reference

| Action | Wrong | Correct |
| :--- | :--- | :--- |
| **HTML Output** | `echo "<div>Hello</div>";` | `$output->render_from_template(...)` |
| **Request Vars** | `$id = $_POST['id'];` | `$id = required_param('id', PARAM_INT);` |
| **URLs** | `echo '/local/plugin/index.php';` | `new moodle_url('/local/plugin/index.php')` |
| **DB Queries** | `mysqli_query(...);` | `$DB->get_records('table', [...]);` |
| **Security** | Skipping checks | `require_login(); require_capability();` |

### Database Access

```php
// Wrong
mysqli_query(...);

// Correct — always use the $DB API
$records = $DB->get_records('example_table', ['userid' => $userid]);
$DB->insert_record('example_table', $dataobject);
$DB->update_record('example_table', $dataobject);
$DB->delete_records('example_table', ['id' => $id]);
```

### Request Parameters

```php
// Wrong
$id = $_POST['id'];

// Correct
$id = required_param('id', PARAM_INT);
$id = optional_param('id', 0, PARAM_INT);
```

### Language Strings

```php
// Wrong
echo "Save";

// Correct
echo get_string('save', 'local_example');
```

### JavaScript — Use AMD Modules

Never use inline `<script>` tags.

```javascript
// amd/src/example.js
define(["jquery"], function ($) {
    return {
        init: function () { /* ... */ },
    };
});
```

```mustache
{{! Load AMD module from template }}
{{#js}}
require(['local_example/example'], function(mod) {
    mod.init();
});
{{/js}}
```

---

## Global Context Guidelines

- Always use the global `$DB` object for all database operations.
- SQL table names in raw queries must use `{bracket_format}`.
- Follow **Moodle Coding Style** (based on PSR-12).
- **Exclude from indexing:** `.git`, `node_modules`, `vendor`, `.grunt`, `moodledata`, `cache`.
