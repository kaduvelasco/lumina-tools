# PHP Development — Lumina Standard

Standards and conventions for PHP development in this project.
Covers PSR standards, architecture patterns, and best practices for modern PHP 8.x applications.

---

## Language

| Context | Language |
|---|---|
| Responses to the user | Brazilian Portuguese (pt-BR) |
| Code comments | English |

---

## Project Structure

Standard layout for PHP projects:

```text
src/
  Domain/           # Business rules — no framework dependencies
    Entity/
    Repository/     # Interfaces only
    Event/
    Exception/
  Application/      # Use cases, commands, queries, handlers
  Infrastructure/   # Framework, DB, HTTP, external services
  Presentation/     # Controllers, CLI commands, templates
tests/
  Unit/
  Integration/
  Functional/
config/
public/             # Web root — index.php only, no business logic
composer.json
```

**Key invariants:**
- Domain layer has zero framework dependencies.
- Interfaces are defined in Domain; implementations live in Infrastructure.
- `public/index.php` bootstraps the application — no business logic here.

---

## Typing Rules

`declare(strict_types=1)` is not used by default. Always declare types on parameters, return types, and class properties instead — this enforces intent without requiring strict mode globally.

```php
class UserService
{
    public function __construct(
        private readonly UserRepositoryInterface $repository,
        private readonly LoggerInterface $logger,
    ) {}

    public function findById(int $id): ?User
    {
        return $this->repository->find($id);
    }
}
```

Never leave parameters or return types untyped when the type can be declared:

```php
// Bad
function process($data) { ... }

// Good
function process(array $data): ProcessResult { ... }
```

Use union types when a value can legitimately be more than one type:

```php
function format(int|float $value): string { ... }
```

---

## PHP 8.x Features

Use modern PHP features — they improve readability and safety.

### match expression

```php
// Bad: verbose switch with implicit type coercion and fallthrough risk
switch ($status) {
    case 'active':   $label = 'Active';   break;
    case 'inactive': $label = 'Inactive'; break;
    default:         $label = 'Unknown';
}

// Good: match is an expression, strict comparison, no fallthrough
$label = match($status) {
    'active'   => 'Active',
    'inactive' => 'Inactive',
    default    => 'Unknown',
};
```

### Nullsafe operator

```php
// Bad
$city = null;
if ($user !== null && $user->getAddress() !== null) {
    $city = $user->getAddress()->getCity();
}

// Good
$city = $user?->getAddress()?->getCity();
```

### Named arguments

```php
// Improves readability for functions with many parameters
$result = array_slice(array: $items, offset: 2, length: 5, preserve_keys: true);
```

### Enums

```php
enum Status: string
{
    case Active   = 'active';
    case Inactive = 'inactive';
    case Pending  = 'pending';

    public function label(): string
    {
        return match($this) {
            Status::Active   => 'Active',
            Status::Inactive => 'Inactive',
            Status::Pending  => 'Pending',
        };
    }
}
```

### Readonly properties

```php
class Money
{
    public function __construct(
        public readonly int $amount,
        public readonly string $currency,
    ) {}
}
```

### First class callable syntax

```php
$fn  = strlen(...);           // instead of Closure::fromCallable('strlen')
$arr = array_map($fn, $items);
```

---

## Error Handling

### Custom exception hierarchy

Define a base exception per domain and derive specific types from it.

```php
// Domain base
class DomainException extends \RuntimeException {}

// Specific exceptions
class UserNotFoundException extends DomainException
{
    public function __construct(int $id)
    {
        parent::__construct("User with ID {$id} not found.");
    }
}

class InsufficientBalanceException extends DomainException
{
    public function __construct(int $available, int $required)
    {
        parent::__construct("Insufficient balance: available {$available}, required {$required}.");
    }
}
```

### Catching specific types

```php
try {
    $user = $this->repository->findOrFail($id);
} catch (UserNotFoundException $e) {
    $this->logger->warning($e->getMessage(), ['id' => $id]);
    return null;
} catch (DomainException $e) {
    $this->logger->error($e->getMessage());
    throw $e;
}
```

Never catch `\Throwable` or `\Exception` without re-throwing or logging a specific reason — it silences unexpected failures.

---

## Security

### Database — always use prepared statements

```php
// Bad — SQL injection
$db->query("SELECT * FROM users WHERE email = '$email'");

// Good — prepared statement via PDO
$stmt = $pdo->prepare('SELECT * FROM users WHERE email = :email');
$stmt->execute(['email' => $email]);
```

