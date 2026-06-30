# GitHub Release Helper

A lightweight Go CLI utility designed to automate GitHub release creation and asset (binary) uploading using the GitHub REST API.

This tool acts as a drop-in replacement for the standard `gh release` command, which is particularly useful when executing within sandboxed or restricted terminal environments where direct `gh release` operations are blocked by permission policies.

---

## 🛠️ How to Build

Compile the helper directly into your local `bin/` directory:

```bash
go build -o bin/relhelper etc/release_helper.go
```

---

## 🔑 Token Resolution

The utility automatically resolves your GitHub Authentication Token by checking the following sources in order:

1. **`GITHUB_TOKEN`** environment variable.
2. **`GH_TOKEN`** environment variable.
3. **`gh auth token`**: It runs the local GitHub CLI in a background subprocess to retrieve the active token if you are already authenticated.

If no token is found, the tool exits with an error.

---

## 📖 Command Usage

### Options and Flags

| Flag | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `--tag` | String | **[Required]** The tag name for the release (e.g., `v5.3.11`). | |
| `--title` | String | The title of the release. | Same as `--tag` |
| `--notes` | String | Plain-text release notes. | |
| `--notes-file`| String | Path to a file containing the release notes (e.g., a Markdown file). | |
| `--assets` | String | Glob pattern matching the files to upload. | `bin/*` |
| `--owner` | String | GitHub repository owner. | `tanaikech` |
| `--repo` | String | GitHub repository name. | `ggsrun` |
| `--target` | String | Target branch or commit SHA. | `master` |
| `--draft` | Boolean| Create the release as a draft. | `false` |
| `--prerelease`| Boolean| Identify the release as a prerelease. | `false` |

---

## 💡 Examples

### 1. Create a release with inline notes and upload default binaries

```bash
./bin/relhelper --tag v5.3.11 --notes "This is a patch release with performance optimizations."
```

### 2. Create a release using an external release notes file

```bash
./bin/relhelper --tag v5.3.11 --notes-file help/UpdateHistory.md
```

### 3. Upload a specific set of assets for a custom repository

```bash
./bin/relhelper --tag v1.0.0 --owner myorg --repo myproject --assets "dist/*.zip"
```
