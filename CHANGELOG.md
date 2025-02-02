# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.1] - 2025-2-1

### Fixed

- history path
- tac path on mac

## [0.5.0] - 2025-2-1

### Changed

- use ctrl- for all fzf bindings
- cache_all_repos.py renamed to cache.py
  - instead of `map ctrl+shift+g kitten meow/cache_all_repos.py ...`
  - use `map ctrl+shift+g kitten meow/cache.py ...`
- kill_old_projects.py renamed to kill.py
  - instead of `map ctrl+shift+x kitten meow/kill_old_projects.py`
  - use `map ctrl+shift+x kitten meow/kill.py`
- kill now defaults to any project, ctrl-o to kill old projects

## [0.4.0] - 2024-4-20

### Changed

- combined python scripts for load & new into single file
  - instead of
    - `ctrl+p kitten meow/load_project.py ...`
    - `ctrl+shift+n kitten meow/load_project.py ...`
  - use
    - `ctrl+p kitten meow/projects.py load ...`
    - `ctrl+shift+n kitten meow/projects.py new ...`

### Added

- load & new projects support multiple selections
- fix 2nd window path on mac

## [0.3.0] - 2023-11-11

### Changed

- loads env EDITOR now, not always vim

### Added

- keybinds listed in fzf header

## [0.2.0] - 2023-7-19

### Added

- new project command. load projects outside your user/org using ctrl+shift+n

## [0.1.1] - 2023-3-31

### Fixed

- can run cache_all_repos.py directly for testing

## [0.1.0] - 2023-3-11

### Added

- repo caching
- user repo support, org is now optional, and both can be repeated

## [0.0.4] - 2023-3-6

### Changed

- default project list now combines tabs and local projects

## [0.0.3] - 2023-2-1

### Fixed

- switching or loading a project now matches exact project name instead of a substring

## [0.0.2] - 2023-01-18

### Added

- Added project history tracking. Switching project stores last view in history file. Ctrl-x lists
  and closes projects that are old.

## [0.0.1] - 2023-01-18

### Added - 2022-11-21

- Initial release

[unreleased]: https://github.com/taylorzr/meow/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/taylorzr/meow/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/taylorzr/meow/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/taylorzr/meow/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/taylorzr/meow/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/taylorzr/meow/compare/v0.0.4...v0.1.0
[0.0.4]: https://github.com/taylorzr/meow/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/taylorzr/meow/releases/tag/v0.0.2..v0.0.3
[0.0.2]: https://github.com/taylorzr/meow/releases/tag/v0.0.1..v0.0.2
[0.0.1]: https://github.com/taylorzr/meow/releases/tag/v0.0.1
