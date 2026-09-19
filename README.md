# Kaitai Struct: runtime library for Go

[![Pipeline](https://github.com/kaitai-io/kaitai_struct_go_runtime/workflows/Go/badge.svg)](https://github.com/kaitai-io/kaitai_struct_go_runtime/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaitai-io/kaitai_struct_go_runtime/kaitai.svg)](https://pkg.go.dev/github.com/kaitai-io/kaitai_struct_go_runtime/kaitai)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaitai-io/kaitai_struct_go_runtime)](https://goreportcard.com/report/github.com/kaitai-io/kaitai_struct_go_runtime)

This library implements Kaitai Struct API for Go.

Kaitai Struct is a declarative language used for describe various binary
data structures, laid out in files or in memory: i.e. binary file
formats, network stream packet formats, etc.

Further reading:

* [About Kaitai Struct](https://kaitai.io/)
* [About API implemented in this library](https://doc.kaitai.io/stream_api.html)

## Experimental serialization

The runtime includes experimental serialization support through `Writer`,
seekable buffers, and child writers. Go compiler integration and its conformance
tests are still in development; this does not yet imply released compiler support.

Child writers accept a `ByteEncoder` with `Encode([]byte) ([]byte, error)` to
transform logical bytes into their on-disk representation. Custom processor
objects can implement this method directly. `ByteEncoderFunc` adapts functions
and closures, for example `kaitai.ByteEncoderFunc(kaitai.UnprocessZlib)`. Encoders
may modify the input copy supplied by a child writer and should be deterministic
because failed write-back can be retried. Their output must match the reserved
outer size, including any schema-specific padding or terminators.

Built-in inverse processing uses `ProcessXOR` again for XOR,
`ProcessRotateRight` for a left rotation (and vice versa), and `UnprocessZlib`
for RFC 1950 zlib compression. These built-in functions leave the source bytes
unchanged and return independent output. Encoder errors propagate through
write-back with wrapping that supports Go's `errors.Is` and `errors.As`.

## Licensing

Copyright 2017-2025 Kaitai Project: MIT license

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
