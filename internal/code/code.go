package code

import (
	"embed"
	"strings"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/api"
)

// CodeSpaceAccessLevel represents the code space access level type.
type CodeSpaceAccessLevel int

const (
	// CodeSpaceAccessLevelReadOnly represents read only code space access level.
	CodeSpaceAccessLevelReadOnly CodeSpaceAccessLevel = 1
	// CodeSpaceAccessLevelReadWrite represents read-write code space access level.
	CodeSpaceAccessLevelReadWrite CodeSpaceAccessLevel = 2
)

var (
	//go:embed _codetemplates/*/*
	CodeTemplatesFS embed.FS
	//go:embed data/adjectives.txt
	AdjectivesFile string
	Adjectives     []string
	//go:embed data/file_extensions.txt
	FileExtensionsFile string
	FileExtensions     []string
	//go:embed data/harry_potter_characters.txt
	HarryPotterCharactersFile string
	HarryPotterCharacters     []string
	//go:embed data/pokemon.txt
	PokemonFile string
	Pokemon     []string
	//go:embed data/programming_terms.txt
	ProgrammingTermsFile string
	ProgrammingTerms     []string
)

// CodingLanguageConfig includes configuration information for coding languages.
var CodingLanguageConfig = map[string]struct {
	fileName string
	version  string
}{
	// C
	api.PistonLanguageC: {
		fileName: "main.c",
		version:  api.PistonVersionC,
	},
	// C++
	api.PistonLanguageCPlusPlus: {
		fileName: "main.cpp",
		version:  api.PistonVersionCPlusPlus,
	},
	// Go
	api.PistonLanguageGo: {
		fileName: "main.go",
		version:  api.PistonVersionGo,
	},
	// Java
	api.PistonLanguageJava: {
		fileName: "Main.java",
		version:  api.PistonVersionJava,
	},
	// JavaScript
	api.PistonLanguageJavaScript: {
		fileName: "index.js",
		version:  api.PistonVersionJavaScript,
	},
	// Python
	api.PistonLanguagePython: {
		fileName: "main.py",
		version:  api.PistonVersionPython,
	},
	// Rust
	api.PistonLanguageRust: {
		fileName: "main.rs",
		version:  api.PistonVersionRust,
	},
	// TypeScript
	api.PistonLanguageTypeScript: {
		fileName: "index.ts",
		version:  api.PistonVersionTypeScript,
	},
}

// CodeSpace represents the database table "code_space".
type CodeSpace struct {
	ID         int64     `db:"id"`
	AuthorUUID *string   `db:"author_uuid"`
	Name       string    `db:"name"`
	Language   string    `db:"language"`
	Contents   string    `db:"contents"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// CodeSpaceAccess represents the database table "code_space_access".
type CodeSpaceAccess struct {
	ID          int64                `db:"id"`
	UserUUID    string               `db:"user_uuid"`
	CodeSpaceID int64                `db:"code_space_id"`
	Level       CodeSpaceAccessLevel `db:"level"`
	CreatedAt   time.Time            `db:"created_at"`
	UpdatedAt   time.Time            `db:"updated_at"`
}

//nolint:gochecknoinits
func init() {
	Adjectives = strings.Split(strings.TrimSpace(AdjectivesFile), "\n")
	FileExtensions = strings.Split(strings.TrimSpace(FileExtensionsFile), "\n")
	HarryPotterCharacters = strings.Split(strings.TrimSpace(HarryPotterCharactersFile), "\n")
	Pokemon = strings.Split(strings.TrimSpace(PokemonFile), "\n")
	ProgrammingTerms = strings.Split(strings.TrimSpace(ProgrammingTermsFile), "\n")
}

// String returns the API string representation of an access level.
func (l CodeSpaceAccessLevel) String() string {
	switch l {
	case CodeSpaceAccessLevelReadOnly:
		return api.CodeSpaceAccessLevelReadOnly
	case CodeSpaceAccessLevelReadWrite:
		return api.CodeSpaceAccessLevelReadWrite
	default:
		return ""
	}
}

// GetAccessLevelFromString gets the access level from the API string representation.
func GetAccessLevelFromString(accessLevel string) CodeSpaceAccessLevel {
	switch accessLevel {
	case api.CodeSpaceAccessLevelReadOnly:
		return CodeSpaceAccessLevelReadOnly
	case api.CodeSpaceAccessLevelReadWrite:
		return CodeSpaceAccessLevelReadWrite
	default:
		return 0
	}
}
