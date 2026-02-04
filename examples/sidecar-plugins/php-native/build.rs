fn main() {
    // --- STATIC LINKING CONFIGURATION ---
    // This script runs before compilation to tell cargo how to link libphp.

    #[cfg(target_os = "windows")]
    {
        // Adjust these paths to point to your PHP dev pack (php-sdk)
        // Example: C:\php-sdk\php-8.2-devel-vs16-x64\lib
        println!("cargo:rustc-link-search=native=C:/php/dev/lib");
        println!("cargo:rustc-link-lib=static=php8"); // links php8.lib
    }

    #[cfg(not(target_os = "windows"))]
    {
        // Linux/Mac configuration
        // Often managed via php-config
        // println!("cargo:rustc-link-lib=php");
    }

    // Rerun if build script changes
    println!("cargo:rerun-if-changed=build.rs");
}
