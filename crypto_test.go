package skynet

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_encodeNumber(t *testing.T) {
	tests := []struct {
		number int64
		want   []byte
	}{
		{
			number: 0,
			want:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			number: 1,
			want:   []byte{1, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			number: 2,
			want:   []byte{2, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			number: 255,
			want:   []byte{255, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			number: 256,
			want:   []byte{0, 1, 0, 0, 0, 0, 0, 0},
		},
		{
			number: 256 * 256,
			want:   []byte{0, 0, 1, 0, 0, 0, 0, 0},
		},
		{
			number: 256 * 256 * 256,
			want:   []byte{0, 0, 0, 1, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_to_byte_slice", tt.number), func(t *testing.T) {
			require.Equal(t, tt.want, encodeNumber(tt.number))
		})
	}
}

func Test_encodeString(t *testing.T) {
	tests := []struct {
		name     string
		toEncode string
		want     []byte
	}{
		{
			name:     "empty_string",
			toEncode: "",
			want:     []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "skynet_string_to_byte_slice",
			toEncode: "skynet",
			want:     []byte{6, 0, 0, 0, 0, 0, 0, 0, 115, 107, 121, 110, 101, 116},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, encodeString(tt.toEncode))
		})
	}
}

func Test_hashDataKey(t *testing.T) {
	tests := []struct {
		name   string
		toHash string
		want   string
	}{
		{
			name:   "empty_string_to_hash",
			toHash: "",
			want:   "81e47a19e6b29b0a65b9591762ce5143ed30d0261e5d24a3201752506b20f15c",
		},
		{
			name:   "skynet_to_hash",
			toHash: "skynet",
			want:   "31c7a4d53ef7bb4c7531181645a0037b9e75c8b1d1285b468ad58bad6262c777",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashDataKey(tt.toHash)
			require.Equal(t, tt.want, hash)
		})
	}
}
