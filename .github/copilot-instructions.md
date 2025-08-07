# AzTUI - Azure Terminal User Interface

AzTUI is a Go-based Terminal User Interface (TUI) application for managing Azure resources interactively from the command line. It uses the tview library for UI and integrates with Azure CLI for authentication.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Prerequisites and Setup

### Required Dependencies
- **Go 1.19 or higher** (tested with Go 1.24.5)
- **Azure CLI** (`az` command) - must be installed and authenticated
- **Linux/Unix environment** (Windows may work but not tested)

### Authentication Setup
- Run `az login` before using aztui
- Ensure you have appropriate Azure subscription access
- The application will use your current Azure CLI authentication context

## Working Effectively

### Bootstrap and Build
- **NEVER CANCEL builds** - Initial builds take ~36 seconds, subsequent builds take ~0.5 seconds
- Always use timeouts of 60+ minutes for first builds, 5+ minutes for subsequent builds

```bash
# Build the application
make all                    # NEVER CANCEL: Takes ~36 seconds on first build, ~0.5s subsequently
                           # Set timeout to 60+ minutes for first build, 5+ minutes for subsequent

# Clean build artifacts
make clean                 # Takes <1 second

# Format Go code (always run before committing)
make format               # Takes ~0.2 seconds, runs gofmt -s -w ./src/
```

### Running the Application
```bash
# Run using built binary (recommended)
AZTUI_CONFIG_PATH=conf/default.yaml bin/aztui

# Run using make (alternative)
make run

# Note: Application requires Azure CLI authentication
# If not authenticated, run: az login
```

### Development Workflow
- **ALWAYS run `make format` before committing** - formats all Go code
- **Build and test changes**: `make clean && make all`
- **No unit tests exist** - manual validation required
- Check Azure CLI authentication if encountering auth errors

## Validation

### Manual Testing Scenarios
After making changes, **ALWAYS** test these scenarios:

1. **Basic Startup Test**:
   ```bash
   # Ensure app starts without crashing
   timeout 10 env AZTUI_CONFIG_PATH=conf/default.yaml bin/aztui
   # Should start TUI interface, timeout is expected
   ```

2. **Build Validation**:
   ```bash
   # Clean build to ensure no dependency issues
   make clean && make all
   # Should complete without errors
   ```

3. **Authentication Check**:
   ```bash
   # Verify Azure CLI is authenticated
   az account show
   # Should show current subscription details
   ```

### Required Validation Steps
- **Always run `make format`** before any code changes
- **Always test basic startup** after changes to main.go or core packages
- **Always verify clean build** after dependency changes
- **Cannot perform full UI testing** in automated environments (TUI application)

## Build Times and Timeouts

### Critical Timing Information
- **First build**: ~36 seconds (downloads dependencies) - **NEVER CANCEL, set 60+ minute timeout**
- **Subsequent builds**: ~0.5 seconds (uses cached dependencies) - **set 5+ minute timeout**
- **Format command**: ~0.2 seconds
- **Clean command**: <1 second

### Commands with Timeout Requirements
```bash
# CRITICAL: Use appropriate timeouts
make all     # TIMEOUT: 60+ minutes (first build), 5+ minutes (subsequent)
make run     # May run indefinitely (TUI app), use timeout for testing
make format  # TIMEOUT: 1 minute (very fast)
make clean   # TIMEOUT: 1 minute (very fast)
```

## Repository Structure

### Key Directories and Files
```
├── src/                    # Go source code
│   ├── cmd/main.go        # Application entry point
│   ├── cmd/test/          # Test CLI utility (not unit tests)
│   ├── pkg/               # Application packages
│   │   ├── azcli/         # Azure CLI integration
│   │   ├── config/        # Configuration handling
│   │   ├── consoles/      # Serial console functionality
│   │   ├── logger/        # Logging utilities
│   │   ├── resourceviews/ # TUI views and components
│   │   └── utils/         # Utility functions
│   ├── go.mod             # Go module definition
│   └── go.sum             # Go module checksums
├── conf/default.yaml      # Default configuration file
├── bin/                   # Build output directory
├── Makefile              # Build system
└── .github/workflows/    # CI/CD pipeline
```

## Common Tasks

The following are outputs from frequently run commands. Reference them instead of viewing, searching, or running bash commands to save time.

### Repository Root Listing
```bash
$ ls -la
total 1236
drwxr-xr-x 7 runner docker    4096 Aug  7 02:40 .
drwxr-xr-x 3 runner docker    4096 Aug  7 02:39 ..
drwxr-xr-x 7 runner docker    4096 Aug  7 02:40 .git
drwxr-xr-x 3 runner docker    4096 Aug  7 02:40 .github
-rw-r--r-- 1 runner docker      29 Aug  7 02:40 .gitignore
-rw-r--r-- 1 runner docker    3368 Aug  7 02:40 Makefile
-rw-r--r-- 1 runner docker    2232 Aug  7 02:40 README.md
drwxr-xr-x 2 runner docker    4096 Aug  7 02:40 conf
-rw-r--r-- 1 runner docker 1220815 Aug  7 02:40 demo.gif
drwxr-xr-x 2 runner docker    4096 Aug  7 02:40 images
drwxr-xr-x 4 runner docker    4096 Aug  7 02:40 src
```

### Source Directory Structure
```bash
$ ls -la src/
cmd  go.mod  go.sum  pkg
```

### Available Make Targets
```bash
# Primary commands
make all        # Build the binary to bin/aztui
make clean      # Remove bin/ directory
make run        # Run application with default config
make format     # Format Go code using gofmt
make rpm        # Build RPM package (requires Docker)
make srpm       # Build source RPM
```

### Go Module Information
```bash
$ head -5 src/go.mod
module github.com/brendank310/aztui

go 1.19

require (
```

## Development Guidelines

### Code Style
- **Always run `make format`** before committing
- Follow standard Go conventions
- Use meaningful variable and function names
- Keep functions focused and small

### Testing
- **No unit tests exist** in this codebase
- Manual testing is required for all changes
- Test basic startup after any changes
- Validate Azure integration when modifying azcli package

### CI/CD
- GitHub Actions workflow builds on push to main/hackathon2024 branches
- Build creates binary artifact and RPM package
- **Always ensure local build succeeds** before pushing

### Known Limitations
- No automated tests (manual validation required)
- Requires Azure CLI authentication
- TUI cannot be fully tested in headless environments
- Application may hang if Azure authentication expires

## Troubleshooting

### Common Issues
1. **Build fails with dependency errors**:
   ```bash
   make clean && make all  # Clean rebuild usually fixes
   ```

2. **Application fails to start**:
   ```bash
   az account show  # Verify Azure CLI authentication
   ```

3. **Authentication errors**:
   ```bash
   az login  # Re-authenticate with Azure
   ```

4. **Long build times**:
   - First build downloads dependencies (~36s)
   - Subsequent builds are fast (~0.5s)
   - **NEVER CANCEL** long-running builds

### Emergency Commands
```bash
# Force clean and rebuild
make clean && make all

# Reset Azure authentication
az logout && az login

# Check build artifacts
ls -la bin/aztui && file bin/aztui
```