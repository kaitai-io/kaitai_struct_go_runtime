rm -rf src
mkdir src

git clone https://github.com/kaitai-io/kaitai_struct_tests
mv kaitai_struct_tests/spec/go src/spec
mv kaitai_struct_tests/src/* src/
rm -rf kaitai_struct_tests
git clone https://github.com/kaitai-io/ci_targets/
mv ci_targets/compiled/go/src/test_formats src/test_formats
rm -rf ci_targets

mkdir -p $HOME/gopath/src/github.com/github.com/kaitai-io/kaitai_struct_go_runtime/kaitai
cp -r $HOME/gopath/src/github.com/cugu/kaitai_struct_go_runtime $HOME/gopath/src/github.com/github.com/kaitai-io/kaitai_struct_go_runtime/kaitai

go get github.com/stretchr/testify/assert
go get golang.org/x/text
go get github.com/jstemmer/go-junit-report
go get -u gopkg.in/alecthomas/gometalinter.v2
