pub use zenocore::*;

pub mod std {
    pub use zeno_std::*;
}

pub mod apidoc {
    pub use zeno_apidoc::*;
}

pub mod blade {
    pub use zeno_blade::*;
}

/// Convenience helper to create a fully-loaded engine with core and stdlib slots pre-registered.
pub fn new_engine() -> zenocore::executor::Engine {
    let mut engine = zenocore::executor::Engine::new();
    zenocore::slots::register_logic_slots(&mut engine);
    zeno_std::register_std_slots(&mut engine);
    zeno_blade::register_blade_slots(&mut engine);
    engine
}
