package api

const (
	// PistonFileEncoding represents the content encoding used for files on Piston requests.
	PistonFileEncoding = "utf8"

	// PistonLanguageC represents the C language runtime in Piston.
	PistonLanguageC = "c"
	// PistonVersionC represents the version of C language runtime in Piston.
	PistonVersionC = "10.2.0"

	// PistonLanguageCPlusPlus represents the C++ language runtime in Piston.
	PistonLanguageCPlusPlus = "c++"
	// PistonVersionCPlusPlus represents the version of C++ language runtime in Piston.
	PistonVersionCPlusPlus = "10.2.0"

	// PistonLanguageGo represents the Go language runtime in Piston.
	PistonLanguageGo = "go"
	// PistonVersionGo represents the version of Go language runtime in Piston.
	PistonVersionGo = "1.16.2"

	// PistonLanguageJava represents the Java language runtime in Piston.
	PistonLanguageJava = "java"
	// PistonVersionJava represents the version of Java language runtime in Piston.
	PistonVersionJava = "15.0.2"

	// PistonLanguageJavaScript represents the JavaScript language runtime in Piston.
	PistonLanguageJavaScript = "javascript"
	// PistonVersionJavaScript represents the version of JavaScript language runtime in Piston.
	PistonVersionJavaScript = "18.15.0"

	// PistonLanguagePython represents the Python language runtime in Piston.
	PistonLanguagePython = "python"
	// PistonVersionPython represents the version of Python language runtime in Piston.
	PistonVersionPython = "3.10.0"

	// PistonLanguageRust represents the Rust language runtime in Piston.
	PistonLanguageRust = "rust"
	// PistonVersionRust represents the version of Rust language runtime in Piston.
	PistonVersionRust = "1.68.2"

	// PistonLanguageTypeScript represents the TypeScript language runtime in Piston.
	PistonLanguageTypeScript = "typescript"
	// PistonVersionTypeScript represents the version of TypeScript language runtime in Piston.
	PistonVersionTypeScript = "1.32.3"
)

// PistonFile represents a file sent for code execution to Piston.
type PistonFile struct {
	Name     *string `json:"name"`
	Content  string  `json:"content"`
	Encoding *string `json:"encoding"`
}

// PistonExecuteRequest represents the request body for Piston code execution requests.
type PistonExecuteRequest struct {
	Language           string       `json:"language"`
	Version            string       `json:"version"`
	Files              []PistonFile `json:"files"`
	Stdin              *string      `json:"stdin"`
	Args               []string     `json:"args"`
	CompileTimeout     *int64       `json:"compile_timeout"`
	RunTimeout         *int64       `json:"run_timeout"`
	CompileMemoryLimit *int64       `json:"compile_memory_limit"`
	RunMemoryLimit     *int64       `json:"run_memory_limit"`
}

// PistonResults represents code execution results from Piston.
type PistonResults struct {
	Stdout string  `json:"stdout"`
	Stderr string  `json:"stderr"`
	Output string  `json:"output"`
	Code   *int    `json:"code"`
	Signal *string `json:"signal"`
}

// PistonExecuteResponse represents the response body for Piston code execution requests.
type PistonExecuteResponse struct {
	Language string         `json:"language"`
	Version  string         `json:"version"`
	Compile  *PistonResults `json:"compile"`
	Run      PistonResults  `json:"run"`
}
