package network

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"math/rand"
	"time"
)

// Util

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func compressData(data []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	w := zlib.NewWriter(b)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func uncompressData(data []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	if _, err := b.Write(data); err != nil {
		return nil, err
	}
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}
