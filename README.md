
# HexDox

is a lightweight and efficient tool to identify potential dependency confusion risks by scanning `package.json` files hosted at given URLs. The tool works with piped input, processes URLs concurrently, and supports optional verbose logging for detailed outputs.

## Features

- **Concurrent URL Processing**: Specify the number of threads using the `-c` flag for faster scanning.
- **Dependency Validation**: Checks if dependencies listed in `package.json` exist on npm.
- **Potential Dependency Confusion Alerts**: Flags dependencies not found on npm.
- **Piped Input**: Accepts a list of URLs via `stdin` for streamlined workflows.
- **Verbose Logging**: Enable detailed logs using the `-v` flag.

## Requirements

- Go 1.22 or higher
- Internet connection

## Installation

`go install -v github.com/Vulnpire/HexDox@latest`

## Usage

### Input Format

The tool accepts a list of URLs (one per line) via `stdin`. Each URL should point to a `package.json` file.

### Command-line Options

| Flag     | Description                                     | Default |
|----------|-------------------------------------------------|---------|
| `-c`     | Number of concurrent threads for processing     | 5       |
| `-v`     | Enable verbose output for informational logs    | Disabled |

### Example

1. Create a file `urls.txt` with a list of URLs:
   ```
   https://example.com/path/to/package.json
   https://another-example.com/package.json
   ```

2. Run the tool with piped input:
   ```
   cat urls.txt | ./HexDox -c=10
   ```

   This will process the URLs with a concurrency level of 10.

3. Enable verbose output for detailed logging:
   ```
   cat urls.txt | ./HexDox -c=10 -v
   ```

### Example Output

#### Without `-v`
```
[WARNING] Potential Dependency Confusion: 'my-fake-dependency' not found on npm
```

#### With `-v`
```
[INFO] Dependency 'express' exists on npm
[WARNING] Potential Dependency Confusion: 'my-fake-dependency' not found on npm
[ERROR] Failed to fetch 'https://invalid-url.com': no such host
```

## How It Works

1. **URL Fetching**: The tool fetches the `package.json` file from each URL.
2. **Dependency Parsing**: It parses `dependencies` and `devDependencies` from the JSON file.
3. **Validation**: Each dependency is checked against npm using the `api.allorigins.win` proxy to avoid bot detection.
4. **Output**:
   - `[WARNING]`: Dependencies not found on npm are flagged for potential dependency confusion.
   - `[INFO]`: Valid dependencies are logged when `-v` is enabled.
   - `[ERROR]`: Fetching or parsing issues are logged when `-v` is enabled.

## Contributing

Contributions are welcome! Please fork the repository, create a feature branch, and submit a pull request with your changes.

### To Do

- Add support for npm scopes (e.g., `@org/package`).
- Improve error handling for edge cases like malformed JSON or unreachable URLs.

## Disclaimer

This tool is for educational and bug bounty purposes only. Ensure you have proper permissions before testing third-party systems.
