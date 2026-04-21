---
name: build-makedo
description: Build makedo binary to the project root under the bin folder
---

# Build `makedo` Binary

**Where to run:**
From within the `src` directory.

**What command to run:**
`go mod tidy && go build -ldflags "-X main.versionString=$(cat ../VERSION)" -o ../bin/makedo`