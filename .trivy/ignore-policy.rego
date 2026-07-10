package trivy

import rego.v1

# Findings in these modules are ignored. They are indirect dependencies we cannot bump ourselves
indirect_only_modules := {
	"golang.org/x/crypto",
}

default ignore := false

ignore if input.PkgName in indirect_only_modules
