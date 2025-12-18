package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pythonTemplate = `import gorunpy


@gorunpy.export
def add(a: int, b: int) -> int:
    """Add two numbers."""
    return a + b


@gorunpy.export
def multiply(a: float, b: float) -> float:
    """Multiply two numbers."""
    return a * b
`

const mainGoTemplate = `//go:generate .gorunpy/venv/bin/gorunpy gen

package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewPylibClient()

	sum, err := client.Add(ctx, 2, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2 + 3 = %d\n", sum)

	product, err := client.Multiply(ctx, 2.5, 4.0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2.5 * 4.0 = %.1f\n", product)
}
`

const gitignoreEntries = `# gorunpy
.gorunpy/
gorunpy_client.go
__pycache__/
*.pyc
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := initProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: gorunpy <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init    Initialize a new gorunpy project")
	fmt.Println("  help    Show this help message")
}

// findPython returns the path to python executable
func findPython() (string, error) {
	// Try python first
	if path, err := exec.LookPath("python"); err == nil {
		// Verify it's Python 3
		cmd := exec.Command(path, "--version")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "Python 3") {
			return path, nil
		}
	}

	// Try python3
	if path, err := exec.LookPath("python3"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("Python 3 not found. Please install Python 3.9+ and try again")
}

func initProject() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check for Python first
	pythonPath, err := findPython()
	if err != nil {
		return err
	}
	fmt.Printf("Found Python: %s\n", pythonPath)

	// Check if go.mod exists
	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Run 'go mod init <module>' first")
	}

	fmt.Println("Initializing gorunpy project...")
	fmt.Println("")

	// 1. Create .gorunpy directory
	gorunpyDir := filepath.Join(cwd, ".gorunpy")
	if err := os.MkdirAll(gorunpyDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gorunpy: %w", err)
	}

	// 2. Create virtual environment
	venvDir := filepath.Join(gorunpyDir, "venv")
	if _, err := os.Stat(venvDir); os.IsNotExist(err) {
		fmt.Println("[1/5] Creating virtual environment...")
		cmd := exec.Command(pythonPath, "-m", "venv", venvDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create venv: %w", err)
		}
	} else {
		fmt.Println("[1/5] Virtual environment already exists")
	}

	// 3. Install gorunpy[build]
	fmt.Println("[2/5] Installing gorunpy[build]...")
	pipPath := filepath.Join(venvDir, "bin", "pip")
	cmd := exec.Command(pipPath, "install", "-q", "gorunpy[build]")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install gorunpy: %w", err)
	}

	// 4. Create Python package
	pylibDir := filepath.Join(cwd, "pylib")
	if _, err := os.Stat(pylibDir); os.IsNotExist(err) {
		fmt.Println("[3/5] Creating pylib/ package...")
		if err := os.MkdirAll(pylibDir, 0755); err != nil {
			return fmt.Errorf("failed to create pylib: %w", err)
		}

		// Create __init__.py
		initFile := filepath.Join(pylibDir, "__init__.py")
		if err := os.WriteFile(initFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create __init__.py: %w", err)
		}

		// Create calc.py
		calcFile := filepath.Join(pylibDir, "calc.py")
		if err := os.WriteFile(calcFile, []byte(pythonTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create calc.py: %w", err)
		}
	} else {
		fmt.Println("[3/5] pylib/ already exists, skipping")
	}

	// 5. Create main.go if it doesn't exist
	mainGoPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		fmt.Println("[4/5] Creating main.go...")
		if err := os.WriteFile(mainGoPath, []byte(mainGoTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create main.go: %w", err)
		}
	} else {
		fmt.Println("[4/5] main.go already exists, skipping")
	}

	// 6. Update .gitignore
	fmt.Println("[5/5] Updating .gitignore...")
	if err := appendGitignore(filepath.Join(cwd, ".gitignore")); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	// 7. Add Go dependency (silent)
	exec.Command("go", "get", "github.com/younseoryu/gorunpy/gorunpy").Run()

	fmt.Println("")
	fmt.Println("✓ Project initialized!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("")
	fmt.Println("  1. Run the project:")
	fmt.Println("     go generate && go run .")
	fmt.Println("")
	fmt.Println("  2. Edit pylib/calc.py to add your own functions")
	fmt.Println("")
	fmt.Println("  3. Regenerate after changes:")
	fmt.Println("     go generate")
	fmt.Println("")

	return nil
}

func appendGitignore(path string) error {
	// Read existing content
	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	// Check what needs to be added
	lines := strings.Split(strings.TrimSpace(gitignoreEntries), "\n")
	var toAdd []string
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(existing, line) {
			toAdd = append(toAdd, line)
		}
	}

	if len(toAdd) == 0 {
		return nil // Nothing to add
	}

	// Open file for appending (create if not exists)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		f.WriteString("\n")
	}

	// Add header and entries
	f.WriteString("\n# gorunpy\n")
	for _, entry := range toAdd {
		f.WriteString(entry + "\n")
	}

	return nil
}
