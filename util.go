package main

func extensionToHlAndAceLangs(extension string) (hlExt, aceExt string) {
	hlExt, exists := extensionToHl[extension]
	if !exists {
		hlExt = "text"
	}

	aceExt, exists = extensionToAce[extension]
	if !exists {
		aceExt = "text"
	}
	return
}

func supportedBinExtension(extension string) bool {
	_, exists := extensionToHl[extension]
	return exists
}

var extensionToAce = map[string]string{
	"c":      "c_cpp",
	"h":      "c_cpp",
	"cpp":    "c_cpp",
	"clj":    "clojure",
	"coffee": "coffee",
	"cfc":    "coldfusion",
	"cs":     "csharp",
	"sh":     "sh",
	"bash":   "sh",
	"css":    "css",
	"go":     "golang",
	"diff":   "diff",
	"html":   "html",
	"xml":    "xml",
	"ini":    "ini",
	"java":   "java",
	"js":     "javascript",
	"json":   "json",
	"jsp":    "jsp",
	"tex":    "latex",
	"lisp":   "lisp",
	"less":   "less",
	"lua":    "lua",
	"md":     "markdown",
	"ocaml":  "ocaml",
	"tcl":    "tcl",
	"yaml":   "yaml",
	"php":    "php",
	"pl":     "perl",
	"py":     "python",
	"rb":     "ruby",
	"sql":    "sql",
	"apache": "apache",
	"cmake":  "cmake",
	"bat":    "dos",
	"scala":  "scala",
	"txt":    "text",
}

var extensionToHl = map[string]string{
	"c":      "cpp",
	"h":      "cpp",
	"cpp":    "c_cpp",
	"clj":    "clojure",
	"coffee": "coffee",
	"cfc":    "coldfusion",
	"cs":     "csharp",
	"sh":     "sh",
	"bash":   "sh",
	"css":    "css",
	"go":     "go",
	"diff":   "diff",
	"html":   "html",
	"htm":    "html",
	"ini":    "ini",
	"java":   "java",
	"js":     "javascript",
	"json":   "json",
	"jsp":    "jsp",
	"tex":    "latex",
	"lisp":   "lisp",
	"less":   "less",
	"lua":    "lua",
	"ocaml":  "ocaml",
	"tcl":    "tcl",
	"nginx":  "nginx",
	"xml":    "xml",
	"yaml":   "yaml",
	"php":    "php",
	"pl":     "perl",
	"py":     "python",
	"rb":     "ruby",
	"sql":    "sql",
	"apache": "apache",
	"cmake":  "cmake",
	"bat":    "dos",
	"scala":  "scala",
	"txt":    "text",
}