### Output — always escape before rendering

```php
// Bad
echo $userInput;

// Good
echo htmlspecialchars($userInput, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8');
```

### Never expose raw exceptions to the user

```php
// Bad — leaks stack trace and internal paths
echo $e->getMessage();

// Good — wrap with a safe message
throw new HttpException(500, 'An unexpected error occurred.');
```

---

## Testing

### PHPUnit — data providers

```php
final class UserServiceTest extends TestCase
{
    /**
     * @dataProvider provideValidEmails
     */
    public function test_accepts_valid_email(string $email): void
    {
        $user = User::create($email, 'Test User');
        self::assertSame($email, $user->email());
    }

    public static function provideValidEmails(): array
    {
        return [
            'standard'     => ['user@example.com'],
            'subdomain'    => ['user@mail.example.com'],
            'plus-address' => ['user+tag@example.com'],
        ];
    }
}
```

### Test doubles — prefer fake implementations over mocks

```php
final class InMemoryUserRepository implements UserRepositoryInterface
{
    private array $users = [];

    public function save(User $user): void
    {
        $this->users[$user->id()->toString()] = $user;
    }

    public function find(UserId $id): ?User
    {
        return $this->users[$id->toString()] ?? null;
    }
}
```

---

## Quality

```bash
composer phpcs    # PHP_CodeSniffer — PSR-1/PSR-12 compliance
composer phpstan  # PHPStan — static analysis
composer phpunit  # PHPUnit — test suite
```

```json
"scripts": {
    "phpcs":   "phpcs --standard=PSR12 src/ tests/",
    "phpstan": "phpstan analyse src/ tests/ --level=8",
    "phpunit": "phpunit --coverage-text"
}
```

- **phpcs** — enforces PSR-1 and PSR-12 formatting; non-negotiable.
- **phpstan** — catches type errors, dead code, and incorrect usage without running the code. Aim for level 8; never go below level 6.
- **phpunit** — run the full test suite before every commit.

---

## What is PSR?

PSR (PHP Standards Recommendations) are specifications published by the PHP Framework Interoperability Group (PHP-FIG). They establish common standards for PHP code to ensure interoperability between frameworks and libraries.

## Accepted PSRs Summary

| PSR | Name | Category | Status |
|-----|------|----------|--------|
| PSR-1 | Basic Coding Standard | Coding Style | Accepted |
| PSR-3 | Logger Interface | Logging | Accepted |
| PSR-4 | Autoloader | Autoloading | Accepted |
| PSR-6 | Caching Interface | Caching | Accepted |
| PSR-7 | HTTP Message Interface | HTTP | Accepted |
| PSR-11 | Container Interface | DI Container | Accepted |
| PSR-12 | Extended Coding Style | Coding Style | Accepted |
| PSR-13 | Hypermedia Links | Hypermedia | Accepted |
| PSR-14 | Event Dispatcher | Events | Accepted |
| PSR-15 | HTTP Handlers | HTTP | Accepted |
| PSR-16 | Simple Cache | Caching | Accepted |
| PSR-17 | HTTP Factories | HTTP | Accepted |
| PSR-18 | HTTP Client | HTTP | Accepted |
| PSR-20 | Clock | Time | Accepted |

## PSR Categories

### Coding Style (PSR-1, PSR-12)

Standards for writing clean, consistent PHP code.

| Aspect | PSR-1 | PSR-12 |
|--------|-------|--------|
| Scope | Basic rules | Extended formatting |
| File encoding | UTF-8 without BOM | Inherits PSR-1 |
| Class names | StudlyCaps | Inherits PSR-1 |
| Method names | camelCase | Inherits PSR-1 |
| Indentation | - | 4 spaces |
| Line length | - | 120 chars soft limit |
| Keywords | - | Lowercase |

### Autoloading (PSR-4)

Standard for autoloading classes from file paths.

```
Namespace Prefix → Base Directory
App\             → src/

FQCN                          → File Path
App\Domain\User\Entity\User   → src/Domain/User/Entity/User.php
```

### HTTP (PSR-7, PSR-15, PSR-17, PSR-18)

Standards for HTTP messages, handlers, factories, and clients.

| PSR | Purpose | Key Interfaces |
|-----|---------|----------------|
| PSR-7 | HTTP Messages | `RequestInterface`, `ResponseInterface`, `StreamInterface` |
| PSR-15 | HTTP Handlers | `MiddlewareInterface`, `RequestHandlerInterface` |
| PSR-17 | HTTP Factories | `RequestFactoryInterface`, `ResponseFactoryInterface` |
| PSR-18 | HTTP Client | `ClientInterface` |

