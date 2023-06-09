// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	_ "embed"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

//go:embed zstd_dict
var zStdDict []byte

// Create a writer that caches compressors.
// For this operation type we supply a nil Reader.
var zStdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderDict(zStdDict))

// Compress a buffer.
// If you have a destination buffer, the allocation in the call can also be eliminated.
func ZStdCompress(src []byte) []byte {
	return zStdEncoder.EncodeAll(src, make([]byte, 0, len(src)))
}

// Create a reader that caches decompressors.
// For this operation type we supply a nil Reader.
var zStdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))

// Decompress a buffer. We don't supply a destination buffer,
// so it will be allocated by the decoder.
func ZStdDecompress(src []byte) ([]byte, error) {
	return zStdDecoder.DecodeAll(src, nil)
}

func BrotliCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, brotli.BestCompression)
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func BrotliDecompress(data []byte) ([]byte, error) {
	var b []byte
	r := brotli.NewReader(bytes.NewBuffer(data))
	_, err := r.Read(b)
	return b, err
}
