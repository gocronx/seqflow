//go:build arm || mips || mipsle || mips64 || mips64le

package seqflow

// CacheLineBytes 是当前架构的 CPU 缓存行大小（字节）
const CacheLineBytes = 32
