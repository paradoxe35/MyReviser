#!/usr/bin/env bash
#
# Build the Rust FFI static library for one target triple.
#
#   scripts/build-rust-ffi.sh [target-triple]
#
# With no triple it builds for the host. RUSTFLAGS and the rest of the
# environment are passed through to cargo untouched.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root/rust-ffi"

target="${1:-}"
cargo_args=(build --release)
if [ -n "$target" ]; then
  cargo_args+=(--target "$target")
  out_dir="target/$target/release"
else
  out_dir="target/release"
fi

# ggml's CMakeLists clears CMAKE_STATIC_LIBRARY_PREFIX on WIN32 ("remove the lib
# prefix on win32 mingw"), so transcribe-cpp-sys installs ggml.a / ggml-cpu.a /
# ggml-base.a while its link manifest still names them ggml, ggml-cpu and
# ggml-base. rustc targeting *-windows-gnu only ever looks for libggml.a, so a
# cold build dies with "could not find native static library `ggml`" with the
# archive sitting right there in the search path. Give each archive the name
# rustc searches for; libtranscribe.a is built outside ggml's scope and already
# has it.
normalize_native_archives() {
  shopt -s nullglob
  local archive name dir
  for archive in "$out_dir"/build/transcribe-cpp-sys-*/out/lib/*.a; do
    name="$(basename "$archive")"
    dir="$(dirname "$archive")"
    case "$name" in
      lib*) continue ;;
    esac
    cp -f "$archive" "$dir/lib$name"
    echo "  aliased $name -> lib$name"
  done
}

if cargo "${cargo_args[@]}"; then
  exit 0
fi

# The native build itself succeeded above (cmake installed the archives); only
# the Rust link line could not resolve them. Rename and re-run: the second pass
# reuses the cached build-script output, so nothing native is rebuilt.
echo "Rust build failed; normalizing native archive names and retrying..."
normalize_native_archives
cargo "${cargo_args[@]}"
