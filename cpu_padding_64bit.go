//go:build 386 || amd64 || (arm64 && !darwin) || loong64 || riscv64 || wasm

package seqflow

// CacheLineBytes 是当前架构的 CPU 缓存行大小（字节）
const CacheLineBytes = 64
