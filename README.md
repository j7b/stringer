# Stringer
[![GoDoc](https://godoc.org/github.com/j7b/stringer?status.svg)](http://godoc.org/github.com/j7b/stringer)
Stringer is a simpler version of the "official" [stringer](https://godoc.org/golang.org/x/tools/cmd/stringer) tool with different runtime options.

# Usage
```
  -dir directory
    	Package directory, defaults to current.
  -o filename
    	Output filename, defaults to stringer_gen.go in package directory.  If a literal dash (meaning '-o -') writes to standard output.
  -types list
    	Comma-delimited list of names of types to generate String methods for, defaults to all public types with named constants.
```

# Notes
The generated String() method contains the constant tests in the order they're found in source files. If two different named constants have the same value, the first matching name will be returned by the String() method.