package main  
import (  
    "fmt"
    "log"
    "os"
    "os/exec"
)

// CrossCompileTarget represents a target operating system and architecture for cross-compilation.
type CrossCompileTarget struct {
    OS   string
    Arch string
}

// goOsAndArch checks if GOOS and GOARCH are set to valid values.
func goOsAndArch() bool {
    if os.Getenv("GOOS") == "" || os.Getenv("GOARCH") == "" {
        fmt.Println("Please set GOOS and GOARCH environment variables.")
        return false
    }
    return true
}

// crossCompile performs the cross-compilation for a given target.
func crossCompile(target CrossCompileTarget) {
    fmt.Printf("Cross-compiling for %s/%s...\n", target.OS, target.Arch)
    cmd := exec.Command("go", "build", "-o", fmt.Sprintf("binary-%s-%s", target.OS, target.Arch), "main.go")
    cmd.Env = append(os.Environ(), "GOOS="+target.OS, "GOARCH="+target.Arch)
    if err := cmd.Run(); err != nil {
        log.Fatalf("Cross-compilation failed: %v", err)
    }
    fmt.Println("Cross-compilation successful!")
}

func main() {
    if !goOsAndArch() {
        return
    }
    // Define the targets for cross-compilation.
    targets := []CrossCompileTarget{
        {OS: "windows", Arch: "amd64"},
        {OS: "linux", Arch: "amd64"},
        {OS: "darwin", Arch: "amd64"},
        {OS: "linux", Arch: "arm64"},
    }

    // Run cross-compilation for each target.
    for _, target := range targets {
        crossCompile(target)
    }
    fmt.Println("All cross-compilations successful!")
} 