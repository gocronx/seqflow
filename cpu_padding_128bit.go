//go:build (arm64 && darwin) || ppc64 || ppc64le

package seqflow

// CacheLineBytes 是当前架构的 CPU 缓存行大小（字节）
const CacheLineBytes = 128
