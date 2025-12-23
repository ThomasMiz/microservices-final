#[cfg(test)]
mod integration_tests {
    #[test]
    fn test_base_path_environment_variable() {
        // Test that we can read the BASE_PATH environment variable
        unsafe {
            std::env::set_var("BASE_PATH", "/api/billing");
        }

        let base_path = std::env::var("BASE_PATH").unwrap_or_else(|_| "/api/billing".to_string());
        assert_eq!(base_path, "/api/billing");

        // Test default value
        unsafe {
            std::env::remove_var("BASE_PATH");
        }
        let default_path =
            std::env::var("BASE_PATH").unwrap_or_else(|_| "/api/billing".to_string());
        assert_eq!(default_path, "/api/billing");
    }
}
