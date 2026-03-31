---
name: build-makedo
description: Build makedo binary to the project root under the bin folder
---

# Build `makedo` Binary

This skill ensures the `makedo` project is compiled correctly according to the repository's constraints.

## Build Rules
1. **Working Directory:** The build process must always be executed from within the `src` directory.
2. **Dependencies:** Always ensure Go modules are tidied (`go mod tidy`) before compiling.
3. **Version Injection:** The `-ldflags` argument is strictly required to inject the version string from the `VERSION` file located in the project root.
4. **Output Location:** The final compiled binary **must** be placed in the directory root, specifically under the `bin` folder. Because the build command is executed from within the `src` folder, the output flag must be explicitly routed up one level using `-o ../bin/makedo`.
