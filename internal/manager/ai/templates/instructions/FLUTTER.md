# AI Rules for Flutter (Strict Guidelines)

You are an expert Flutter/Dart agent. You must strictly follow these structural, architectural, and design principles.

## Architectural Rules (MVVM & Clean Code)
- **State Management**: Use ONLY Flutter's built-in solutions (`ValueNotifier` with `ValueListenableBuilder`, `ChangeNotifier` with `ListenableBuilder`, or `Streams`). DO NOT use third-party packages (Bloc, Riverpod) unless explicitly requested.
- **Dependency Injection**: Use manual constructor dependency injection to keep layer dependencies explicit.
- **Routing**: Use `go_router` for declarative navigation and auth redirects.
- **Layers**: Organize features into Presentation (widgets/screens), Domain (business logic), Data (models/repositories), and Core (shared utils).
- **Code Generation**: Use `build_runner` with `json_serializable` (using `fieldRename: FieldRename.snake` for JSON models).

## Coding Standards & Quality
- **Functions**: Keep functions short, single-purpose, and strictly under 20 lines.
- **Styling**: Line length must be 80 characters or fewer. Use `PascalCase` for classes, `camelCase` for members/variables, and `snake_case` for files.
- **Logging**: NEVER use `print()`. Use `developer.log()` from `dart:developer`.
- **Widgets**: Prefer small private Widget classes over helper methods returning a `Widget`. Break down large `build()` methods. Use `const` constructors everywhere possible.
- **Lists**: Always use `ListView.builder` or `SliverList` for long lists.
- **Concurrence**: Use `compute()` to offload expensive operations (like JSON parsing) to separate isolates.

## Visual Design & UI
- **Material 3**: Use `ColorScheme.fromSeed` to centralize light and dark themes. Use `ThemeExtension` for custom tokens and `WidgetStateProperty.resolveWith` for interactive states.
- **Premium Look**: Apply multi-layered drop shadows on cards for depth, subtle noise texture to main backgrounds, and an elegant color "glow" effect on interactive elements (buttons, sliders).
- **Typography**: Adhere to WCAG 2.1 contrast standards (4.5:1 minimum for normal text). Set line height between 1.4x and 1.6x. Use `google_fonts`.

## Testing & Documentation
- **Testing**: Follow Arrange-Act-Assert. Use `package:test` and `package:flutter_test`. Prefer fakes/stubs over mocks. Use `package:checks` for assertions.
- **Documentation**: All public APIs must have `dartdoc` (`///`) comments starting with a single-sentence summary.
