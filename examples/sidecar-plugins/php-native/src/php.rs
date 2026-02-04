use libc::{c_char, c_int, c_void};

// PHP Embed SAPI Bindings
// This mirrors the C API needed to run PHP.

#[repr(C)]
pub struct zend_file_handle {
    // Simplified opaque structure sufficient for binding
    // In real usage, you might need the full struct definition if you pass it by value
    // but here we mostly use pointers.
    pub handle: *mut c_void,
}

extern "C" {
    // Initialises the PHP embed SAPI
    // int php_embed_init(int argc, char **argv);
    pub fn php_embed_init(argc: c_int, argv: *mut *mut c_char) -> c_int;

    // Shuts down the PHP embed SAPI
    // void php_embed_shutdown(void);
    pub fn php_embed_shutdown();

    // Evaluates a string of PHP code
    // zend_result zend_eval_string(char *str, zval *retval_ptr, char *string_name);
    // retval_ptr can be null if we don't care about the return value (output capture handled via buffer)
    pub fn zend_eval_string(
        str: *const c_char,
        retval_ptr: *mut c_void, // zval pointer
        string_name: *const c_char,
    ) -> c_int;

    // Request lifecycle
    pub fn php_request_startup() -> c_int;
    pub fn php_request_shutdown(dummy: *mut c_void);
}

// Helpers for Rust interaction
pub fn init() -> bool {
    unsafe {
        // argc=0, argv=null
        let res = php_embed_init(0, std::ptr::null_mut());
        res == 0 // SUCCESS
    }
}

pub fn request_startup() -> bool {
    unsafe {
        let res = php_request_startup();
        res == 0 // SUCCESS
    }
}

pub fn request_shutdown() {
    unsafe {
        php_request_shutdown(std::ptr::null_mut());
    }
}

pub fn shutdown() {
    unsafe {
        php_embed_shutdown();
    }
}

pub fn eval(code: &str) -> bool {
    let c_code = std::ffi::CString::new(code).unwrap();
    let c_name = std::ffi::CString::new("zeno-rust-bridge").unwrap();

    unsafe {
        // retval_ptr = null (we rely on output buffering or side effects)
        let res = zend_eval_string(c_code.as_ptr(), std::ptr::null_mut(), c_name.as_ptr());
        res == 0 // SUCCESS
    }
}
