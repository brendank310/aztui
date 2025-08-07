This is a Go based repository that utilizes the Azure SDK as a client. It is primarily responsible for presenting users with a text based user interface to interact with their Azure resources. Please follow these guidelines when contributing:

## Code Standards

### Required Before Each Commit
- Run `make fmt` before committing any changes to ensure proper code formatting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow
- Build: `make build`
- Test: `make test`

## Repository Structure
- `src/cmd/`: Main service entry points and executables
- `src/pkg/azcli`: Code to allow user's to run azcli commands on selected resources
- `src/pkg/consoles`: Code to allow user's to interact with Azure resources with various console access options (serial console, cloud shell, etc)
- `src/pkg/resourceviews`: Representations of the user's view of available Azure resources
- `conf/`: User configuration files (default keybindings, vim keybindings, etc)

## Key Guidelines
1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Use dependency injection patterns where appropriate
4. Write unit tests for new functionality. Use table-driven unit tests when possible.
5. Document public APIs and complex logic. Suggest changes to the `docs/` folder when appropriate
