package main

func extensionToHlLang(extension string) (hlExt string) {
	hlExt, exists := extensionToHl[extension]
	if !exists {
		hlExt = "text"
	}
	return
}

func supportedBinExtension(extension string) bool {
	_, exists := extensionToHl[extension]
	return exists
}

var extensionToHl = map[string]string{
	"ahk":         "autohotkey",
	"apache":      "apache",
	"applescript": "applescript",
	"bas":         "basic",
	"bash":        "sh",
	"bat":         "dos",
	"c":           "cpp",
	"cfc":         "coldfusion",
	"clj":         "clojure",
	"cmake":       "cmake",
	"coffee":      "coffee",
	"cpp":         "c_cpp",
	"cs":          "csharp",
	"css":         "css",
	"d":           "d",
	"dart":        "dart",
	"diff":        "diff",
	"dockerfile":  "dockerfile",
	"elm":         "elm",
	"erl":         "erlang",
	"for":         "fortran",
	"go":          "go",
	"h":           "cpp",
	"htm":         "html",
	"html":        "html",
	"ini":         "ini",
	"java":        "java",
	"js":          "javascript",
	"json":        "json",
	"jsp":         "jsp",
	"kt":          "kotlin",
	"less":        "less",
	"lisp":        "lisp",
	"lua":         "lua",
	"m":           "objectivec",
	"nginx":       "nginx",
	"ocaml":       "ocaml",
	"php":         "php",
	"pl":          "perl",
	"proto":       "protobuf",
	"ps":          "powershell",
	"py":          "python",
	"rb":          "ruby",
	"rs":          "rust",
	"scala":       "scala",
	"scm":         "scheme",
	"scpt":        "applescript",
	"scss":        "scss",
	"sh":          "sh",
	"sql":         "sql",
	"tcl":         "tcl",
	"tex":         "latex",
	"toml":        "ini",
	"ts":          "typescript",
	"txt":         "text",
	"xml":         "xml",
	"yaml":        "yaml",
	"yml":         "yaml",
}
