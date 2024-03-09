type file struct {
	name  string
	fsize int
}

mkdir! "testgsh"

mkdir "testgsh2"
lastErr!

mkdir "testgsh3"
if lastErr != nil {
	panic lastErr
}

capout => { ls }
println output.fields

capout => { ls "-l" }
files := [file{flds[8], flds[4].int!} for e <- output.split("\n") if flds := e.fields; flds.len > 2]
println files

rmdir "testgsh", "testgsh2", "testgsh3"
