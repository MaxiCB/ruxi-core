# ruxi-core

---
[![codecov](https://codecov.io/gh/MaxiCB/ruxi-core/branch/master/graph/badge.svg?token=ZAA0KGZWGK)](https://codecov.io/gh/MaxiCB/ruxi-core)
![build](https://github.com/maxicb/ruxi-core/actions/workflows/go.yml/badge.svg)
[![semantic-release: angular](https://img.shields.io/badge/semantic--release-angular-e10079?logo=semantic-release)](https://github.com/semantic-release/semantic-release)

---

### Overview
---
Ruxi Core is a general purpose helper go module for use in varying services. This module includes helpers for logging, a base gin router (with health-check), and gorm DB connection. This module is used throughout the various ruxi services.

### Deploy Steps
---
Any changes after first stable release should follow the workflow of a feature branch being merged into the master branch.

Deployments are triggered on a change being detected on `master`. This is handled by [Semantic Release](https://github.com/semantic-release/semantic-release). semantic-release expects the [Angular Commit Message Format](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#-commit-message-format)