```
PSR-17 (Factory) → PSR-7 (Message) → PSR-15 (Handler) → PSR-7 (Response)
                                          ↓
                                   PSR-18 (Client)
```

### Caching (PSR-6, PSR-16)

Standards for caching implementations.

| Aspect | PSR-6 | PSR-16 |
|--------|-------|--------|
| Complexity | Full-featured | Simple |
| Key interfaces | `CacheItemPoolInterface`, `CacheItemInterface` | `CacheInterface` |
| Deferred saves | Yes | No |
| Use case | Complex caching needs | Simple get/set |

### Logging (PSR-3)

Standard for logging libraries.

```php
interface LoggerInterface {
    public function emergency(string|\Stringable $message, array $context = []): void;
    public function alert(string|\Stringable $message, array $context = []): void;
    public function critical(string|\Stringable $message, array $context = []): void;
    public function error(string|\Stringable $message, array $context = []): void;
    public function warning(string|\Stringable $message, array $context = []): void;
    public function notice(string|\Stringable $message, array $context = []): void;
    public function info(string|\Stringable $message, array $context = []): void;
    public function debug(string|\Stringable $message, array $context = []): void;
    public function log(mixed $level, string|\Stringable $message, array $context = []): void;
}
```

### DI Container (PSR-11)

Standard for dependency injection containers.

```php
interface ContainerInterface {
    public function get(string $id): mixed;
    public function has(string $id): bool;
}
```

### Events (PSR-14)

Standard for event dispatching.

```php
interface EventDispatcherInterface {
    public function dispatch(object $event): object;
}

interface ListenerProviderInterface {
    public function getListenersForEvent(object $event): iterable;
}

interface StoppableEventInterface {
    public function isPropagationStopped(): bool;
}
```

### Hypermedia (PSR-13)

Standard for hypermedia links (HATEOAS).

```php
interface LinkInterface {
    public function getHref(): string;
    public function isTemplated(): bool;
    public function getRels(): array;
    public function getAttributes(): array;
}
```

### Time (PSR-20)

Standard for clock abstraction.

```php
interface ClockInterface {
    public function now(): DateTimeImmutable;
}
```

## When to Use Each PSR

### Decision Matrix

| Need | PSR |
|------|-----|
| Code formatting | PSR-1, PSR-12 |
| Class autoloading | PSR-4 |
| Logging | PSR-3 |
| Simple caching (get/set) | PSR-16 |
| Complex caching (pools, tags) | PSR-6 |
| HTTP requests/responses | PSR-7 |
| HTTP middleware | PSR-15 |
| Creating HTTP objects | PSR-17 |
| HTTP client for external APIs | PSR-18 |
| Dependency injection | PSR-11 |
| Event system | PSR-14 |
| REST API with links | PSR-13 |
| Testing with time | PSR-20 |

### Common Combinations

| Use Case | PSRs |
|----------|------|
| HTTP API | PSR-7 + PSR-15 + PSR-17 |
| HTTP Client | PSR-7 + PSR-17 + PSR-18 |
| Web Application | PSR-1 + PSR-4 + PSR-12 + PSR-3 + PSR-11 |
| CQRS/Event-Driven | PSR-14 + PSR-3 + PSR-11 |
| Microservice | All of the above |

## PHP Package Implementations

### PSR-3: Logger

| Package | Description |
|---------|-------------|
| `monolog/monolog` | De facto standard logger |
| `psr/log` | Interface only |

### PSR-4: Autoloader

| Package | Description |
|---------|-------------|
| Composer | Built-in autoloader |

### PSR-6: Cache

| Package | Description |
|---------|-------------|
| `symfony/cache` | Full-featured cache |
| `cache/filesystem-adapter` | File-based cache |

### PSR-7/PSR-17: HTTP

| Package | Description |
|---------|-------------|
| `guzzlehttp/psr7` | Guzzle implementation |
| `nyholm/psr7` | Lightweight implementation |
| `laminas/laminas-diactoros` | Laminas implementation |

### PSR-11: Container

| Package | Description |
|---------|-------------|
| `php-di/php-di` | Autowiring DI container |
| `league/container` | Flexible container |
| `pimple/pimple` | Simple container |

### PSR-14: Event Dispatcher

