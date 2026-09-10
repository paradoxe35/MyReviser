pub mod core;
pub mod ffi;
pub mod stt;

pub use ffi::*;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_library_smoke() {
        assert!(true);
    }
}