| Package | Description |
|---------|-------------|
| `symfony/event-dispatcher` | Symfony implementation |
| `league/event` | League implementation |

### PSR-15: HTTP Handlers

| Package | Description |
|---------|-------------|
| `middlewares/utils` | Middleware utilities |
| `relay/relay` | Simple dispatcher |

### PSR-18: HTTP Client

| Package | Description |
|---------|-------------|
| `guzzlehttp/guzzle` | Full HTTP client |
| `symfony/http-client` | Symfony HTTP client |

### PSR-20: Clock

| Package | Description |
|---------|-------------|
| `psr/clock` | Interface only |
| `symfony/clock` | Symfony implementation |
| `lcobucci/clock` | Simple implementation |

## Composer Requirements

```json
{
    "require": {
        "psr/log": "^3.0",
        "psr/cache": "^3.0",
        "psr/http-message": "^2.0",
        "psr/http-factory": "^1.0",
        "psr/http-client": "^1.0",
        "psr/container": "^2.0",
        "psr/event-dispatcher": "^1.0",
        "psr/link": "^2.0",
        "psr/clock": "^1.0",
        "psr/simple-cache": "^3.0"
    }
}
```

## Integration with DDD

### Layer Mapping

| DDD Layer | Relevant PSRs | Note |
|-----------|---------------|------|
| Domain | PSR-3, PSR-14, PSR-20 | PSR interfaces are pure contracts — acceptable in Domain |
| Application | PSR-3, PSR-11, PSR-14, PSR-20 | Service orchestration layer |
| Infrastructure | PSR-6, PSR-16, PSR-18 | Implementation layer |
| Presentation | PSR-7, PSR-15, PSR-17 | HTTP layer |

> **Important:** PSR packages (`psr/log`, `psr/clock`, `psr/event-dispatcher`) contain **only interfaces** —
> no implementation code. They are PHP community standards equivalent to a standard library.
> Using PSR interfaces in Domain layer is acceptable and common practice.
> What is NOT acceptable in Domain: implementation packages like `monolog/monolog`, `symfony/cache`, `guzzlehttp/guzzle`.

### Example: CQRS Application

```php
<?php

declare(strict_types=1);

namespace App\Application\User\Handler;

use App\Application\User\Command\CreateUserCommand;
use App\Domain\User\Entity\User;
use App\Domain\User\Repository\UserRepositoryInterface;
use Psr\EventDispatcher\EventDispatcherInterface;  // PSR-14
use Psr\Log\LoggerInterface;                       // PSR-3
use Psr\Clock\ClockInterface;                      // PSR-20

final readonly class CreateUserHandler
{
    public function __construct(
        private UserRepositoryInterface $repository,
        private EventDispatcherInterface $eventDispatcher,
        private LoggerInterface $logger,
        private ClockInterface $clock,
    ) {
    }

    public function __invoke(CreateUserCommand $command): void
    {
        $this->logger->info('Creating user', ['email' => $command->email]);

        $user = User::create(
            $command->email,
            $command->name,
            $this->clock->now(),
        );

        $this->repository->save($user);

        foreach ($user->pullEvents() as $event) {
            $this->eventDispatcher->dispatch($event);
        }
    }
}
```

## Compliance Checklist

| PSR | Required For | Check |
|-----|--------------|-------|
| PSR-1 | All PHP projects | `phpcs --standard=PSR1` |
| PSR-12 | All PHP projects | `phpcs --standard=PSR12` |
| PSR-4 | All PHP projects | `composer dump-autoload --strict` |
| PSR-3 | Projects with logging | Implement `LoggerInterface` |
| PSR-6/16 | Projects with caching | Implement `CacheInterface` |
| PSR-7 | HTTP APIs | Use PSR-7 implementations |
| PSR-11 | Projects with DI | Implement `ContainerInterface` |
| PSR-14 | Event-driven projects | Implement `EventDispatcherInterface` |
| PSR-15 | HTTP middleware | Implement `MiddlewareInterface` |
| PSR-17 | HTTP object creation | Use factory interfaces |
| PSR-18 | External API calls | Implement `ClientInterface` |
| PSR-20 | Time-sensitive code | Implement `ClockInterface` |

## See Also

- `php-references/accepted-psrs.md` - Detailed PSR descriptions and interfaces
- `php-references/when-to-use.md` - Decision matrix for PSR selection
- `php-references/compatibility.md` - Inter-PSR relationships and dependency graph
- `php-references/php-fig-process.md` - How PSRs are created and approved
